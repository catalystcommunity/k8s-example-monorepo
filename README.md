# k8s-example-monorepo

An example for learning monorepos of some complexity with k8s deployment local and production

## Goals

We wish to have a repository for using in community education efforts around kubernetes and monorepos. This should be applicable to cloud native thinking both for personal projects and for professional applications, and all the building blocks are decoupled enough they can be learned from as a whole while carried forward without the rest.

For example, the auth API will use a backend database with the postgres protocol, but it will only know it is a URL and that protocol, and won't care if it's in the same network, if it's deployed the same, or if it is even really postgres. Same with the other components, where the auth API will not know about the webapp, so if someone were to write a different webapp that used the same API, nothing else should care.

## Layout

Several directories will house subprojects, and we'll evolve (and update this readme) as we go. At current we are thinking along the lines of:

- /coredb : All DB related functionality for handling definitions, migrations, etc. Some of this might be generated, much of it not, and it will largely be raw sql files for migrations using `goose` for applying those.
- /auth_api : A python API built on FastAPI for handling user authentication and authorization needs. It will be production ready.
- /app_api : The API for whatever app thing we're building, using Go and very little third party dependencies, maybe grpc based if we get to protos.
- /web_app : A frontend for the web, built in whatever we decide to use for the frontend. Undecided there.
- /cli : A CLI built in Rust for some interactions/tooling, which may be exposed by the root `tools` script or might not. We'll see.

The only assumption is that something else has setup the kubernetes cluster, but it will not care if it's local or public. We'll likely have a `skaffold` workflow for local dev if desired, with several profiles for different use cases like building the webapp locally with just the API running. We will, however, have all the required things to go in k8s and make this a production ready app.

## K8s deployables

At the very least we want this to have some options for common production environments. By default the setup will use minimal deployments and to make it "production ready" people will want to specify different values like requests/limits for their services, but we do want to have options for deploying the Grafana/Prometheus/Loki stack for obervability as they are efficient, open source, and very well known in the community.

Deployment options will likely be based on helm.

## Bazel

To build and test the entire project, run the following commands:

```shell
bazel build //...
bazel test //...
```

### Gazelle

We are using Gazelle to manage the BUILD files in the project. Gazelle is a tool that will generate BUILD files for
Bazel from the source code. This is to ensure that the BUILD files are up to date with the source code, and that we
don't have to manually manage the BUILD files. To update dependencies and BUILD files, run the following commands:

```shell
bazel run //:gazelle
bazel mod tidy
```

### Bazelisk

We are using Bazel for building and testing. The specific bazel version and setup is managed by `bazelisk`, which is a
wrapper around bazel that will download the correct version for the project. This is to ensure that everyone is using 
the same version of bazel, and that we can update the version without breaking everyone's build.
https://github.com/bazelbuild/bazelisk?tab=readme-ov-file#installation

Bazelisk is controlled by the `.bazelversion` file in the root of the project. To update the version of bazel, update 
the `.bazelversion`, and the next time you run bazel, it will download the new version. In `.bazeliskrc` we have set
a wrapper around the core bazel command to use [Aspect CLI](https://docs.aspect.build/cli/) (https://github.com/aspect-build/aspect-cli).
Aspect CLI adds some convenience functions, plus an extensible plugin system if we need to add more functionality.

### .bazelrc

We have a `.bazelrc` file that sets up some defaults for the project. This is to ensure that everyone is using the 
same settings for building and testing. We are using [opinionated presets from Aspect](https://docs.aspect.build/guides/bazelrc) 
to manage these settings. The presets are stored in the `.aspect/bazelrc` directory. `bazel test //...` enforces that
those files match the versioned presets. To update the presets, run the following command:

```shell
bazel run //.aspect/bazelrc:update_aspect_bazelrc_presets
```

If you want persistent personal settings that are not shared with the project, you can create a `.user.bazelrc` file
in the root of the project. This gitignored file will be included after the `.bazelrc` file, and can be used to 
override any settings that you need to change.

## Contributing

For now, outside contributors need to talk to TodPunk in the Catalyst Community Discord or the [Forge Utah Slack](https://forgeutah.tech) and we may have guides for more self-service options later.

## License

We are releasing all of this as MIT Licensed code. We don't care what you do with it, use it for the basis of your startup if you want, or fork pieces and use them to build some open source thing. The point is learning.

Copyright Tod Hansmann
