package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsCloudfrontDistribution) NormalizeForState() (resource.Resource, error) {
	r.normalizeNilPtr()
	return r, nil
}

func (r *AwsCloudfrontDistribution) NormalizeForProvider() (resource.Resource, error) {
	r.normalizeNilPtr()
	return r, nil
}

func (r *AwsCloudfrontDistribution) normalizeNilPtr() {
	if r.Aliases != nil && len(*r.Aliases) == 0 {
		r.Aliases = nil
	}
	if r.OriginGroup != nil && len(*r.OriginGroup) == 0 {
		r.OriginGroup = nil
	}

	// Way too dirty, we must find another way to normalize to nil fields in deeply nested struct
	// Here, this is for 3 fields, but in the CloudfrontDistribution resource there
	// could be more fields that we would need to do that...
	if r.DefaultCacheBehavior != nil {
		for i, b := range *r.DefaultCacheBehavior {
			if b.ForwardedValues != nil {
				for j, f := range *b.ForwardedValues {
					if f.Headers != nil && len(*f.Headers) == 0 {
						(*(*r.DefaultCacheBehavior)[i].ForwardedValues)[j].Headers = nil
					}
					if f.Cookies != nil {
						for k, c := range *f.Cookies {
							if c.WhitelistedNames != nil && len(*c.WhitelistedNames) == 0 {
								(*(*(*r.DefaultCacheBehavior)[i].ForwardedValues)[j].Cookies)[k].WhitelistedNames = nil
							}
						}
					}
				}
			}
		}
	}
	if r.Restrictions != nil {
		for i, v := range *r.Restrictions {
			if v.GeoRestriction != nil {
				for j, g := range *v.GeoRestriction {
					if g.Locations != nil && len(*g.Locations) == 0 {
						(*(*r.Restrictions)[i].GeoRestriction)[j].Locations = nil
					}
				}
			}
		}
	}
	if r.OrderedCacheBehavior != nil {
		for i, b := range *r.OrderedCacheBehavior {
			if b.LambdaFunctionAssociation != nil && len(*b.LambdaFunctionAssociation) == 0 {
				(*r.OrderedCacheBehavior)[i].LambdaFunctionAssociation = nil
			}
			if b.ForwardedValues != nil && len(*b.ForwardedValues) == 0 {
				(*r.OrderedCacheBehavior)[i].ForwardedValues = nil
			}
			if b.TrustedSigners != nil && len(*b.TrustedSigners) == 0 {
				(*r.OrderedCacheBehavior)[i].TrustedSigners = nil
			}
			if b.ForwardedValues != nil {
				for j, f := range *b.ForwardedValues {
					if f.Headers != nil && len(*f.Headers) == 0 {
						(*(*r.OrderedCacheBehavior)[i].ForwardedValues)[j].Headers = nil
					}
					if f.Cookies != nil {
						for k, c := range *f.Cookies {
							if c.WhitelistedNames != nil && len(*c.WhitelistedNames) == 0 {
								(*(*(*r.OrderedCacheBehavior)[i].ForwardedValues)[j].Cookies)[k].WhitelistedNames = nil
							}
						}
					}
				}
			}
			if b.FieldLevelEncryptionId != nil && *b.FieldLevelEncryptionId == "" {
				(*r.OrderedCacheBehavior)[i].FieldLevelEncryptionId = nil
			}
		}
	}
}
