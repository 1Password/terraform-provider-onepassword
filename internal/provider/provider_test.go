package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `terraform {
	required_providers {
	  onepassword = {
		source  = "1Password/onepassword"
		version = "~> 1.3.0"
	  }
	}
  }
  # Configure the connection details for the Inventory service
  provider "onepassword" {
  }`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"onepassword": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func testAccProviderConfig(url string) string {
	return fmt.Sprintf(`terraform {
		required_providers {
		  onepassword = {
			source  = "1Password/onepassword"
			version = "~> 1.3.0"
		  }
		}
	  }
	  # Configure the connection details for the Inventory service
	  provider "onepassword" {
		url = "%s"
		token = "<PASSWORD>"
	  }`, url)
}
