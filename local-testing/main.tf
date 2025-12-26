terraform {
  required_providers {
    onepassword = {
      source = "1Password/onepassword"
    }
  }
}

provider "onepassword" {
  # account = "B5test"
  # token = "a"
  # connect_token = "1"
  # token = "b"
}

data "onepassword_vault" "a" {
  name = "A"
}


resource "onepassword_item" "item1" {
  vault = data.onepassword_vault.a.uuid
  title = "item1"

  # password_wo         = "test-password-wo"
  # password_wo_version = 1

  section {
    label = "test-section"
    field {
      label = "test-field"
      type  = "STRING"
      value = "test-value"
    }
  }
}

# resource "onepassword_item" "item1" {
#   vault = data.onepassword_vault.a.uuid
#   title = "item1"

#   password_wo         = "test-password-wo"
#   password_wo_version = 1

#   section {
#     label = "test-section"
#     field {
#       label = "test-field"
#       type  = "STRING"
#       value = "test-value"
#     }
#   }
# }

# output "item1" {
# value = onepassword_item.item1.section[0].field[0].value

# resource "onepassword_item" "item2" {
#   vault = data.onepassword_vault.a.uuid
#   title = "re-create-issue"
#   password_recipe {}
#   tags = [1, 2]
# }


# resource "onepassword_item" "item3" {
#   vault = data.onepassword_vault.a.uuid
#   title = "re-create-issue"
#   tags  = [1, 2, 3, 4, 5, 6, 7, 8, 9]
# }
#
# resource "onepassword_item" "item4" {
#   vault = data.onepassword_vault.a.uuid
#   title = "re-create-issue"
#   password_recipe {}
#   tags = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
# }

# output "test" {
#   value     = data.onepassword_item.travel.file
#   sensitive = true
# }
