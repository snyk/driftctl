package scaleway

import (
	"fmt"
)

// IDs of some resources are regional, meaning the same ID (uuid) can exist in multiple regions
// To distinguish them, the resources defined in Terraform have their ID prepended with the region where they belong
func getRegionalID(region, id string) string {
	return fmt.Sprintf("%s/%s", region, id)
}
