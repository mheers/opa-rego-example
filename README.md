# opa-rego-example

Defines a simple OPA bundle compatible folder structure for rego policies and data.

Publishes a documentation to GitHub Pages: https://mheers.github.io/opa-rego-example/

## Run OPA server

```bash
opa run --server --log-level debug --addr :8181 ./bundle
```

## Run a query using curl

```bash
curl -X POST localhost:8181/v1/data/simple/authz -d '{"input": {"email": "marcel@heers.it", "password": "abcdef", "salary": 5, "card": "123-456-7890"}}' | jq
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

It also uses the [opa-tools](https://github.com/mheers/opa-tools) Docker image that contains all the necessary tools to run the pipeline.

Before running the pipeline, you need to set the following secrets in your GitHub repository:

- `REGISTRY_ACCESS_TOKEN`: The access token to authenticate with the registry.

Also, you need to adjust the following constants in the `ci/dagger/src/index.ts` file:

```ts
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
dagger call test-regos --bundle-directory ../bundle

# run build and push pipeline
export $(cat .env | xargs)
dagger call test-build-and-push-bundle \
   --bundle-directory ../bundle \
   --git-directory ../.git \
   --test-directory ../tests \
   --registry-token=env:REGISTRY_ACCESS_TOKEN

# run build and push an OPA demo image with batteries included
export $(cat .env | xargs)
dagger call build-and-push-opa-demo \
  --bundle-directory ../bundle \
  --git-directory ../.git \
  --test-directory ../tests \
  --docs-directory ../docs \
  --config-demo-file ../config-demo.yaml \
  --registry-token=env:REGISTRY_ACCESS_TOKEN
```

### Run the pipeline in local k8s/k3d cluster

```bash
docker pull registry.dagger.io/engine:v0.14.0
k3d image import registry.dagger.io/engine:v0.14.0
kubectl create -f ci/daggerengine/
kubectl port-forward -n dagger svc/daggerengine 1234 --address='0.0.0.0'
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:1234
```

#### Policy Documentation

```bash
dagger call get-documentation --bundle-directory ../bundle --git-directory ../.git --docs-directory ../docs export --path ./public
```

Publish the documentation to GitHub Pages: https://mheers.github.io/opa-rego-example/

## Dev

Install:

- [opa](https://www.openpolicyagent.org/docs/latest/#running-opa)
- [regal](https://docs.styra.com/regal)

# Monitoring

You can monitor the OPA server by querying the metrics endpoint:

```bash
curl -X GET localhost:8181/v1/metrics
```

This should return a list of metrics in Prometheus format.

# TODO
- [x] policy documentation using https://github.com/zenitysec/sphinx-rego
  - [x] pipeline
  - [x] publish to github pages
- [ ] signing using `policy build --signing-key=STRING`
- [ ] encrypting
