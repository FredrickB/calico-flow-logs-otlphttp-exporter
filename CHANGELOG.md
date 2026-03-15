# Changelog

## [0.13.14] - 2026-03-15

### Changes

Bump deps

## [0.13.13] - 2026-02-18

### Changes

Bump deps

## [0.13.12] - 2026-02-08

### Changes

`main.go` is now placed in the root
dir instead of under `cmd/`.

updated `Makefile` and fixed typos.

## [0.13.11] - 2026-02-06

### Changes

Bump deps

## [0.13.10] - 2026-02-06

### Changes

Bump deps

## [0.13.9] - 2026-02-06

### Changes

Bump deps

## [0.13.8] - 2026-02-04

### Changes

- Bump deps
- Change `semconv` to schema version `v1.39.0`

## [0.13.7] - 2026-01-11

### Changes

Refactored internal protobuf
related code.

## [0.13.6] - 2026-01-10

### Changes

Print version of binary and Goldmane
protobuf version as part of startup.

## [0.13.5] - 2026-01-10

### Changes

Checkin grpc-generated protobuf code
from Goldmane.

## [0.13.4] - 2026-01-10

### Changes

Upgrade to Go `1.25`.

## [0.13.3] - 2026-01-05

### Changes

Bump deps.

## [0.13.2] - 2026-01-05

### Changes

Refactored internals.

## [0.13.1] - 2026-01-04

### Changes

Cleaned up dependencies.

## [0.13.0] - 2026-01-02

### Breaking changes

`RECONNECT_WAIT_TIME_IN_SECONDS` has been removed
in favor of `RECONNECT_WAIT_TIME_IN_MILLISECONDS`
has been introduced. `RECONNECT_WAIT_TIME_IN_MILLISECONDS`
takes milliseconds to wait in-between reconnect
attempts.
