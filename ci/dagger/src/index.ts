/**
 * A generated module for Ci functions
 *
 * This module has been generated via dagger init and serves as a reference to
 * basic module structure as you get started with Dagger.
 *
 * Two functions have been pre-created. You can modify, delete, or add to them,
 * as needed. They demonstrate usage of arguments and return types using simple
 * echo and grep commands. The functions can be called from the dagger CLI or
 * from one of the SDKs.
 *
 * The first line in this comment block is a short description line and the
 * rest is a long description with more detail on the module's purpose or usage,
 * if appropriate. All modules should have a short description.
 */
import { dag, Directory, object, func, Container, Secret } from "@dagger.io/dagger"

const baseImage = "mheers/opa-tools:latest"
const registry = "registry-1.docker.io"
const repository = "mheers/opa-rego-example"
const tag = "1.0.0"
const username = "mheers"

@object()
export class Ci {
  @func()
  async lintRegos(directoryArg: Directory): Promise<string> {
    return dag.container().from(baseImage)
      .withMountedDirectory("/bundle", directoryArg)
      .withWorkdir("/bundle")
      .withExec(["regal", "lint", "/bundle"]) // lint
      .stdout()
  }

  async checkRegos(directoryArg: Directory): Promise<string> {
    return dag.container().from(baseImage)
      .withMountedDirectory("/bundle", directoryArg)
      .withWorkdir("/bundle")
      .withExec(["opa", "check", "--strict", "/bundle"]) // check // TODO: add schema to check and run bench
      .stdout()
  }

  @func()
  async testRegos(directoryArg: Directory): Promise<string> {
    return dag.container().from(baseImage)
      .withMountedDirectory("/bundle", directoryArg)
      .withWorkdir("/bundle")
      .withExec(["opa", "test", "-v", "--coverage", "--format=json", "/bundle"]) // test
      .stdout()
  }

  @func()
  buildBundle(directoryArg: Directory): Container {
    return dag.container().from(baseImage)
      .withMountedDirectory("/bundle", directoryArg)
      .withWorkdir("/bundle")
      .withExec(["policy", "build", "/bundle", "--ignore", "*_test.rego", "-t", `${registry}/${repository}:${tag}`]) // build
  }

  @func()
  async testBuildAndPushBundle(directoryArg: Directory, registryToken: Secret): Promise<string> {
    await this.checkRegos(directoryArg)
    await this.lintRegos(directoryArg)
    await this.testRegos(directoryArg)
    return this.buildBundle(directoryArg)
      .withSecretVariable("REGISTRY_ACCESS_TOKEN", registryToken)
      .withExec(["sh", "-c", `policy login -s ${registry} -u ${username} -p $REGISTRY_ACCESS_TOKEN`]) // login
      .withExec(["policy", "push", `${registry}/${repository}:${tag}`]) // push
      .stdout()
  }
}
