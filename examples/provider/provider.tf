provider "onepassword" {
  url                   = "http://localhost:8080"
  token                 = "CONNECT_TOKEN"
  service_account_token = "SERVICE_ACCOUNT_TOKEN"
  account               = "ACCOUNT_ID_OR_SIGN_IN_ADDRESS"
  op_cli_path           = "OP_CLI_PATH"
}
