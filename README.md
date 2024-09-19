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

## Contributing

For now, outside contributors need to talk to TodPunk in the Catalyst Community Discord or the [Forge Utah Slack](https://forgeutah.tech) and we may have guides for more self-service options later.

## License

We are releasing all of this as MIT Licensed code. We don't care what you do with it, use it for the basis of your startup if you want, or fork pieces and use them to build some open source thing. The point is learning.

Copyright Tod Hansmann
