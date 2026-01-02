# Changelog

## 0.9.1

### Fixes

Change `RECONNECT_WAIT_TIME_IN_SECONDS` back to
string datatype.

## 0.9.0

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
