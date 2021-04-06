package resource

import "github.com/cloudskiff/driftctl/pkg/dctlcty"

var normalizers = map[string]func(val *dctlcty.CtyAttributes){}

func AddNormalizer(key string, f func(val *dctlcty.CtyAttributes)) {
	normalizers[key] = f
}

func Normalizers() map[string]func(val *dctlcty.CtyAttributes) {
	return normalizers
}
