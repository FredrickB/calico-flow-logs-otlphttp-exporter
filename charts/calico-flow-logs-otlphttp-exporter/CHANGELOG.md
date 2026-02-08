# Changelog

## [0.12.5] - 2026-02-08

### Changes

Upgrade container image to `0.13.12`.

## [0.12.4] - 2026-02-06

### Changes

Upgrade container image to `0.13.11`.

## [0.12.3] - 2026-02-06

### Changes

Upgrade container image to `0.13.10`.

## [0.12.2] - 2026-02-06

### Changes

Upgrade container image to `0.13.9`.

## [0.12.1] - 2026-02-04

### Changes

Upgrade container image to `0.13.8`.

## [0.12.0] - 2026-01-20

### Breaking changes

Remove HPA from Helm chart.

## [0.11.4] - 2026-01-11

### Changes

Upgrade container image to `0.13.7`.

## [0.11.3] - 2026-01-10

### Changes

Upgrade container image to `0.13.6`.

## [0.11.2] - 2026-01-10

### Changes

Upgrade container image to `0.13.5`.

## [0.11.1] - 2026-01-10

### Changes

Upgrade container image to `0.13.4`.

## [0.11.0] - 2026-01-10

### Breaking changes

Remove default value for `env.OTEL_EXPORTER_OTLP_ENDPOINT`,
to retain backwards-compatibility with existing deployments,
add the following to `values.yaml`:

```yaml
env:
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://localhost:4318"
```

## [0.10.3] - 2026-01-05

### Changes

Upgrade container image to `0.13.3`.

## [0.10.2] - 2026-01-05

### Changes

Upgrade container image to `0.13.2`.

## [0.10.1] - 2026-01-04

### Changes

Upgrade container image to `0.13.1`.

## [0.10.0] - 2026-01-02

### Breaking changes

Upgrade to version `0.13.0` of binary, change
to `RECONNECT_WAIT_TIME_IN_MILLISECONDS` from
`RECONNECT_WAIT_TIME_IN_SECONDS`. Default
reconnect wait time remains at 5 seconds.

## [0.9.2] - 2026-01-02

### Fixes

Cast all environment variables to string as part
of rendering. Fixes the issue where passing
`RECONNECT_WAIT_TIME_IN_SECONDS` was always cast
as int.

## [0.9.1] - 2026-01-02 - [YANKED]

### Fixes

Change `RECONNECT_WAIT_TIME_IN_SECONDS` back to
string datatype.

## [0.9.0] - 2026-01-02 - [YANKED]

### Breaking changes

Environment variables are now passed as map instead
of array. Change `env` in `values.yaml` to be a map
instead of an array.

Old format:

```yaml
env:
  - name: GOLDMANE_HOST
    value: "goldmane:7443"
```

New format:

```yaml
env:
  GOLDMANE_HOST: "goldmane:7443"
```
