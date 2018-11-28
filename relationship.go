package gopc

import (
	"encoding/xml"
	"errors"
)

const relationshipName = "Relationship"

// Relationship defines an OPC Package Relationship Object.
type Relationship struct {
	id            string
	relType       string
	targetPartURI string
}

type relationshipXML struct {
	ID            string `xml:"Id,attr"`
	RelType       string `xml:"Type,attr"`
	TargetPartURI string `xml:"Target,attr"`
}

var (
	// ErrInvalidOPCPartURI throw error for invalid URI.
	ErrInvalidOPCPartURI = errors.New("Invalid OPC Part URI")
)

// NewRelationship creates a new relationship.
func NewRelationship(id, relType, targetPartURI string) (*Relationship, error) {
	if len(targetPartURI) == 0 || targetPartURI[0] == '.' {
		return nil, ErrInvalidOPCPartURI
	}

	return &Relationship{id: id, relType: relType, targetPartURI: targetPartURI}, nil
}

// ID returns the ID of the relationship.
func (r *Relationship) ID() string {
	return r.id
}

// Type returns the type of the relationship.
func (r *Relationship) Type() string {
	return r.relType
}

// TargetPartURI returns the targetpartURI of the relationship.
func (r *Relationship) TargetPartURI() string {
	return r.targetPartURI
}

func (r *Relationship) toXML() *relationshipXML {
	x := &relationshipXML{ID: r.id, RelType: r.relType, TargetPartURI: r.targetPartURI}
	if len(x.TargetPartURI) == 0 {
		x.TargetPartURI = "/"
		return x
	}

	if x.TargetPartURI[0] != '/' && x.TargetPartURI[0] != '\\' {
		x.TargetPartURI = "/" + x.TargetPartURI
	}
	return x
}

// WriteToXML encodes the relationship to the target.
func (r *Relationship) WriteToXML(e *xml.Encoder) error {
	return e.EncodeElement(r.toXML(), xml.StartElement{Name: xml.Name{Space: "", Local: relationshipName}})
}
