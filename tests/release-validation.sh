#!/usr/bin/env bash
# Run only in the manually dispatched release-validation workflow.
set -euo pipefail
: "${GITHUB_REPOSITORY:?}" "${GITHUB_SHA:?}" "${GITHUB_REF:?}" "${RUNNER_TEMP:?}" "${GITHUB_STEP_SUMMARY:?}"
[[ "$GITHUB_REF" == refs/heads/main ]]
[[ "$GITHUB_SHA" =~ ^[a-f0-9]{40}$ ]]
workflows=(control-api.yml event-fabric.yml automation.yml kubernetes.yml gitops.yml supply-chain.yml mtls.yml kafka-ha.yml postgres-ha.yml)
check_head(){
  test "$(gh api "repos/$GITHUB_REPOSITORY/git/ref/heads/main" --jq .object.sha)" = "$GITHUB_SHA"
}
check_head
started=$(date -u +%Y-%m-%dT%H:%M:%SZ)
for workflow in "${workflows[@]}"; do
  check_head
  gh workflow run "$workflow" --repo "$GITHUB_REPOSITORY" --ref main
done
report="$RUNNER_TEMP/aether-release-validation.json"
# Poll runs that were manually dispatched after this validation began, at exactly this SHA.
# A skipped, failed, cancelled, missing or timed-out workflow cannot qualify.
for ((i=0;i<150;i++)); do
  check_head
  rows='[]'
  for workflow in "${workflows[@]}"; do
    runs=$(gh api "repos/$GITHUB_REPOSITORY/actions/workflows/$workflow/runs?branch=main&event=workflow_dispatch&head_sha=$GITHUB_SHA&per_page=30")
    row=$(jq -c --arg start "$started" --arg wf "$workflow" --arg sha "$GITHUB_SHA" '
      [.workflow_runs[] | select(.created_at >= $start and .head_sha==$sha)] | sort_by(.created_at,.id) | last |
      if .==null then {workflow:$wf,status:"missing"} else {workflow:$wf,id,status,conclusion,html_url} end' <<< "$runs")
    rows=$(jq -c --argjson row "$row" '. + [$row]' <<< "$rows")
  done
  jq -n --arg revision "$GITHUB_SHA" --argjson runs "$rows" '{revision:$revision,runs:$runs}' > "$report"
  if jq -e 'all(.[]; .status=="completed")' <<< "$rows" >/dev/null; then
    break
  fi
  sleep 15
done
{
  echo "## Release candidate $GITHUB_SHA"
  echo
  echo '| Workflow | Status | Conclusion | Evidence |'
  echo '| --- | --- | --- | --- |'
  jq -r '.runs[] | "| \(.workflow) | \(.status) | \(.conclusion // "pending") | \(if .html_url then "[run](\(.html_url))" else "missing" end) |"' "$report"
} >> "$GITHUB_STEP_SUMMARY"
check_head
jq -e --argjson count "${#workflows[@]}" '(.runs|length)==$count and all(.runs[]; .status=="completed" and .conclusion=="success")' "$report"
echo "All required workflows passed at the same main revision; eligible for local release review."
