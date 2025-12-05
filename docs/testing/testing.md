# Testing

## Unit & Integration tests

**When**: Unit (pure Go) and integration tests for provider logic.
**Where**: `internal/...`
**Add files in**: `*_test.go` next to the code.
**Run**: `make test`

These tests verify the internal logic of the provider without requiring a live 1Password connection. They include:

- Provider configuration and validation
- Resource and data source logic
- Utility functions (date parsing, utilities, etc.)

## E2E tests (Acceptance Tests)

**When**: End-to-end tests that verify the provider works correctly with a real 1Password account.
**Where**: `test/e2e/...`
**Add files in**: `*_test.go` files in `test/e2e/`
**Framework**: [Terraform Plugin Testing](https://github.com/hashicorp/terraform-plugin-testing)
**Run**: `make test-e2e`

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
