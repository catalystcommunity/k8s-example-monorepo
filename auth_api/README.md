# auth_api

This is a FastAPI based example project providing authn and authz functionality (at a basic level) for a monorepo example run in k8s with multiple services/tools.

See [the github repo](https://github.com/catalystcommunity/k8s-monorepo-example) for the rest of the project.

This sub-project should contain tests that use an actual postgres db but the only requirements for testing or running are environment variables.

If requirements needs updating, you have to regenerate the requirements_lock.txt for bazel using the following from inside this directory:

`uv pip compile pyproject.toml -o ../requirements_lock.txt`