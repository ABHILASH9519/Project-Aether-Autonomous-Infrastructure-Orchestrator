#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?required}"
: "${AETHER_APPROVER_TOKEN:?required}"
compose=(docker compose -p aether-ha -f deploy/compose/events.yaml -f deploy/compose/kafka-ha.yaml --profile workflow)
# Use a different live broker for administration after stopping the leader.
admin=kafka-2
describe(){
 "${compose[@]}" exec -T "$admin" /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:29092 --describe --topic aether.events.v1
}
replicas_ready(){
 describe | awk '/Partition: [0-9]+/ {for(i=1;i<=NF;i++){if($i=="Replicas:"){n=split($(i+1),r,",");if(n!=3)bad=1}if($i=="Isr:"){n=split($(i+1),r,",");if(n!='"$1"')bad=1}};count++} END{exit !(count==3&&!bad)}'
}
for i in {1..60}; do
 if replicas_ready 3; then break; fi
 sleep 2
done
replicas_ready 3
export AETHER_EXECUTION_ID="ha-before-$(date +%s)-$$"
AETHER_INGEST_ATTEMPTS=10 bash tests/execution-smoke.sh
before="$AETHER_EXECUTION_ID"
leader=$(describe | awk '/Partition: 0([[:space:]]|$)/ {for(i=1;i<=NF;i++)if($i=="Leader:"){print $(i+1);exit}}')
case "$leader" in
 1) victim=kafka; admin=kafka-2 ;;
 2) victim=kafka-2; admin=kafka-3 ;;
 3) victim=kafka-3; admin=kafka-2 ;;
 *) echo "Invalid partition leader: $leader" >&2; exit 1 ;;
esac
trap '"${compose[@]}" start "$victim" >/dev/null || true' EXIT
"${compose[@]}" stop "$victim"
# Two remaining replicas must be in sync before testing write availability.
for i in {1..60}; do
 if replicas_ready 2; then break; fi
 sleep 2
done
replicas_ready 2
export AETHER_EXECUTION_ID="ha-during-$(date +%s)-$$"
AETHER_INGEST_ATTEMPTS=10 bash tests/execution-smoke.sh
curl -fsS -H "Authorization: Bearer $AETHER_OPERATOR_TOKEN" "http://127.0.0.1:8082/api/v1/incidents/$before" | jq -e '.state=="resolved" and .event_count==1 and .version==6'
"${compose[@]}" start "$victim"
for i in {1..90}; do
 if replicas_ready 3; then break; fi
 sleep 2
done
replicas_ready 3
trap - EXIT
echo "Three-broker Kafka survived partition-leader loss, approved workflow and replica recovery"
