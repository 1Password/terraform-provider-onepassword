# Testing

## Unit & Integration tests
**When**: Unit (pure Go) and integration tests for provider logic.
**Where**: `internal/...`
**Add files in**: `*_test.go` next to the code.
**Run**: `make test`

These tests verify the internal logic of the provider without requiring a live 1Password connection. They include:
- Provider configuration and validation
- Resource and data source logic
- Utility functions (date parsing, CLI utilities, etc.)

## E2E tests (Acceptance Tests)
**When**: End-to-end tests that verify the provider works correctly with a real 1Password account.
**Where**: `test/e2e/...`
**Add files in**: `*_test.go` files in `test/e2e/`
**Framework**: [Terraform Plugin Testing](https://github.com/hashicorp/terraform-plugin-testing)

These tests create, read, update, and delete actual resources in a 1Password vault to verify the provider's behavior. The test suite covers.

**Local prep**:
1. Ensure you have a 1Password account with a service account token
2. Create or use an existing test vault in your 1Password account
3. Set required environment variables:
   ```bash
   export OP_SERVICE_ACCOUNT_TOKEN=<your-service-account-token>
   ```
4. Optional environment variables for debugging:
   ```bash
   export TF_LOG=INFO          # Set Terraform logging level
   export TF_LOG_PATH=./e2e.log # Path to log file
   ```
5. Run the tests:
   ```bash
   make test-e2e
   ```
