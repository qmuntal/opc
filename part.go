package opc

import (
	"fmt"
	"mime"
	"path"
	"strings"
	"unicode/utf8"
)

// A Part is a stream of bytes defined in ISO/IEC 29500-2 §9.1..
// Parts are analogous to a file in a file system or to a resource on an HTTP server.
type Part struct {
	Name          string          // The name of the part.
	ContentType   string          // The type of content stored in the part.
	Relationships []*Relationship // The relationships associated to the part. Can be modified until the Writer is closed.
}

func (p *Part) validate() error {
	if err := validatePartName(p.Name); err != nil {
		return err
	}

	if err := p.validateContentType(); err != nil {
		return err
	}

	return validateRelationships(p.Name, p.Relationships)
}

// ResolveRelationship returns the absolute URI (from the package root) of the part pointed by a relationship of a source part.
// This method should be used in places where we have a target relationship URI and we want to get the
// name of the part it targets with respect to the source part.
// The source can be a valid part URI, for part relationships, or "/", for package relationships.
func ResolveRelationship(source string, rel string) string {
	source = strings.Replace(source, "\\", "/", -1)
	rel = strings.Replace(rel, "\\", "/", -1)
	if source == "/" && !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	if !strings.HasPrefix(rel, "/") {
		rel = fmt.Sprintf("%s/%s", path.Dir(source), rel)
	}
	return rel
}

// NormalizePartName transforms the input name as an URI string
// so it follows the constrains specified in the ISO/IEC 29500-2 §9.1.1.
// This method is recommended to be used before adding a new Part to a package to avoid errors.
// If, for whatever reason, the name can't be adapted to the specs, the return value is empty.
// Warning: This method can heavily modify the name if it differs a lot from the specs, which could led to duplicated part names.
func NormalizePartName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "/" || name == "\\" || name == "." {
		return ""
	}
	name, _ = split(name, '#')
	name = strings.NewReplacer("\\", "/", "//", "/").Replace(name)
	name = unescape(name)
	name = escape(name)
	name = cleanSegments(name)
	return strings.TrimSuffix(name, "/")
}

func (p *Part) validateContentType() error {
	if strings.TrimSpace(p.ContentType) == "" {
		return newError(102, p.Name)
	}

	if p.ContentType[0] == ' ' || strings.HasSuffix(p.ContentType, " ") {
		return newError(114, p.Name)
	}

	// mime package accepts Content-Disposition, which does not start with slash
	if t, _, err := mime.ParseMediaType(p.ContentType); err != nil || !strings.Contains(t, "/") {
		return newError(113, p.Name)
	}

	return nil
}

func validatePartName(name string) error {
	if strings.EqualFold(name, contentTypesName) {
		return nil
	}
	if strings.TrimSpace(name) == "" {
		return newError(101, name)
	}

	if name[0] != '/' {
		return newError(104, name)
	}

	if err := validateChars(name); err != nil {
		return err
	}

	if err := validateSegments(name); err != nil {
		return err
	}

	if !validEncoded(name) {
		return newError(106, name)
	}
	return nil
}

func validateChars(name string) error {
	if strings.HasSuffix(name, "/") {
		return newError(105, name)
	}

	if strings.HasSuffix(name, ".") {
		return newError(109, name)
	}

	if strings.Contains(name, "//") {
		return newError(103, name)
	}
	return nil
}

func validateSegments(name string) error {
	if strings.Contains(name, "/./") || strings.Contains(name, "/../") {
		return newError(110, name)
	}

	u := strings.ToUpper(name)
	// "/" "\"
	if strings.Contains(u, "%5C") || strings.Contains(u, "%2F") {
		return newError(107, name)
	}

	// "-" "." "_" "~"
	if strings.Contains(u, "%2D") || strings.Contains(u, "%2E") || strings.Contains(u, "%5F") || strings.Contains(u, "%7E") {
		return newError(108, name)
	}
	return nil
}

func cleanSegments(s string) string {
	src := strings.Split(s, "/")
	dst := make([]string, 0, len(src))
	for _, elem := range src {
		switch elem {
		case ".", "..":
			// drop
		default:
			dst = append(dst, elem)
		}
	}
	if last := src[len(src)-1]; last == "." || last == ".." {
		// Add final slash to the joined path.
		dst = append(dst, "")
	}
	return "/" + strings.TrimPrefix(strings.Join(dst, "/"), "/")
}

