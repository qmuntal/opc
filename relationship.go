package gopc

import (
	"crypto/rand"
	"encoding/xml"
	"errors"
	"fmt"
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

// Relationship is used to express a relationship between a source and a target part.
// The only way to create a Relationship, is to call the Part.CreateRelationship()
// or Package.CreateRelationship(). A relationship is owned by a part or by the package itself.
// If the source part is deleted all the relationships it owns are also deleted.
// A target of the relationship need not be present.
// Defined in ISO/IEC 29500-2 §9.3.
type Relationship struct {
	ID         string
	RelType    string
	TargetURI  string
	TargetMode TargetMode
	sourceURI  string
}

type relationshipXML struct {
	ID        string `xml:"Id,attr"`
	RelType   string `xml:"Type,attr"`
	TargetURI string `xml:"Target,attr"`
	Mode      string `xml:"TargetMode,attr,omitempty"`
}

func (r *Relationship) validate() error {
	// ISO/IEC 29500-2 M1.26
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("OPC: relationship identifier cannot be empty string or a string with just spaces")
	}
	if strings.TrimSpace(r.RelType) == "" {
		return errors.New("OPC: relationship type cannot be empty string or a string with just spaces")
	}
	if err := validateRelationshipTarget(r.sourceURI, r.TargetURI, r.TargetMode); err != nil {
		return err
	}
	return nil
}

func (r *Relationship) toXML() *relationshipXML {
	var targetMode string
	if r.TargetMode == ModeExternal {
		targetMode = externalMode
	}
	x := &relationshipXML{ID: r.ID, RelType: r.RelType, TargetURI: r.TargetURI, Mode: targetMode}
	if r.TargetMode == ModeInternal {
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
			result = errors.New("OPC: The relationships part shall not have relationships to any other part")
		}
	}

	return result
}

func uniqueRelationshipID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
