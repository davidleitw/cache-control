package cache

import (
	"errors"
	"strings"
)

var (
	ErrMissingClosingQuote = errors.New("missing closing quote")
)

// parseCacheControlv is a function that parses a Cache-Control header directive value
// and sets the corresponding directive in the given directive object.
//
// The function iterates through the string value of the directive, skipping any leading whitespace or commas.
// It then finds the end of the current token (a sequence of characters that conform to the token ABNF definition),
// converts the token to lowercase, and determines whether an extension field is required for the token.
//
// If the current character is an equals sign, indicating that the token is a key-value pair,
// the function checks whether the value is enclosed in quotes.
// If it is, it uses parseQuotedString to extract the value, updates the index to skip over the value,
// and sets the directive pair with the token and value.
// If the value is not enclosed in quotes, the function finds the end of the value and sets the directive pair with the token and value.
//
// If the current character is not an equals sign, indicating that the token is a single value,
// the function sets the directive token with the current token.
//
// The function returns an error if an unexpected character is encountered or if a quoted string is not closed.
func parseCacheControlv(d directive, val string) error {
	var (
		index = 0
		vl    = len(val)
	)

	for index < vl {
		if isWhiteSpace(val[index]) || val[index] == ',' {
			index++
			continue
		}

		// Find the end of the token
		tokenEnd := index + 1
		for tokenEnd < vl {
			if !isToken(val[tokenEnd]) {
				break
			}
			tokenEnd++
		}

		// Get the lowercase token string and check if it requires an extension field
		token := strings.ToLower(val[index:tokenEnd])
		requireExtensionField := tokenRequireExtensionFields(token)

		// If the token has an equals sign, it's a pair
		if tokenEnd+1 < vl && val[tokenEnd] == '=' {
			valueStart := tokenEnd + 1

			// If the value is quoted, parse the quoted string
			if valueStart < vl && val[valueStart] == '"' {
				eaten, value := parseQuotedString(val[valueStart:])
				if eaten == -1 {
					return ErrMissingClosingQuote
				}

				index = valueStart + eaten
				if err := d.setPair(token, value); err != nil {
					return err
				}
			} else {
				// If the value is not quoted, find the end of the pair value
				valueEnd := valueStart
				for valueEnd < vl {
					if isWhiteSpace(val[valueEnd]) ||
						(!requireExtensionField && val[valueEnd] == ',') {
						break
					}
					valueEnd++
				}
				index = valueEnd

				// Remove the trailing comma if there is one and update the directive with the pair
				value := val[valueStart:valueEnd]
				if value != "" && value[len(value)-1] == ',' {
					value = value[:len(value)-1]
				}
				if err := d.setPair(token, value); err != nil {
					return err
				}
			}
		} else {
			// If the token doesn't have an equals sign, it's a simple token
			if token != "," {
				if err := d.setToken(token); err != nil {
					return err
				}
			}
			index = tokenEnd
		}
	}

	return nil
}

func tokenRequireExtensionFields(token string) bool {
	switch token {
	case "no-cache", "private":
		return true
	default:
		return false
	}
}

func parseQuotedString(raw string) (int, string) {
	if raw[0] != '"' {
		return -1, ""
	}

	var (
		rl     = len(raw)
		buf    = make([]byte, rl)
		eat    = 1
		bufIdx = 0
	)

	for i := 1; i < rl; i++ {
		switch b := raw[i]; b {
		case '"':
			eat++
			buf = buf[0:bufIdx]
			return i + 1, string(buf)
		case '\\':
			if len(raw) < i+2 {
				return -1, ""
			}

			buf[bufIdx] = unquotePair(raw[i+1])
			i++
			bufIdx++
			eat += 2
		default:
			if isQdText(b) {
				buf[bufIdx] = b
			} else {
				buf[bufIdx] = '?'
			}
			eat++
			bufIdx++
		}
	}
	return -1, ""
}

func unquotePair(b byte) byte {
	switch b {
	case 'a':
		return '\a'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	case '\\':
		return '\\'
	case '\'':
		return '\''
	case '"':
		return '"'
	}
	return '?'
}
