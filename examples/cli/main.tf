variable "op_account" {
  type = string
}
variable "op_password" {
  type = string
}

terraform {
  required_providers {
    onepassword = {
      source  = "terraform.example.com/local/onepassword"
      version = "~> 1.0.2"
    }
  }
}


provider "onepassword" {
  account  = var.op_account
  password = var.op_password
}

data "onepassword_item" "item" {
  vault = "Private"
  title = "Example"
}

output "password" {
  value = data.onepassword_item.item.section
}
