package uuid

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/1Password/terraform-provider-onepassword/v2/test/e2e/utils/cleanup"
)

// CaptureItemUUID captures the UUID of a resource item
func CaptureItemUUID(t *testing.T, resourceName string, uuidPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		*uuidPtr = rs.Primary.Attributes["uuid"]

		if *uuidPtr == "" {
			return fmt.Errorf("UUID is empty")
		}

		return nil
	}
}

// VerifyItemUUIDUnchanged verifies that the resource UUID matches the expected UUID
func VerifyItemUUIDUnchanged(t *testing.T, resourceName string, expectedUUID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		currentUUID := rs.Primary.Attributes["uuid"]

		if currentUUID == "" {
			return fmt.Errorf("UUID is empty")
		}

		if currentUUID != *expectedUUID {
			return fmt.Errorf("UUID changed from %s to %s - resource was replaced instead of updated", *expectedUUID, currentUUID)
		}

		return nil
	}
}

// CaptureItemUUIDAndRegisterCleanup captures the UUID and registers cleanup
func CaptureItemUUIDAndRegisterCleanup(t *testing.T, resourceName string, uuidPtr *string, vaultID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		captureUUIDFunc := CaptureItemUUID(t, resourceName, uuidPtr)
		err := captureUUIDFunc(s)
		if err != nil {
			return err
		}

		cleanup.RegisterItemCleanup(t, *uuidPtr, vaultID)
		return nil
	}
}
