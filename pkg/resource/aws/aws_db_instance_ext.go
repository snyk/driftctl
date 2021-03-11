package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsDbInstance) NormalizeForState() (resource.Resource, error) {
	if r.SnapshotIdentifier != nil && *r.SnapshotIdentifier == "" {
		r.SnapshotIdentifier = nil
	}
	if r.AllowMajorVersionUpgrade != nil && !*r.AllowMajorVersionUpgrade {
		r.AllowMajorVersionUpgrade = nil
	}
	if r.ApplyImmediately != nil && !*r.ApplyImmediately {
		r.ApplyImmediately = nil
	}
	return r, nil
}

func (r *AwsDbInstance) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}
