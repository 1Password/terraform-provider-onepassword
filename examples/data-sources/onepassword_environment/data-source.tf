# Read environment variables from a 1Password Environment.
# Requires service account or desktop app authentication (not 1Password Connect).
data "onepassword_environment" "example" {
  environment_id = "your-environment-id"
}

# Use the variables map in a resource or output
output "env_variables" {
  value     = data.onepassword_environment.example.variables
  sensitive = true
}
