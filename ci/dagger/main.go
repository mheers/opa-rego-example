// A generated module for Ci functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
)

type Ci struct{}

const (
	baseImage   = "mheers/opa-tools:latest"
	registry    = "registry-1.docker.io"
	repository  = "mheers/opa-rego-example"
	tag         = "1.0.0"
	username    = "mheers"
	userDataURL = "https://github.com/mheers/opa-rego-example/releases/download/v0.0.1/data.json"

	opaImageSrc        = "openpolicyagent/opa:0.70.0-static"
	httpServerImageSrc = "projectdiscovery/simplehttpserver:latest"
	opaImageDst        = "docker.io/mheers/opa-demo:latest"

	docsImageSrc       = "mheers/sphinx-rego:latest"
	playgroundImageSrc = "mheers/opa-live-playground:latest"

	daggerImageSrc = "mheers/dagger-tools:v0.14.0"
	daggerImageDst = "mheers/opa-rego-example-ci"
)

// builds and pushes the container for the ci pipeline -> self contained ci pipeline - defined in this pipeline
func (m *Ci) BuildCiImage(repoDirectory *dagger.Directory, registryToken *dagger.Secret) (string, error) {
	imageDst := fmt.Sprintf("%s/%s:%s", registry, daggerImageDst, tag)
	return dag.Container().From(daggerImageSrc).
		WithMountedDirectory("/repo", repoDirectory).
		WithWorkdir("/repo/ci").
		WithRegistryAuth(imageDst, username, registryToken).
		Publish(context.Background(), imageDst)
}

func (m *Ci) BaseContainer(bundleDirectory *dagger.Directory, useExternalUserData bool) *dagger.Container {
	c := dag.Container().From(baseImage).
		WithMountedDirectory("/bundle", bundleDirectory).
		WithWorkdir("/bundle")

	if useExternalUserData {
		// download/replace user data from the api
		c.
			WithExec([]string{"mkdir", "-p", "/bundle/users/"}).
			WithExec([]string{"wget", "-O", "/bundle/users/data.json", userDataURL})
	}

	return c
}

func (m *Ci) LintRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory, false).
		WithExec([]string{"regal", "lint", "/bundle"}). // lint
		Stdout(context.Background())
}

func (m *Ci) CheckRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory, false).
		WithExec([]string{"opa", "check", "--strict", "/bundle"}). // check // TODO: add schema to check and run bench
		Stdout(context.Background())
}

func (m *Ci) TestRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory, false).
		WithExec([]string{"opa", "test", "-v", "--coverage", "--format=json", "/bundle"}). // test
		Stdout(context.Background())
}

func (m *Ci) BuildBundle(bundleDirectory, gitDirectory *dagger.Directory, useExternalUserData bool) *dagger.Container {
	return m.BaseContainer(bundleDirectory, useExternalUserData).
		WithMountedDirectory("/git/.git", gitDirectory).
		WithWorkdir("/git").

		// build the bundle
		WithExec([]string{"sh", "-c", fmt.Sprintf("policy build /bundle --revision $(git rev-parse HEAD) --ignore *_test.rego -t %s/%s:%s", registry, repository, tag)})
}

func (m *Ci) TestBlackBox(bundleContainer *dagger.Container, testDir *dagger.Directory) (string, error) {
	return bundleContainer.
		WithMountedDirectory("/tests", testDir).
		WithExec([]string{"mkdir", "-p", "/data"}).
		WithWorkdir("/data").
		WithExec([]string{"policy", "save", fmt.Sprintf("%s/%s:%s", registry, repository, tag)}). // save/export.
		WithExec([]string{"sh", "-c", "cp -r /tests/* /data/"}).
		WithExec([]string{"raygun", "execute", "--verbose", "--opa-log", "/tmp/opa.log", "."}). // blackbox test.
		Stdout(context.Background())
}

