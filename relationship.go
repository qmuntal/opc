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
// Defined in ISO/IEC 29500-2 §9.3.
type Relationship struct {
	id         string
	relType    string
	sourceURI  string
	targetURI  string
	targetMode TargetMode
}

type relationshipXML struct {
	ID        string `xml:"Id,attr"`
	RelType   string `xml:"Type,attr"`
	TargetURI string `xml:"Target,attr"`
	Mode      string `xml:"TargetMode,attr,omitempty"`
}

const (
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

// isRelationshipURI returns true if the uri points to a relationship part.
func isRelationshipURI(uri string) bool {
	up := strings.ToUpper(uri)
	if !strings.HasSuffix(up, ".RELS") {
		return false
	}

	if strings.EqualFold(up, "/_RELS/.RELS") {
		return true
	}

	eq := false
	// Look for pattern that matches: "XXX/_rels/YYY.rels" where XXX is zero or more part name characters and
	// YYY is any legal part name characters
	segments := strings.Split(up, "/")
	ls := len(segments)
	if ls >= 3 && len(segments[ls-1]) > len(".RELS") {
		eq = strings.EqualFold(segments[ls-2], "_RELS")
	}
	return eq
}

// validateRelationshipTarget checks that a relationship target follows the constrains specified in the ISO/IEC 29500-2 §9.3.
func validateRelationshipTarget(sourceURI, targetURI string, targetMode TargetMode) error {
	// ISO/IEC 29500-2 M1.28
	uri, err := url.Parse(strings.TrimSpace(targetURI))
	if err != nil || uri.String() == "" {
		return errors.New("OPC: relationship target URI reference shall be a URI or a relative reference")
	}

	// ISO/IEC 29500-2 M1.29
	if targetMode == ModeInternal && uri.IsAbs() {
		return errors.New("OPC: relationship target URI must be relative if the TargetMode is Internal")
	}

	var result error
	if targetMode != ModeExternal && !uri.IsAbs() {
		source, err := url.Parse(strings.TrimSpace(sourceURI))
		if err != nil || source.String() == "" {
			result = errors.New("OPC: relationship source URI reference shall be a URI or a relative reference")
		} else if isRelationshipURI(source.ResolveReference(uri).String()) {
			result = errors.New("OPC: The Relationships part shall not have relationships to any other part")
		}
	}

	return result
}

// newRelationship creates a new Relationship
func newRelationship(id, relType, sourceURI, targetURI string, targetMode TargetMode) (*Relationship, error) {
	// ISO/IEC 29500-2 M1.26
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("OPC: relationship identifier cannot be empty string or a string with just spaces")
	}

	if strings.TrimSpace(relType) == "" {
		return nil, errors.New("OPC: relationship type cannot be empty string or a string with just spaces")
	}

	if err := validateRelationshipTarget(sourceURI, targetURI, targetMode); err != nil {
		return nil, err
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

// TargetURI returns the targetURI of the relationship.
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

// writeToXML encodes the relationship to the target.
func (r *Relationship) writeToXML(e *xml.Encoder) error {
	return e.EncodeElement(r.toXML(), xml.StartElement{Name: xml.Name{Space: "", Local: relationshipName}})
}

type relationable struct {
	sourceURI     string
	relationships map[string]*Relationship
}

func (r *relationable) CreateRelationship(id, relType, targetURI string, targetMode TargetMode) (*Relationship, error) {
	if _, ok := r.relationships[id]; ok {
		return nil, errors.New("OPC: relationship ID shall be unique within the Relationship part")
	}
	rel, err := newRelationship(id, relType, r.sourceURI, targetURI, targetMode)
	if err != nil {
		return nil, err
	}
	r.relationships[id] = rel
	return rel, nil
}

func (r *relationable) HasRelationship() bool {
	return len(r.relationships) > 0
}

func (r *relationable) Relationships() []*Relationship {
	v := make([]*Relationship, 0, len(r.relationships))
	for _, rel := range r.relationships {
		v = append(v, rel)
	}
	return v
}
