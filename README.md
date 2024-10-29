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

## CI/CD Pipeline

This repository contains a GitHub Actions pipeline that runs the rego tests, builds and pushes the bundle to an OCI registry.

The pipeline leverages [dagger](https://dagger.io/) that allows the development of CI/CD pipelines independently of the CI/CD platform and can be run locally but also on GitHub Actions, GitLab CI, Jenkins, etc.

Before running the pipeline, you need to set the following secrets in your GitHub repository:

- `REGISTRY_ACCESS_TOKEN`: The access token to authenticate with the registry.

Also, you need to adjust the following constants in the `ci/dagger/src/index.ts` file:

```ts
const baseImage = "mheers/opa-tools:latest"
const registry = "registry-1.docker.io"
const repository = "mheers/opa-policy"
const tag = "1.0.0"
const username = "mheers"
```

### Run the pipeline locally

Before running the pipeline locally, you need to install the `dagger` CLI. You can find the installation instructions [here](https://docs.dagger.io/install).

You also need to put a `.env` file in the `ci/` directory with the following content:

```bash
REGISTRY_ACCESS_TOKEN=<your-registry-access-token>
```

You can then run the pipeline locally using the following command:

```bash
cd ci/

# run test pipeline
dagger call test-regos --directory-arg ../bundle

# run build and push pipeline
dagger call test-build-and-push-bundle --directory-arg ../bundle --registry-token=env:REGISTRY_ACCESS_TOKEN
```
