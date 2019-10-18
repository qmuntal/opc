package opc

import (
	"fmt"
)

var errorsString = map[int]string{
	101: "a part name shall not be empty",
	102: "a part content type shall not be empty",
	103: "a part name shall not have empty segments",
	104: "a part name shall start with a forward slash character",
	105: "a part name shall not have a forward slash as the last character",
	106: "a part name segment shall not hold any characters other than pchar characters",
	107: "a part name segment shall not contain percent-encoded forward slash or backward slash characters",
	108: "a part name segment shall not contain percent-encoded unreserved characters",
	109: "a part name segment shall not end with a dot character",
	110: "a part name segment shall include at least one non-dot character",
	111: "a package shall not contain a part with a part name derived from another part name by appending segments to it",
	112: "a package shall not contain equivalent part names",
	113: "a part content type shall fit the definition and syntax for media types as specified in RFC 2616 §3.7",
	114: "a part content type shall not have linear, leading or trailing white space",
	125: "a relationship shall not have relationships to any other part",
	126: "a relationship identifier cannot be empty and shall be unique within the relationships part",
	127: "a relationship type cannot be empty",
	128: "a relationship target URI reference shall be a URI or a relative reference",
	129: "a relationship target URI must be relative if the TargetMode is Internal",
	205: "a Default content type shall not have more than one content type for each extension and a Override shall not have more than one content type for each PartName",
	206: "a package shall not have an empty extension in a Default element",
	208: "a part content type shall appear in [Content_Types].xml",
	310: "a package shall contain a file named [Content_Types].xml to store all the data content types",
}

// An Error from this package is always associated to an OPC entity that is not conformant with the OPC specs.
type Error struct {
	code     int
	partName string
	relID    string
}

func newError(code int, partName string) *Error {
	return &Error{code, partName, ""}
}

func newErrorRelationship(code int, partName, relID string) *Error {
	return &Error{code, partName, relID}
}

// Code of the error as described in the OPC specs.
// The first number is the top level topic and the second and third digits are the specific error code.
// The top level topics are described as follows:
// 1. Package Model requirements
// 2. Physical Packages requirements
// 3. ZIP Physical Mapping requirements
// 4. Core Properties requirements
// 5. Thumbnail requirements
// 6. Digital Signatures requirements
// 7. Pack URI requirements
func (e *Error) Code() int {
	return e.code
}

// PartName returns the name of the Part associated to the error.
func (e *Error) PartName() string {
	return e.partName
}

// RelationshipID returns the ID of the relationship associated to the error.
// If the error is not associated to a relationship, the value is empty.
func (e *Error) RelationshipID() string {
	return e.relID
}

func (e *Error) Error() string {
	s, ok := errorsString[e.code]
	if !ok {
		panic("opc: undefined error")
	}
	return fmt.Sprintf("opc: %s: %s", e.partName, s)
}
