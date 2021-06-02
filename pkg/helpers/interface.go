package helpers

import "strings"

func Join(elems []interface{}, sep string) string {
	firstElemt, ok := elems[0].(string)
	if !ok {
		panic("cannot join a slice that contains something else than strings")
	}
	switch len(elems) {
	case 0:
		return ""
	case 1:

		return firstElemt
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i].(string))
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(firstElemt)
	for _, s := range elems[1:] {
		b.WriteString(sep)
		elem, ok := s.(string)
		if !ok {
			panic("cannot join a slice that contains something else than strings")
		}
		b.WriteString(elem)
	}
	return b.String()
}
