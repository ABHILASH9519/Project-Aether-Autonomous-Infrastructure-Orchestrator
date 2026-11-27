#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_IMAGE_DIGEST:?expected registry digest}"
[[ "$AETHER_IMAGE_DIGEST" =~ ^sha256:[a-f0-9]{64}$ ]]
expected="localhost:5001/aether/event-fabric@$AETHER_IMAGE_DIGEST"
for app in ingest consumer relay workflow executor fixture; do
  kubectl -n aether rollout status deployment/"$app" --timeout=180s
  kubectl -n aether get deployment "$app" -o json | jq -e --arg image "$expected" '.spec.template.spec.containers[0].image == $image and .spec.template.spec.containers[0].imagePullPolicy == "Always"'
  kubectl -n aether get pods -l app="$app" -o json | jq -e --arg image "$expected" '(.items|length)>0 and all(.items[]; .spec.containers[0].image == $image and .status.containerStatuses[0].ready == true and (.status.containerStatuses[0].imageID|length)>0)'
done
echo "All six workloads are ready using the exact registry manifest digest"
