# opa-rego-example

Defines a simple OPA bundle compatible folder structure for rego policies and data.

## Run OPA server

```bash
opa run --server --log-level debug --addr :8181 ./bundle
```

## Run a query using curl

```bash
curl -X POST localhost:8181/v1/data/simple/authz -d '{"input": {"username": "marcel"}}' | jq
```

This should return:

```json
{
  "result": {
    "allow": true,
    "amountAllowed": true,
  }
}
```