func (m *Ci) TestBuildAndPushBundle(bundleDirectory, testDirectory, gitDirectory *dagger.Directory, registryToken *dagger.Secret) (string, error) {
	result, err := m.CheckRegos(bundleDirectory)
	if err != nil {
		return "", err
	}
	fmt.Println(result)

	result, err = m.LintRegos(bundleDirectory)
	if err != nil {
		return "", err
	}
	fmt.Println(result)

	result, err = m.TestRegos(bundleDirectory)
	if err != nil {
		return "", err
	}
	fmt.Println(result)

	bundle := m.BuildBundle(bundleDirectory, gitDirectory, true)

	m.TestBlackBox(bundle, testDirectory)

	return bundle.
		WithSecretVariable("REGISTRY_ACCESS_TOKEN", registryToken).
		WithExec([]string{"sh", "-c", fmt.Sprintf("policy login -s %s -u %s -p $REGISTRY_ACCESS_TOKEN", registry, username)}). // login
		WithExec([]string{"policy", "push", fmt.Sprintf("%s/%s:%s", registry, repository, tag)}).                              // push
		Stdout(context.Background())
}

func (m *Ci) BuildBundleDocumentation(bundleDirectory, gitDirectory, docsDirectory *dagger.Directory) *dagger.Container {
	return dag.Container().From(docsImageSrc).
		WithMountedDirectory("/bundle", bundleDirectory).
		WithMountedDirectory("/git/.git", gitDirectory).
		WithMountedDirectory("/docs", docsDirectory).
		WithExec([]string{"mkdir", "-p", "/work/build"}).
		WithWorkdir("/work").
		WithExec([]string{"sh", "-c", "cp -r /bundle/simple/ /work/source/"}).
		WithExec([]string{"sh", "-c", "cp -r /docs/* /work/"}).
		WithExec([]string{"sh", "-c", "sphinx-build . /work/build/"})
}

func (m *Ci) GetDocumentation(bundleDirectory, gitDirectory, docsDirectory *dagger.Directory) *dagger.Directory {
	return m.BuildBundleDocumentation(bundleDirectory, gitDirectory, docsDirectory).Directory("/work/build/")
}

func (m *Ci) BuildAndPushOpaDemo(bundleDirectory, gitDirectory, docsDirectory *dagger.Directory, configDemoFile *dagger.File, registryToken *dagger.Secret) (string, error) {
	bundleContainer := m.BuildBundle(bundleDirectory, gitDirectory, false)
	opaContainer := dag.Container().From(opaImageSrc)
	playgroundContainer := dag.Container().From(playgroundImageSrc)
	simpleHTTPServerContainer := dag.Container().From(httpServerImageSrc)
	docs := m.GetDocumentation(bundleDirectory, gitDirectory, docsDirectory)

	return bundleContainer.
		WithFile("/opa", opaContainer.File("/opa")).
		WithFile("/opa-live-playground", playgroundContainer.File("/opa-live-playground")).
		WithFile("/config.yaml", configDemoFile).
		WithFile("/simplehttpserver", simpleHTTPServerContainer.File("/usr/local/bin/simplehttpserver")).
		WithDirectory("/docs", docs).
		// entrypoint for the opa container with EOF
		WithNewFile("/entrypoint.sh", `#!/bin/bash
set -eo pipefail

echo "Starting docs"
/simplehttpserver -path /docs -listen 0.0.0.0:8080 &

echo "Starting opa live playground"
export OPA_URL=http://localhost:8181
/opa-live-playground &

echo "Starting opa"
exec /opa "$@"
`, dagger.ContainerWithNewFileOpts{Permissions: int(0755)}).
		WithExec([]string{"mkdir", "-p", "/data"}).
		WithWorkdir("/data").
		WithExec([]string{"policy", "save", fmt.Sprintf("%s/%s:%s", registry, repository, tag)}). // save/export
		WithEntrypoint([]string{"/entrypoint.sh"}).
		WithDefaultArgs([]string{"run", "--server", "--log-level", "debug", "--addr", ":8181", "/data", "--config-file", "/config.yaml"}).
		WithRegistryAuth(opaImageDst, username, registryToken).
		Publish(context.Background(), opaImageDst)
}
