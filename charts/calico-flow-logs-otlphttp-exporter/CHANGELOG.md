# Changelog

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
