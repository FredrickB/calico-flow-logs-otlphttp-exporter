# Changelog

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
