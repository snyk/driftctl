package helpers

func SafeDelete(given *map[string]interface{}, path []string) {
	val := *given
	for i, key := range path {
		_, exists := val[key]
		if !exists {
			return
		}
		if i == len(path)-1 {
			delete(val, key)
		}
	}
}
