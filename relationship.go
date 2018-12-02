package gopc

import (
	"encoding/xml"
	"errors"
	"net/url"
	"strings"
)

// TargetMode is an enumerable for the different target modes.
type TargetMode int

const (
	// ModeInternal when the target targetMode is Internal (default value).
	// Target points to a part within the package and target uri must be relative.
	ModeInternal TargetMode = iota
	// ModeExternal when the target targetMode is External.
	// Target points to an external resource and target uri can be relative or absolute.
	ModeExternal
)

const relationshipName = "Relationship"
const externalMode = "External"

// Relationship is used to express a relationship between a source and a target part.
// The only way to create a Relationship, is to call the Part.NewRelationship()
// or Package.NewRelationship(). A relationship is owned by a part or by the package itself.
// If the source part is deleted all the relationships it owns are also deleted.
// A target of the relationship need not be present.
type Relationship struct {
	id         string
	relType    string
	targetURI  string
	targetMode TargetMode
}

type relationshipXML struct {
	ID        string `xml:"Id,attr"`
	RelType   string `xml:"Type,attr"`
	TargetURI string `xml:"Target,attr"`
	Mode      string `xml:"TargetMode,attr,omitempty"`
}

var (
	// ErrInvalidTargetURI happens when the target uri is invalid.
	ErrInvalidTargetURI = errors.New("OPC: invalid target URI")
	// ErrInvalidRelType happens when a relation type is empty.
	ErrInvalidRelType = errors.New("OPC: relationship type cannot be empty string or a string with just spaces")
	// ErrRelationshipInternalAbs happens when a relationship is internal and has an absolute targetURI.
	ErrRelationshipInternalAbs = errors.New("OPC: relationship target must be relative if the TargetMode is Internal")
)

var (
	// RelTypeMetaDataCoreProps defines a core properties relationship.
	RelTypeMetaDataCoreProps = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	// RelTypeDigitalSignature defines a digital signature relationship.
	RelTypeDigitalSignature = "http://schemas.openxmlformats.org/package/2006/relationships/digital-signature/signature"
	// RelTypeDigitalSignatureOrigin defines a digital signature origin relationship.
	RelTypeDigitalSignatureOrigin = "http://schemas.openxmlformats.org/package/2006/relationships/digital-signature/origin"
	// RelTypeDigitalSignatureCert defines a digital signature certificate relationship.
	RelTypeDigitalSignatureCert = "http://schemas.openxmlformats.org/package/2006/relationships/digital-signature/certificate"
	// RelTypeThumbnail defines a thmbnail relationship.
	RelTypeThumbnail = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"
)

// NewRelationship creates a new internal relationship.
func NewRelationship(id, relType, targetURI string) (*Relationship, error) {
	return NewRelationshipMode(id, relType, targetURI, ModeInternal)
}

// NewRelationshipMode creates a new relationship.
func NewRelationshipMode(id, relType, targetURI string, targetMode TargetMode) (*Relationship, error) {
	if strings.TrimSpace(targetURI) == "" {
		return nil, ErrInvalidTargetURI
	}

	if strings.TrimSpace(relType) == "" {
		return nil, ErrInvalidRelType
	}

	uri, err := url.Parse(targetURI)
	if err != nil {
		return nil, ErrInvalidTargetURI
	}

	if targetMode == ModeInternal && uri.IsAbs() {
		return nil, ErrRelationshipInternalAbs
	}

	return &Relationship{id: id, relType: relType, targetURI: targetURI, targetMode: targetMode}, nil
}

// ID returns the ID of the relationship.
func (r *Relationship) ID() string {
	return r.id
}

// Type returns the type of the relationship.
func (r *Relationship) Type() string {
	return r.relType
}

// targetURI returns the targetURI of the relationship.
func (r *Relationship) TargetURI() string {
	return r.targetURI
}

func (r *Relationship) toXML() *relationshipXML {
	var targetMode string
	if r.targetMode == ModeExternal {
		targetMode = externalMode
	}
	x := &relationshipXML{ID: r.id, RelType: r.relType, TargetURI: r.targetURI, Mode: targetMode}
	if r.targetMode == ModeInternal {
		if !strings.HasPrefix(x.TargetURI, "/") && !strings.HasPrefix(x.TargetURI, "\\") && !strings.HasPrefix(x.TargetURI, ".") {
			x.TargetURI = "/" + x.TargetURI
		}
	}
	return x
}

// WriteToXML encodes the relationship to the target.
func (r *Relationship) WriteToXML(e *xml.Encoder) error {
	return e.EncodeElement(r.toXML(), xml.StartElement{Name: xml.Name{Space: "", Local: relationshipName}})
}
