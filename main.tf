############################################################
# Variable for vault
############################################################

terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
    }
  }
}

provider "onepassword" {
}

variable "test_vault_id" {
  type        = string
  default     = "bbucuyq2nn4fozygwttxwizpcy"
}


############################################################
# 4. FIELD MATRIX — covers ALL field types
############################################################

resource "onepassword_item" "field_matrix" {
  vault    = var.test_vault_id
  title    = "TF Test - Field Types & Purposes"
  category = "login"





}


