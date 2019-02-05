package opc

import (
	"encoding/xml"
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

type relationshipsXML struct {
	XMLName xml.Name           `xml:"Relationships"`
	XML     string             `xml:"xmlns,attr"`
	RelsXML []*relationshipXML `xml:"Relationship"`
}

type relationshipXML struct {
	ID        string `xml:"Id,attr"`
	RelType   string `xml:"Type,attr"`
	TargetURI string `xml:"Target,attr"`
	Mode      string `xml:"TargetMode,attr,omitempty"`
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
	if strings.TrimSpace(r.ID) == "" {
		return newErrorRelationship(126, sourceURI, r.ID)
	}
	if strings.TrimSpace(r.Type) == "" {
		return newErrorRelationship(127, sourceURI, r.ID)
	}
	return r.validateRelationshipTarget(sourceURI)
}

func (r *Relationship) normalizeTargetURI() {
	if r.TargetMode == ModeInternal {
		if !strings.HasPrefix(r.TargetURI, "/") && !strings.HasPrefix(r.TargetURI, "\\") && !strings.HasPrefix(r.TargetURI, ".") {
			r.TargetURI = "/" + r.TargetURI
		}
	}
}

func (r *Relationship) toXML() *relationshipXML {
	var targetMode string
	if r.TargetMode == ModeExternal {
		targetMode = externalMode
	}
	r.normalizeTargetURI()
	x := &relationshipXML{ID: r.ID, RelType: r.Type, TargetURI: r.TargetURI, Mode: targetMode}
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
func (r *Relationship) validateRelationshipTarget(sourceURI string) error {
	uri, err := url.Parse(strings.TrimSpace(r.TargetURI))
	if err != nil || uri.String() == "" {
		return newErrorRelationship(128, sourceURI, r.ID)
	}

	// ISO/IEC 29500-2 M1.29
	if r.TargetMode == ModeInternal && uri.IsAbs() {
		return newErrorRelationship(129, sourceURI, r.ID)
	}

	if r.TargetMode != ModeExternal && !uri.IsAbs() {
		source, err := url.Parse(strings.TrimSpace(sourceURI))
		if err != nil || source.String() == "" || isRelationshipURI(source.ResolveReference(uri).String()) {
			return newErrorRelationship(125, sourceURI, r.ID)
		}
	}

	return nil
}

func validateRelationships(sourceURI string, rs []*Relationship) error {
	var s struct{}
	ids := make(map[string]struct{}, 0)
	for _, r := range rs {
		if err := r.validate(sourceURI); err != nil {
			return err
		}
		// ISO/IEC 29500-2 M1.26
		if _, ok := ids[r.ID]; ok {
			return newErrorRelationship(126, sourceURI, r.ID)
		}
		ids[r.ID] = s
	}
	return nil
}

func encodeRelationships(w io.Writer, rs []*Relationship) error {
	w.Write(([]byte)(`<?xml version="1.0" encoding="UTF-8"?>`))
	re := &relationshipsXML{XML: "http://schemas.openxmlformats.org/package/2006/relationships"}
	for _, r := range rs {
		re.RelsXML = append(re.RelsXML, r.toXML())
	}
	return xml.NewEncoder(w).Encode(re)
}

func decodeRelationships(r io.Reader) ([]*Relationship, error) {
	relDecode := new(relationshipsXML)
	if err := xml.NewDecoder(r).Decode(relDecode); err != nil {
		return nil, err
	}
	rel := make([]*Relationship, len(relDecode.RelsXML))
	for i, rl := range relDecode.RelsXML {
		newRel := &Relationship{ID: rl.ID, TargetURI: rl.TargetURI, Type: rl.RelType}
		if rl.Mode == "" || rl.Mode == "Internal" {
			newRel.TargetMode = ModeInternal
		} else {
			newRel.TargetMode = ModeExternal
		}
		newRel.normalizeTargetURI()
		rel[i] = newRel
	}
	return rel, nil
}

type relationshipsPart struct {
	relation map[string][]*Relationship // partname:relationship
}

func (rp *relationshipsPart) findRelationship(name string) []*Relationship {
	if rp.relation == nil {
		rp.relation = make(map[string][]*Relationship)
	}
	return rp.relation[strings.ToUpper(name)]
}

func (rp *relationshipsPart) addRelationship(name string, r []*Relationship) {
	if rp.relation == nil {
		rp.relation = make(map[string][]*Relationship)
	}
	rp.relation[strings.ToUpper(name)] = r
}
