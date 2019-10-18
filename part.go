package opc

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
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

var defaultRef, _ = url.Parse("http://defaultcontainer/")

// ResolveRelationship returns the absolute URI (from the package root) of the part pointed by a relationship of a source part.
// This method should be used in places where we have a target relationship URI and we want to get the
// name of the part it targets with respect to the source part.
// The source can be a valid part URI, for part relationships, or "/", for package relationships.
func ResolveRelationship(source string, rel string) string {
	if source == "/" || source == "\\" {
		return "/" + rel
	}
	if !strings.HasPrefix(rel, "/") && !strings.HasPrefix(rel, "\\") {
		sourceDir := strings.Replace(filepath.Dir(source), "\\", "/", -1)
		return fmt.Sprintf("%s/%s", sourceDir, rel)
	}
	return rel
}

// NormalizePartName transforms the input name so it follows the constrains specified in the ISO/IEC 29500-2 §9.1.1:
//     part-URI = 1*( "/" segment )
//     segment = 1*( pchar )
// pchar is defined in RFC 3986:
//     pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//     unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
//     pct-encoded = "%" HEXDIG HEXDIG
//     sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
// This method is recommended to be used before adding a new Part to a package to avoid errors.
// If, for whatever reason, the name can't be adapted to the specs, the return value will be the same as the original.
// Warning: This method can heavily modify the original if it differs a lot from the specs, which could led to duplicated part names.
func NormalizePartName(name string) string {
	if str := strings.TrimSpace(name); str == "" || str == "/" {
		return name
	}

	normalized := strings.Replace(name, "\\", "/", -1)
	normalized = strings.Replace(normalized, "//", "/", -1)
	normalized = strings.Replace(normalized, "%2e", ".", -1)
	if strings.HasSuffix(normalized, "/") {
		normalized = normalized[:len(normalized)-1]
	}

	encodedURL, err := url.Parse(normalized)
	if err != nil || encodedURL.IsAbs() {
		return name
	}

	// Normalize url, decode unnecessary escapes and encode necessary
	p, err := url.Parse(defaultRef.ResolveReference(encodedURL).Path)
	if err != nil {
		return name
	}
	return p.EscapedPath()
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
	if strings.TrimSpace(name) == "" {
		return newError(101, name)
	}

	if err := validateChars(name); err != nil {
		return err
	}

	if err := validateSegments(name); err != nil {
		return err
	}

	return validateURL(name)
}

func validateURL(name string) error {
	encodedURL, err := url.Parse(name)
	if err != nil {
		return fmt.Errorf("opc: %s: invalid url: %w", name, err)
	}

	if name[0] != '/' || encodedURL.IsAbs() {
		return newError(104, name)
	}

	if name != encodedURL.EscapedPath() {
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
