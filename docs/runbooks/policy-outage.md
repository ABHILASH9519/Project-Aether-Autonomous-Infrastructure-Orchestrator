# Local failure drill: policy decision service outage

Run only against a disposable local Compose stack. The drill intentionally stops OPA, not production infrastructure. Start the workflow profile with distinct operator and approver tokens as described in the README.

```sh
bash tests/policy-outage-smoke.sh
```

The script creates a unique synthetic incident, stops OPA, expects 503 on an authorized transition, checks that the incident remains at version zero, restarts OPA and verifies that the transition succeeds once. An EXIT trap attempts to restart OPA on failure. An interrupted shell or Docker daemon failure may require manual recovery:

```sh
docker compose -f deploy/compose/events.yaml --profile workflow start opa
```

Expected invariants: no workflow mutation while authorization is unavailable; read access remains authenticated; policy recovery requires no database reset. This drill verifies availability behavior, not production chaos tolerance. CI result is the source of truth.
