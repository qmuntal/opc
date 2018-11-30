package gopc

// Part defines an OPC Package Object.
type Part struct {
	uri           string
	relationships []Relationship
}
