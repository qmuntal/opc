package gopc

import (
	"errors"
	"mime"
	"net/url"
	"strings"
)

// A Part is a stream of bytes defined in ISO/IEC 29500-2 §9.1..
// Parts are analogous to a file in a file system or to a resource on an HTTP server.
// The part properties will be validated before writing or reading from disk.
type Part struct {
	Name          string          // The name of the part.
	ContentType   string          // The type of content stored in the part.
	Relationships []*Relationship // The relationships associated to the part.
}

// CreateRelationship adds a new relationship to the Relationships slice.
// The ID can be an empty string, if so a unique ID will be generated.
// The input properties are not validated.
func (p *Part) CreateRelationship(id, targetURI, relType string, targetMode TargetMode) *Relationship {
	if id == "" {
		id = uniqueRelationshipID()
	}
	r := &Relationship{
		ID:         id,
		RelType:    relType,
		TargetURI:  targetURI,
		TargetMode: targetMode,
		sourceURI:  p.Name,
	}
	p.Relationships = append(p.Relationships, r)
	return r
}

func (p *Part) validate() error {
	if err := validatePartName(p.Name); err != nil {
		return err
	}

	return validateContentType(p.ContentType)
}

var defaultRef, _ = url.Parse("http://defaultcontainer/")

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

func validateContentType(contentType string) error {
	if len(contentType) == 0 {
		return nil
	}

	// ISO/IEC 29500-2 M1.14
	if contentType[0] == ' ' || strings.HasSuffix(contentType, " ") {
		return errors.New("OPC: a content type also shall not have linear, leading or trailing white space")
	}

	// ISO/IEC 29500-2 M1.13 and M1.14
	t, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}

	// mime package accepts Content-Disposition, which does not start with slash
	if !strings.Contains(t, "/") {
		return errors.New("OPC: expected slash in content type")
	}

	return nil
}

func validatePartName(name string) error {
	// ISO/IEC 29500-2 M1.1
	if strings.TrimSpace(name) == "" {
		return errors.New("OPC: a part name shall not be empty")
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
		return err
	}

	// ISO/IEC 29500-2 M1.4
	if name[0] != '/' || encodedURL.IsAbs() {
		return errors.New("OPC: a part name shall start with a forward slash character")
	}

	// ISO/IEC 29500-2 M1.6
	if name != encodedURL.EscapedPath() {
		return errors.New("OPC: a segment shall not hold any characters other than pchar characters")
	}
	return nil
}

func validateChars(name string) error {
	// ISO/IEC 29500-2 M1.5
	if strings.HasSuffix(name, "/") {
		return errors.New("OPC: a part name shall not have a forward slash as the last character")
	}

	// ISO/IEC 29500-2 M1.9
	if strings.HasSuffix(name, ".") {
		return errors.New("OPC: a segment shall not end with a dot character")
	}

	// ISO/IEC 29500-2 M1.3
	if strings.Contains(name, "//") {
		return errors.New("OPC: a part name shall not have empty segments")
	}
	return nil
}

func validateSegments(name string) error {
	// ISO/IEC 29500-2 M1.10
	if strings.Contains(name, "/./") || strings.Contains(name, "/../") {
		return errors.New("OPC: a segment shall include at least one non-dot character")
	}

	u := strings.ToUpper(name)
	// ISO/IEC 29500-2 M1.7
	// "/" "\"
	if strings.Contains(u, "%5C") || strings.Contains(u, "%2F") {
		return errors.New("OPC: a segment shall not contain percent-encoded forward slash or backward slash characters")
	}

	// ISO/IEC 29500-2 M1.8
	// "-" "." "_" "~"
	if strings.Contains(u, "%2D") || strings.Contains(u, "%2E") || strings.Contains(u, "%5F") || strings.Contains(u, "%7E") {
		return errors.New("OPC: a segment shall not contain percent-encoded unreserved characters")
	}
	return nil
}
