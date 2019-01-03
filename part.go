package gopc

import (
	"errors"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
)

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
func NormalizePartName(name string) (string, error) {
	if str := strings.TrimSpace(name); str == "" || str == "/" {
		return "", errors.New("OPC: a part URI shall not be empty")
	}

	normalized := strings.Replace(filepath.ToSlash(name), "//", "/", -1)
	if strings.HasSuffix(normalized, "/") {
		normalized = normalized[:len(normalized)-1]
	}

	encodedURL, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}

	if encodedURL.IsAbs() {
		return "", errors.New("OPC: a part URI shall be relative")
	}

	// Normalize url, decode unnecessary escapes and encode necessary
	p, _ := url.Parse(defaultRef.ResolveReference(encodedURL).Path)
	return p.EscapedPath(), nil
}

func validatePartName(name string) error {
	// ISO/IEC 29500-2 M1.1
	if strings.TrimSpace(name) == "" {
		return errors.New("OPC: a part URI shall not be empty")
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
		return errors.New("OPC: a part URI shall start with a forward slash character")
	}

	// ISO/IEC 29500-2 M1.6
	if name != encodedURL.EscapedPath() {
		return errors.New("OPC: segment shall not hold any characters other than pchar characters")
	}
	return nil
}

func validateChars(name string) error {
	// ISO/IEC 29500-2 M1.5
	if strings.HasSuffix(name, "/") {
		return errors.New("OPC: a part URI shall not have a forward slash as the last character")
	}

	// ISO/IEC 29500-2 M1.9
	if strings.HasSuffix(name, ".") {
		return errors.New("OPC: a segment shall not end with a dot character")
	}

	// ISO/IEC 29500-2 M1.3
	if strings.Contains(name, "//") {
		return errors.New("OPC: a part URI shall not have empty segments")
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

// A part is a stream of bytes whose main properties are the name and the content type.
// Parts are analogous to a file in a file system or to a resource on an HTTP server.
// Defined in ISO/IEC 29500-2 §9.1.
type Part struct {
	relationable
	uri               string
	contentType       string
	compressionOption CompressionOption
}

// newPart creates a new part.
func newPart(uri, contentType string, compressionOption CompressionOption) (*Part, error) {
	if err := validatePartName(uri); err != nil {
		return nil, err
	}

	if !strings.Contains(contentType, "/") {
		return nil, errors.New("OPC: expected slash in content type")
	}

	mediatype, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	return &Part{
		relationable:      relationable{uri, make(map[string]*Relationship, 0)},
		uri:               uri,
		contentType:       mime.FormatMediaType(mediatype, params),
		compressionOption: compressionOption}, nil
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
