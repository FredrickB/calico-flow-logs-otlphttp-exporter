# Contributing to calico-flow-logs-otlphttp-exporter

> [!NOTE]
> Currently in development, subject to change

You're welcome to contribute to the project!

## Goal of the calico-flow-logs-otlphttp-exporter

The intention is to keep `calico-flow-logs-otlphttp-exporter`
small and focus on doing one thing well, and that is to forward
flow logs from Calico to anything thats compatible with OTLP/HTTP.
Log storage, processing, and exporting to other formats is left to
tools such as
[OpenTelemetry Collector](https://opentelemetry.io/docs/collector/).

## Creating an issue

Bugs, feature requests, suggestions, etc are all recorded
as issues in the repository.
File an issue [here](https://github.com/FredrickB/calico-flow-logs-otlphttp-exporter/issues),
provide a clear title and description.

## Contributing code

Code submissions are welcome, to create a Pull Request:

1. Fork the repository from the `main` branch
1. Setup a development environment: [README.md#development](./README.md#development)
1. Implement changes in a branch reflecting the change (bug, feature, etc)
   1. Use [Conventional Commits](https://www.conventionalcommits.org/)
   1. Adhere to the [CONTRIBUTING.md#project-structure](./CONTRIBUTING.md#project-structure) when making changes
1. Ensure changes work with all versions of Calico in the
[README.md#compatibility-matrix](./README.md#compatibility-matrix) marked `Yes` in the `Compatible` column
   1. Ensure that the [README.md#datamodel](./README.md#datamodel)
   still reflects the structure of the logs
1. Bump the `Chart.version` and `Chart.appVersion` in the
[Chart.yaml](./charts/calico-flow-logs-otlphttp-exporter/Chart.yaml) in a separate pull request once the new container image version has been released
1. Ensure workflows for linting and CI pass
1. Update the documentation if necessary
1. Create a Pull Request to `main` from your fork
1. Wait for review by the maintainers

### Project structure

```bash
.
├── Dockerfile  # Container image definition
├── LICENSE
├── Makefile
├── README.md
├── certs       # Goldmane certs (only used for development)
├── charts      # Helm charts
├── cmd         # contains main.go for executables
├── docs        # documentation
├── gen         # code generated from .proto files
├── go.mod
├── go.sum
├── hack        # utility dir for scripts, docker compose files, ect
├── internal    # code
├── out         # build output
└── protos      # .proto files
```