func escape(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				hexCount++
			}
		default:
			if shouldEscape(s[i]) {
				hexCount++
			}
		}
	}
	if hexCount == 0 {
		return s
	}
	var buf [64]byte
	var t []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		t = buf[:required]
	} else {
		t = make([]byte, required)
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				t[j] = '%'
				t[j+1] = '2'
				t[j+2] = '5'
				j += 3
			} else {
				t[j], t[j+1], t[j+3] = '%', s[i+1], s[i+2]
				j += 3
			}
		default:
			c := s[i]
			if shouldEscape(c) {
				t[j] = '%'
				t[j+1] = upperhex[c>>4]
				t[j+2] = upperhex[c&15]
				j += 3
			} else {
				t[j] = s[i]
				j++
			}
		}
	}
	return string(t)
}

func unescape(s string) string {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == '%' {
			if i+2 < len(s) && ishex(s[i+1]) && ishex(s[i+2]) {
				c := unpct(s[i+1], s[i+2])
				if c == '%' || isReserved(c) {
					i++
				} else {
					n++
					i += 3
				}
			} else {
				i++
			}
		} else {
			i++
		}
	}

	if n == 0 {
		return s
	}

	var t strings.Builder
	t.Grow(len(s) - 2*n)
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+2 < len(s) && ishex(s[i+1]) && ishex(s[i+2]) {
				c := unpct(s[i+1], s[i+2])
				if c == '%' || isReserved(c) {
					t.WriteByte(s[i])
				} else {
					t.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
					i += 2
				}
			} else {
				t.WriteByte(s[i])
			}
		} else {
			t.WriteByte(s[i])
		}
	}
	return t.String()
}

func split(s string, sep byte) (string, string) {
	i := strings.IndexByte(s, sep)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i:]
}

const upperhex = "0123456789ABCDEF"

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func isAlpha(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isUnreserved(c byte) bool {
	return isAlpha(c) || isDigit(c) || c == '-' || c == '.' || c == '_' || c == '~'
}

func isReserved(c byte) bool {
	if c == '/' || c == ':' || c == '@' {
		return true
	}
	if c == '!' || c == '$' || c == '&' || c == '\'' || c == '(' || c == ')' ||
		c == '*' || c == '+' || c == ',' || c == ';' || c == '=' {
		return true
	}
	return false
}

func isUcsChar(r rune) bool {
	return 0xA0 <= r && r <= 0xD7FF || 0xF900 <= r && r <= 0xFDCF || 0xFDF0 <= r && r <= 0xFFEF ||
		0x10000 <= r && r <= 0x1FFFD || 0x20000 <= r && r <= 0x2FFFD || 0x30000 <= r && r <= 0x3FFFD ||
		0x40000 <= r && r <= 0x4FFFD || 0x50000 <= r && r <= 0x5FFFD || 0x60000 <= r && r <= 0x6FFFD ||
		0x70000 <= r && r <= 0x7FFFD || 0x80000 <= r && r <= 0x8FFFD || 0x90000 <= r && r <= 0x9FFFD ||
		0xA0000 <= r && r <= 0xAFFFD || 0xB0000 <= r && r <= 0xBFFFD || 0xC0000 <= r && r <= 0xCFFFD ||
		0xD0000 <= r && r <= 0xDFFFD || 0xE1000 <= r && r <= 0xEFFFD
}

func shouldEscape(c byte) bool {
	return !isUnreserved(c) && !isReserved(c)
}

func unpct(c1, c2 byte) byte {
	return unhex(c1)<<4 | unhex(c2)
}

func validEncoded(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			if i+2 < len(s) && isUnreserved(unpct(s[i+1], s[i+2])) {
				return false
			}
			// ok
		default:
			if shouldEscape(s[i]) {
				// Check if IRI supported shar
				r, wid := utf8.DecodeRuneInString(s[i:])
				if !isUcsChar(r) {
					return false
				}
				i += wid
			}
		}
	}
	return true
}
