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
import { dag, Directory, File, object, func, Container, Secret } from "@dagger.io/dagger"

const baseImage = "mheers/opa-tools:latest"
const registry = "registry-1.docker.io"
const repository = "mheers/opa-rego-example"
const tag = "1.0.0"
const username = "mheers"
const userDataURL = "https://github.com/mheers/opa-rego-example/releases/download/v0.0.1/data.json"

const opaImageSrc = "openpolicyagent/opa:0.70.0-static"
const opaImageDst = "docker.io/mheers/opa-demo:latest"

const docsImageSrc = "mheers/sphinx-rego:latest"

@object()
export class Ci {
  @func()
  async lintRegos(bundleDirectory: Directory): Promise<string> {
    return this.baseContainer(bundleDirectory)
      .withExec(["regal", "lint", "/bundle"]) // lint
      .stdout()
  }

  async checkRegos(bundleDirectory: Directory): Promise<string> {
    return this.baseContainer(bundleDirectory)
      .withExec(["opa", "check", "--strict", "/bundle"]) // check // TODO: add schema to check and run bench
      .stdout()
  }

  @func()
  async testRegos(bundleDirectory: Directory): Promise<string> {
    return this.baseContainer(bundleDirectory)
      .withExec(["opa", "test", "-v", "--coverage", "--format=json", "/bundle"]) // test
      .stdout()
  }

  @func()
  baseContainer(bundleDirectory: Directory): Container {
    return dag.container().from(baseImage)
      .withMountedDirectory("/bundle", bundleDirectory)
      .withWorkdir("/bundle")

      // download/replace user data from the api
      .withExec(["mkdir", "-p", "/bundle/users/"])
      .withExec(["wget", "-O", "/bundle/users/data.json", userDataURL])
  }

  @func()
  buildBundle(bundleDirectory: Directory, gitDirectory: Directory): Container {
    return this.baseContainer(bundleDirectory)
      .withMountedDirectory("/git/.git", gitDirectory)
      .withWorkdir("/git")

      // build the bundle
      .withExec(["sh", "-c", `policy build /bundle --revision $(git rev-parse HEAD) --ignore *_test.rego -t ${registry}/${repository}:${tag}`])
  }

  @func()
  async testBlackBox(bundleContainer: Container, testDir: Directory): Promise<string> {
    return bundleContainer
      .withMountedDirectory("/tests", testDir)
      .withExec(["mkdir", "-p", "/data"])
      .withWorkdir("/data")
      .withExec(["policy", "save", `${registry}/${repository}:${tag}`]) // save/export
      .withExec(["sh", "-c", "cp -r /tests/* /data/"])
      .withExec(["raygun", "execute", "--verbose", "--opa-log", "/tmp/opa.log", "."]) // blackbox test
      .stdout()
  }

  @func()
  async testBuildAndPushBundle(bundleDirectory: Directory, testDirectory: Directory, gitDirectory: Directory, registryToken: Secret): Promise<string> {
    await this.checkRegos(bundleDirectory)
    await this.lintRegos(bundleDirectory)
    await this.testRegos(bundleDirectory)
    const bundle = this.buildBundle(bundleDirectory, gitDirectory)

    await this.testBlackBox(bundle, testDirectory)

    return bundle
      .withSecretVariable("REGISTRY_ACCESS_TOKEN", registryToken)
      .withExec(["sh", "-c", `policy login -s ${registry} -u ${username} -p $REGISTRY_ACCESS_TOKEN`]) // login
      .withExec(["policy", "push", `${registry}/${repository}:${tag}`]) // push
      .stdout()
  }

  @func()
  buildBundleDocumentation(bundleDirectory: Directory, gitDirectory: Directory, docsDirectory: Directory): Container {
    const docsContainer = dag.container().from(docsImageSrc)

    return docsContainer
      .withMountedDirectory("/bundle", bundleDirectory)
      .withMountedDirectory("/git/.git", gitDirectory)
      .withMountedDirectory("/docs", docsDirectory)
      .withExec(["mkdir", "-p", "/work/build"])
      .withExec(["sh", "-c", "cp -r /bundle/ /work/source/"])
      .withExec(["sh", "-c", "cp -r /docs/* /work/source/"])
      .withExec(["sh", "-c", "sphinx-build /work/source/ /work/build/"])
  }

  @func()
  getDocumentation(bundleDirectory: Directory, gitDirectory: Directory, docsDirectory: Directory): Directory {
    return this.buildBundleDocumentation(bundleDirectory, gitDirectory, docsDirectory).directory("/work/build/")
  }


  @func()
  async buildAndPushOpaDemo(bundleDirectory: Directory, gitDirectory: Directory, configDemoFile: File, registryToken: Secret): Promise<string> {
    const bundleContainer = this.buildBundle(bundleDirectory, gitDirectory)
    const opaContainer = dag.container().from(opaImageSrc)

    const imageDigest = bundleContainer
      .withFile("/opa", opaContainer.file("/opa"))
      .withFile("/config.yaml", configDemoFile)
      .withExec(["mkdir", "-p", "/data"])
      .withWorkdir("/data")
      .withExec(["policy", "save", `${registry}/${repository}:${tag}`]) // save/export
      .withEntrypoint(["/opa", "run", "--server", "--log-level", "debug", "--addr", ":8181", "/data", "--config-file", "/config.yaml"])
      .withRegistryAuth(opaImageDst, username, registryToken)
      .publish(opaImageDst)

    return imageDigest
  }
}
