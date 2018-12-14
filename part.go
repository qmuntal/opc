package gopc

import (
	"errors"
	"mime"
	"strings"
)

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

var (
	// ErrDuplicatedRelationship throw error for invalid relationship.
	ErrDuplicatedRelationship = errors.New("a relationship is duplicated")
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
	if len(uri) == 0 {
		return nil, ErrInvalidTargetURI
	}

	if !strings.Contains(contentType, "/") {
		return nil, errors.New("mime: expected slash in content type")
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
			return nil, ErrDuplicatedRelationship
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
