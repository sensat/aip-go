package fieldmask

// splits a field mask path into its individual fields.
// supports backticks to escape dots. these can be used to
// address fields with dots in them, such as keys in maps.
// a_map.`entry.key`.name would address a_map["entry.key"].name
func SplitPath(s string) []string {
	tokens := []string{}
	currentToken := ""
	insideBackticks := false

	for _, ch := range s {
		if ch == '`' {
			insideBackticks = !insideBackticks
		} else if ch == '.' && !insideBackticks {
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
		} else {
			currentToken += string(ch)
		}
	}

	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}

	return tokens
}
