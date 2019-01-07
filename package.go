// Package gopc implements the ISO/IEC 29500-2, also known as the "Open Packaging Convention".
//
// The Open Packaging specification describes an abstract model and physical format conventions for the use of
// XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and
// resources of a document within a package.
//
// The OPC is the foundation technology for many new file formats: .docx, .pptx, .xlsx, .3mf, .dwfx, ...
package gopc

import (
	"errors"
	"sort"
	"strings"
)

// A Package is a container that holds a collection of parts. The purpose of the package is to aggregate constituent
// components of a document (or other type of content) into a single object.
// The package is also capable of storing relationships between parts.
// Defined in ISO/IEC 29500-2 §9.
type Package struct {
	parts         map[string]*Part
	Relationships []*Relationship
}

func newPackage() *Package {
	return &Package{
		parts: make(map[string]*Part, 0),
	}
}

func (p *Package) add(part *Part) error {
	if err := part.validate(); err != nil {
		return err
	}
	upperURI := strings.ToUpper(part.Name)
	if _, ok := p.parts[upperURI]; ok {
		return errors.New("OPC: packages shall not contain equivalent part names, and package implementers shall neither create nor recognize packages with equivalent part names [M1.12]")
	}
	if p.checkPrefixCollision(upperURI) {
		return errors.New("OPC: a package implementer shall neither create nor recognize a part with a part name derived from another part name by appending segments to it [M1.11]")
	}
	p.parts[upperURI] = part
	return nil
}

func (p *Package) deletePart(uri string) {
	delete(p.parts, strings.ToUpper(uri))
}

func (p *Package) checkPrefixCollision(uri string) bool {
	keys := make([]string, len(p.parts)+1)
	keys[0] = uri
	i := 1
	for k := range p.parts {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for i, k := range keys {
		if k == uri {
			if i > 0 && p.checkStringsPrefixCollision(uri, keys[i-1]) {
				return true
			}
			if i < len(keys)-1 && p.checkStringsPrefixCollision(keys[i+1], uri) {
				return true
			}
		}
	}
	return false
}

func (p *Package) checkStringsPrefixCollision(s1, s2 string) bool {
	return strings.HasPrefix(s1, s2) && len(s1) > len(s2) && s1[len(s2)] == '/'
}
