package dctlcty

type Metadata struct {
	tags       map[string]string
	normalizer func(val *CtyAttributes)
}

var resourcesMetadata = map[string]Metadata{}

func SetMetadata(typ string, tags map[string]string, f func(val *CtyAttributes)) {
	resourcesMetadata[typ] = Metadata{
		tags:       tags,
		normalizer: f,
	}
}
