package stringutils

// Remove \ that are not escaped
func Unescape(line string) string {
	var res string
	lastEscapeEnd := 0
	for i := range line {
		if line[i] == '\\' {
			if i+1 < len(line) && line[i+1] == '\\' {
				continue
			}
			if i > 1 && line[i-1] == '\\' {
				res += line[lastEscapeEnd:i]
				lastEscapeEnd = i + 1
				continue
			}
			res += line[lastEscapeEnd:i]
			lastEscapeEnd = i + 1
			continue
		}
		if i == len(line)-1 {
			res += line[lastEscapeEnd:]
		}
	}

	return res
}
