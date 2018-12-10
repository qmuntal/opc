package gopc

import "errors"

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
	uri           string
	relationships []*Relationship
}

var (
	// ErrDuplicatedRelationship throw error for invalid relationship.
	ErrDuplicatedRelationship = errors.New("a relationship is duplicated")
)

// NewPart creates a new part.
func newPart(uri string) (*Part, error) {
	if len(uri) == 0 {
		return nil, ErrInvalidTargetURI
	}
	return &Part{uri: uri}, nil
}

//func newPart(uri, contentType string, compressionOption CompressionOption) {}

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
