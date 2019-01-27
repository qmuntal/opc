package gopc

import (
	"encoding/xml"
	"errors"
	"io"
	"math/rand"
	"net/url"
	"strings"
	"time"
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

const externalMode = "External"
const charBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

// Relationship is used to express a relationship between a source and a target part.
// If the ID is not specified a random string with 8 characters will be generated.
// If The TargetMode is not specified the default value is Internal.
// Defined in ISO/IEC 29500-2 §9.3.
type Relationship struct {
	ID         string     // The relationship identifier which shall conform the xsd:ID naming restrictions and unique within the part.
	Type       string     // Defines the role of the relationship.
	TargetURI  string     // Holds a URI that points to a target resource. If expressed as a relative URI, it is resolved against the base URI of the Relationships source part.
	TargetMode TargetMode // Indicates whether or not the target describes a resource inside the package or outside the package.
}

func (r *Relationship) ensureID() {
	if r.ID != "" {
		return
	}

	b := make([]byte, 8)
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charBytes[rd.Intn(len(charBytes))]
	}
	r.ID = string(b)
}

func (r *Relationship) validate(sourceURI string) error {
	// ISO/IEC 29500-2 M1.26
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("OPC: relationship identifier cannot be empty string or a string with just spaces")
	}
	if strings.TrimSpace(r.Type) == "" {
		return errors.New("OPC: relationship type cannot be empty string or a string with just spaces")
	}
	return validateRelationshipTarget(sourceURI, r.TargetURI, r.TargetMode)
}

func (r *Relationship) toXML() *relationshipXML {
	var targetMode string
	if r.TargetMode == ModeExternal {
		targetMode = externalMode
	}
	x := &relationshipXML{ID: r.ID, Type: r.Type, TargetURI: r.TargetURI, Mode: targetMode}
	if r.TargetMode == ModeInternal {
		if !strings.HasPrefix(x.TargetURI, "/") && !strings.HasPrefix(x.TargetURI, "\\") && !strings.HasPrefix(x.TargetURI, ".") {
			x.TargetURI = "/" + x.TargetURI
		}
	}
	return x
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
			// ISO/IEC 29500-2 M1.28
			result = errors.New("OPC: relationship source URI reference shall be a URI or a relative reference")
		} else if isRelationshipURI(source.ResolveReference(uri).String()) {
			// ISO/IEC 29500-2 M1.26
			result = errors.New("OPC: The relationships part shall not have relationships to any other part")
		}
	}

	return result
}

func validateRelationships(sourceUri string, rs []*Relationship) error {
	var s struct{}
	ids := make(map[string]struct{}, 0)
	for _, r := range rs {
		if err := r.validate(sourceUri); err != nil {
			return err
		}
		// ISO/IEC 29500-2 M1.26
		if _, ok := ids[r.ID]; ok {
			return errors.New("OPC: reltionship ID shall be unique within the Relationships part")
		}
		ids[r.ID] = s
	}
	return nil
}

type relationshipXML struct {
	ID        string `xml:"Id,attr"`
	Type      string `xml:"Type,attr"`
	TargetURI string `xml:"Target,attr"`
	Mode      string `xml:"TargetMode,attr,omitempty"`
}

type relationshipsXML struct {
	XMLName xml.Name           `xml:"Relationships"`
	XML     string             `xml:"xmlns,attr"`
	RelsXML []*relationshipXML `xml:"Relationship"`
}

func encodeRelationships(w io.Writer, rs []*Relationship) error {
	w.Write(([]byte)(`<?xml version="1.0" encoding="UTF-8"?>`))
	re := &relationshipsXML{XML: "http://schemas.openxmlformats.org/package/2006/relationships"}
	for _, r := range rs {
		re.RelsXML = append(re.RelsXML, r.toXML())
	}
	return xml.NewEncoder(w).Encode(re)
}
