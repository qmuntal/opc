package gopc

import (
	"errors"
	"mime"
	"net/url"
	"strings"
)

// ValidatePartName checks that the part name is follows the constrains specified in the ISO/IEC 29500-2 Section 9.1.1.1.2
// A part name is the name of a part within a package encoded
// as a URI per ISO/IEC 29500-2 Section 9.1.1:
//     part-URI = 1*( "/" segment )
//     segment = 1*( pchar )
// pchar is defined in RFC 3986:
//     pchar = unreserved / pct-encoded / sub-delims / ":" / "@"
//     unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
//     pct-encoded = "%" HEXDIG HEXDIG
//     sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
func ValidatePartName(name string) error {
	if len(name) == 0 {
		return errors.New("OPC: a part URI shall not be empty [ISO/IEC 29500-2 M1.1]")
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

	if name[0] != '/' || encodedURL.IsAbs() {
		return errors.New("OPC: a part URI shall start with a forward slash character [ISO/IEC 29500-2 M1.4]")
	}

	if name != encodedURL.EscapedPath() {
		return errors.New("OPC: segment shall not hold any characters other than pchar characters [ISO/IEC 29500-2 M1.6]")
	}
	return nil
}

func validateChars(name string) error {
	if strings.HasSuffix(name, "/") {
		return errors.New("OPC: a part URI shall not have a forward slash as the last character [ISO/IEC 29500-2 M1.5]")
	}

	if strings.HasSuffix(name, ".") {
		return errors.New("OPC: a segment shall not end with a dot character [ISO/IEC 29500-2 M1.9]")
	}

	if strings.Contains(name, "//") {
		return errors.New("OPC: a part URI shall not have empty segments [ISO/IEC 29500-2 M1.3]")
	}
	return nil
}

func validateSegments(name string) error {
	if strings.Contains(name, "/./") || strings.Contains(name, "/../") {
		return errors.New("OPC: a segment shall include at least one non-dot character [ISO/IEC 29500-2 M1.10]")
	}

	u := strings.ToUpper(name)
	// "/" "\"
	if strings.Contains(u, "%5C") || strings.Contains(u, "%2F") {
		return errors.New("OPC: a segment shall not contain percent-encoded forward slash or backward slash characters [ISO/IEC 29500-2 M1.7]")
	}

	// "-" "." "_" "~"
	if strings.Contains(u, "%2D") || strings.Contains(u, "%2E") || strings.Contains(u, "%5F") || strings.Contains(u, "%7E") {
		return errors.New("OPC: a segment shall not contain percent-encoded unreserved characters [ISO/IEC 29500-2 M1.8]")
	}
	return nil
}

// CompressionOption is an enumerable for the different compression options.
type CompressionOption int

const (
	// CompressionNone disables the compression.
	CompressionNone CompressionOption = iota - 1
	// CompressionNormal is optimized for a reasonable compromise between size and performance.
	CompressionNormal
	// CompressionMaximum is optimized for size.
	CompressionMaximum
	// CompressionFast is optimized for performance.
	CompressionFast
	// CompressionSuperFast is optimized for super performance.
	CompressionSuperFast
)

// Part defines an OPC Package Object.
type Part struct {
	uri               string
	contentType       string
	compressionOption CompressionOption
	relationships     []*Relationship
}

// newPart creates a new part with no relationships.
func newPart(uri, contentType string, compressionOption CompressionOption) (*Part, error) {
	if err := ValidatePartName(uri); err != nil {
		return nil, err
	}

	if !strings.Contains(contentType, "/") {
		return nil, errors.New("OPC: expected slash in content type")
	}

	mediatype, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	return &Part{uri: uri, contentType: mime.FormatMediaType(mediatype, params), compressionOption: compressionOption}, err
}

// AddRelationship add a relationship to the part.
func (p *Part) AddRelationship(id, reltype, uri string) (*Part, error) {
	r, err := newRelationship(id, reltype, uri, ModeInternal)

	for i := 0; i < len(p.relationships); i++ {
		if p.relationships[i].ID() == id && p.relationships[i].Type() == reltype {
			return nil, errors.New("OPC: trying to add a duplicated relationship")
		}
	}
	p.relationships = append(p.relationships, r)
	return p, err
}

// HasRelationship return true if the part have relationships
func (p *Part) HasRelationship() bool {
	return len(p.relationships) > 0
}

// Relationships return all the relationships of the part
func (p *Part) Relationships() []*Relationship {
	return p.relationships
}

// URI returns the URI of the part.
func (p *Part) URI() string {
	return p.uri
}

// ContentType returns the ContentType of the part.
func (p *Part) ContentType() string {
	return p.contentType
}

// CompressionOption returns the CompressionOption of the part.
func (p *Part) CompressionOption() CompressionOption {
	return p.compressionOption
}
