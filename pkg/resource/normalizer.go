package resource

var normalizers = map[string]func(val *map[string]interface{}){}

func AddNormalizer(key string, f func(val *map[string]interface{})) {
	normalizers[key] = f
}

func Normalizers() map[string]func(val *map[string]interface{}) {
	return normalizers
}
