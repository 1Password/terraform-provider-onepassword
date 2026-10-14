# Testing

The provider has two test tiers. Every behavior change should include tests in at least one of them — use this table to decide where your tests belong:

| If you changed... | Add | Where |
|---|---|---|
| Pure logic: model conversions, utilities, validators, SSH key handling | Unit test | `internal/onepassword/...` |
| A resource, data source, or ephemeral resource: schema, CRUD, plan modifiers, attribute validators | Integration test against the mock Connect server | `internal/provider/...` |
| Behavior that only shows against the real 1Password API: authentication modes, item lifecycle, import, drift detection | E2E test | `test/e2e/...` |
| Documentation or examples only | No tests | — |

Naming conventions:

- Pure unit tests: `TestXxx`, table-driven.
- Terraform-level tests (integration and E2E): `TestAcc*` prefix, built on [Terraform Plugin Testing](https://github.com/hashicorp/terraform-plugin-testing).

## Unit & Integration tests

**Where**: `internal/...`
**Add files in**: `*_test.go` next to the code.
**Run**: `make test`

These tests verify the internal logic of the provider without requiring a live 1Password connection. They include:

- Unit tests in `internal/onepassword/...` covering model conversions and utility functions in isolation.
- Integration tests in `internal/provider/...` exercising resources, data sources, and ephemeral resources through the plugin testing framework against a local mock of the 1Password Connect API — no credentials needed.

> **Note**: `make test` sets `TF_ACC=1`, which the Terraform plugin tests require. Running plain `go test ./...` silently skips most of them.

## E2E tests (Acceptance Tests)

**Where**: `test/e2e/...`
**Add files in**: `*_test.go` files in `test/e2e/`
**Framework**: [Terraform Plugin Testing](https://github.com/hashicorp/terraform-plugin-testing)
**Run**: `make test-e2e`

These tests run the provider against a real 1Password account, covering full item lifecycle, import, and drift detection under each authentication mode (Connect, service account, and desktop app).

E2E tests run automatically on internal PRs. For fork PRs, a maintainer triggers them with an `/ok-to-test` comment — see the [Fork PR Testing Guide](fork-pr-testing.md).

## Setup to run e2e tests locally
1. Set `OP_CONNECT_TOKEN`, `OP_CONNECT_HOST` and `OP_SERVICE_ACCOUNT_TOKEN` env variables.
   - Optionally can set `TF_LOG` and `TF_LOG_PATH` for debugging. 
2. `make test-e2e` to run e2e tests.

Other supported commands:
- `make test-e2e-service-account` - to run e2e tests using a service account only.
  - set `OP_SERVICE_ACCOUNT_TOKEN` env variable before run.
- `make test-e2e-connect` - to run e2e tests using Connect only.
  - set `OP_CONNECT_TOKEN`, `OP_CONNECT_HOST` env variables before run.
- `make test-e2e-account` - to run e2e test using desktop app authentication only.
  - set `OP_ACCOUNT` and `OP_TEST_VAULT_NAME` env variables before run.
