package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsLambdaFunction) NormalizeForProvider() (resource.Resource, error) {
	r.normalizeEmptyStringPtr()

	return r, nil
}

func (r *AwsLambdaFunction) NormalizeForState() (resource.Resource, error) {
	r.normalizeEmptyStringPtr()

	return r, nil
}

func (r *AwsLambdaFunction) normalizeEmptyStringPtr() {
	if r.CodeSigningConfigArn != nil && *r.CodeSigningConfigArn == "" {
		r.CodeSigningConfigArn = nil
	}

	if r.ImageUri != nil && *r.ImageUri == "" {
		r.ImageUri = nil
	}

	if r.PackageType != nil && *r.PackageType == "" {
		r.PackageType = nil
	}

	if r.SigningJobArn != nil && *r.SigningJobArn == "" {
		r.SigningJobArn = nil
	}

	if r.SigningProfileVersionArn != nil && *r.SigningProfileVersionArn == "" {
		r.SigningProfileVersionArn = nil
	}
}
