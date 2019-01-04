// Package gopc implements the ISO/IEC 29500-2, also known as the "Open Packaging Convention".
//
// The Open Packaging specification describes an abstract model and physical format conventions for the use of
// XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and
// resources of a document within a package.
package gopc

import (
	"errors"
	"sort"
	"strings"
)

// A package is a container that holds a collection of parts. The purpose of the package is to aggregate constituent
// components of a document (or other type of content) into a single object.
// The package is also capable of storing relationships between parts.
// Defined in ISO/IEC 29500-2 §9.
type Package struct {
	relationer
	parts map[string]*Part
}

func newPackage() *Package {
	return &Package{
		relationer: relationer{"/", make(map[string]*Relationship, 0)},
		parts:      make(map[string]*Part, 0),
	}
}

func (p *Package) create(uri, contentType string, compressionOption CompressionOption) (*Part, error) {
	part, err := newPart(uri, contentType, compressionOption)
	if err != nil {
		return nil, err
	}
	if err = p.add(part); err != nil {
		return nil, err
	}
	return part, nil
}

func (p *Package) add(part *Part) error {
	upperURI := strings.ToUpper(part.uri)
	if _, ok := p.parts[upperURI]; ok {
		return errors.New("OPC: packages shall not contain equivalent part names, and package implementers shall neither create nor recognize packages with equivalent part names [M1.12]")
	}

	if p.checkPrefixCollision(upperURI) {
		return errors.New("OPC: a package implementer shall neither create nor recognize a part with a part name derived from another part name by appending segments to it [M1.11]")
	}

	return nil
}

func (p *Package) deletePart(uri string) {
	upperURI := strings.ToUpper(uri)
	if part, ok := p.parts[upperURI]; ok {
		part.clearRelationships()
		delete(p.parts, upperURI)
	}
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
