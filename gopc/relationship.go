package gopc

// Relationship defines an OPC Package Relationship Object.
type Relationship struct {
}

// NewRelationship creates a new relationship.
func NewRelationship() *Relationship {
	return new(Relationship)
}
