package checks

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// CaptureItemTimestamps records the item timestamps reported by 1Password so a
// later step can compare against them.
func CaptureItemTimestamps(resourceName string, createdAt, updatedAt *time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		created, updated, err := itemTimestamps(s, resourceName)
		if err != nil {
			return err
		}

		*createdAt = created
		*updatedAt = updated

		return nil
	}
}

// VerifyItemTimestampsAfterUpdate verifies the timestamps describe the item as
// it now exists in 1Password: `created_at` is unchanged and `updated_at` moved
// forward to the time of the update. A write response that reports the previous
// modification time fails here.
func VerifyItemTimestampsAfterUpdate(resourceName string, createdAt, updatedAt *time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		created, updated, err := itemTimestamps(s, resourceName)
		if err != nil {
			return err
		}

		if !created.Equal(*createdAt) {
			return fmt.Errorf("created_at changed from %s to %s, expected it to stay the item's creation time",
				createdAt.Format(time.RFC3339), created.Format(time.RFC3339))
		}

		if !updated.After(*updatedAt) {
			return fmt.Errorf("updated_at is %s, expected it to be later than %s - the update was not reflected",
				updated.Format(time.RFC3339), updatedAt.Format(time.RFC3339))
		}

		return nil
	}
}

func itemTimestamps(s *terraform.State, resourceName string) (createdAt, updatedAt time.Time, err error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return createdAt, updatedAt, fmt.Errorf("resource not found: %s", resourceName)
	}

	createdAt, err = parseTimestamp(rs.Primary.Attributes, "created_at")
	if err != nil {
		return createdAt, updatedAt, err
	}

	updatedAt, err = parseTimestamp(rs.Primary.Attributes, "updated_at")
	if err != nil {
		return createdAt, updatedAt, err
	}

	return createdAt, updatedAt, nil
}

func parseTimestamp(attributes map[string]string, attr string) (time.Time, error) {
	value, ok := attributes[attr]
	if !ok || value == "" {
		return time.Time{}, fmt.Errorf("%s is not set", attr)
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s is %q, which is not an RFC 3339 timestamp: %w", attr, value, err)
	}

	return parsed, nil
}
