package gopc

import "errors"

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
		return nil, ErrInvalidOPCPartURI
	}
	return &Part{uri: uri}, nil
}

// AddRelationship add a relationship to the part.
func (p *Part) AddRelationship(id, reltype, uri string) error {
	r, err := NewRelationship(id, reltype, uri)

	for i := 0; i < len(p.relationships); i++ {
		if p.relationships[i].ID() == id && p.relationships[i].Type() == reltype {
			return ErrDuplicatedRelationship
		}
	}
	p.relationships = append(p.relationships, r)
	return err
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
