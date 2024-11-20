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

	opaImageSrc = "openpolicyagent/opa:0.70.0-static"
	opaImageDst = "docker.io/mheers/opa-demo:latest"

	docsImageSrc = "mheers/sphinx-rego:latest"
)

func (m *Ci) BaseContainer(bundleDirectory *dagger.Directory) *dagger.Container {
	return dag.Container().From(baseImage).
		WithMountedDirectory("/bundle", bundleDirectory).
		WithWorkdir("/bundle").

		// download/replace user data from the api
		WithExec([]string{"mkdir", "-p", "/bundle/users/"}).
		WithExec([]string{"wget", "-O", "/bundle/users/data.json", userDataURL})
}

func (m *Ci) LintRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory).
		WithExec([]string{"regal", "lint", "/bundle"}). // lint
		Stdout(context.Background())
}

func (m *Ci) CheckRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory).
		WithExec([]string{"opa", "check", "--strict", "/bundle"}). // check // TODO: add schema to check and run bench
		Stdout(context.Background())
}

func (m *Ci) TestRegos(bundleDirectory *dagger.Directory) (string, error) {
	return m.BaseContainer(bundleDirectory).
		WithExec([]string{"opa", "test", "-v", "--coverage", "--format=json", "/bundle"}). // test
		Stdout(context.Background())
}

func (m *Ci) BuildBundle(bundleDirectory, gitDirectory *dagger.Directory) *dagger.Container {
	return m.BaseContainer(bundleDirectory).
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

	bundle := m.BuildBundle(bundleDirectory, gitDirectory)

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

func (m *Ci) BuildAndPushOpaDemo(bundleDirectory, gitDirectory *dagger.Directory, configDemoFile *dagger.File, registryToken *dagger.Secret) (string, error) {
	bundleContainer := m.BuildBundle(bundleDirectory, gitDirectory)
	opaContainer := dag.Container().From(opaImageSrc)

	return bundleContainer.
		WithFile("/opa", opaContainer.File("/opa")).
		WithFile("/config.yaml", configDemoFile).
		WithExec([]string{"mkdir", "-p", "/data"}).
		WithWorkdir("/data").
		WithExec([]string{"policy", "save", fmt.Sprintf("%s/%s:%s", registry, repository, tag)}). // save/export
		WithEntrypoint([]string{"/opa", "run", "--server", "--log-level", "debug", "--addr", ":8181", "/data", "--config-file", "/config.yaml"}).
		WithRegistryAuth(opaImageDst, username, registryToken).
		Publish(context.Background(), opaImageDst)
}
