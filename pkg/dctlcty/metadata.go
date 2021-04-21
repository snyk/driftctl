package dctlcty

import (
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/sirupsen/logrus"
)

type AttributeMetadata struct {
	Configshema configschema.Attribute
	JsonString  bool
}

type ResourceMetadata struct {
	AttributeMetadata map[string]AttributeMetadata
	Normalizer        func(val *CtyAttributes)
}

var resourcesMetadata = map[string]*ResourceMetadata{}

func SetMetadata(typ string, tags map[string]string, f func(val *CtyAttributes)) {
	resourcesMetadata[typ] = &ResourceMetadata{
		AttributeMetadata: map[string]AttributeMetadata{},
		Normalizer:        f,
	}
}

func AddMetadata(resourceType string, metadata *ResourceMetadata) {
	resourcesMetadata[resourceType] = metadata
}

func UpdateMetadata(typ string, metadatas map[string]func(metadata *AttributeMetadata)) {
	for s, f := range metadatas {
		metadata, exist := resourcesMetadata[typ]
		if !exist {
			logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set metadata, no schema found")
			return
		}
		m := (*metadata).AttributeMetadata[s]
		f(&m)
		(*metadata).AttributeMetadata[s] = m
	}
}

func SetNormalizer(typ string, normalizer func(val *CtyAttributes)) {
	metadata, exist := resourcesMetadata[typ]
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set metadata, no schema found")
		return
	}
	(*metadata).Normalizer = normalizer
}
