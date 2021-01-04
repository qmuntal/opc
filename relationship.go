package opc

import (
	"encoding/xml"
	"fmt"
	"io"
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

const externalMode = "External"

// Relationship is used to express a relationship between a source and a target part.
// If the ID is not specified a unique ID will be generated following the pattern rIdN.
// If the TargetMode is not specified the default value is Internal.
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

func newRelationshipID(rels []*Relationship) string {
	ids := make([]string, len(rels))
	for i, rel := range rels {
		ids[i] = rel.ID
	}
	idFunc := func(i int) string { return fmt.Sprintf("rId%d", i) }
	var (
		i  int
		id = idFunc(0)
	)
	for isValueInList(id, ids) {
		i++
		id = idFunc(i)
	}
	return id
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
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

func (r *Relationship) toXML() *relationshipXML {
	var targetMode string
	if r.TargetMode == ModeExternal {
		targetMode = externalMode
	}
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
	if !validEncoded(r.TargetURI) {
		return newErrorRelationship(128, sourceURI, r.ID)
	}
	// ISO/IEC 29500-2 M1.29
	if r.TargetMode == ModeInternal {
		if !isInternal(r.TargetURI) {
			return newErrorRelationship(129, sourceURI, r.ID)
		}
		source := strings.TrimSpace(sourceURI)
		if source == "" || isRelationshipURI(ResolveRelationship(sourceURI, r.TargetURI)) {
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
	re := &relationshipsXML{XML: "http://schemas.openxmlformats.org/package/2006/relationships"}
	for _, r := range rs {
		re.RelsXML = append(re.RelsXML, r.toXML())
	}
	w.Write(([]byte)(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(re)
}

func decodeRelationships(r io.Reader, partName string) ([]*Relationship, error) {
	relDecode := new(relationshipsXML)
	if err := xml.NewDecoder(r).Decode(relDecode); err != nil {
		return nil, fmt.Errorf("opc: %s: cannot be decoded: %v", partName, err)
	}
	rel := make([]*Relationship, len(relDecode.RelsXML))
	for i, rl := range relDecode.RelsXML {
		newRel := &Relationship{ID: rl.ID, TargetURI: rl.TargetURI, Type: rl.RelType}
		if rl.Mode == "" || rl.Mode == "Internal" {
			newRel.TargetMode = ModeInternal
		} else {
			newRel.TargetMode = ModeExternal
		}
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

func isInternal(rawurl string) bool {
	for i := 0; i < len(rawurl); i++ {
		c := rawurl[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		// do nothing
		case '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.':
			if i == 0 {
				return true
			}
		case c == ':':
			if i == 0 {
				return true
			}
			return false
		default:
			// we have encountered an invalid character,
			// so there is no valid scheme
			return true
		}
	}
	return true
}
