// Package gopc implements the ISO/IEC 29500-2, also known as the "Open Packaging Convention".
// The Open Packaging specification describes an abstract model and physical format conventions for the use of
// XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and
// resources of a document within a package. It is intended to support the content types and organization for
// various applications and is written for developers who are building systems that process package content.
// In addition, this Open Packaging specification defines common services that can be included in a package, such
// as Core Properties and Digital Signatures.
// A primary goal is to ensure the interoperability of independently created software and hardware systems that
// produce or consume package content and use common services. This Open Packaging specification defines the
// formal requirements that producers and consumers must satisfy in order to achieve interoperability.
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
	parts         map[string]*Part
	relationships map[string]*Relationship
}

// NewPackage returns a new initilized Package.
func NewPackage() *Package {
	return &Package{
		parts:         make(map[string]*Part, 0),
		relationships: make(map[string]*Relationship, 0),
	}
}

// CreatePart adds a new Part to the Package.
func (p *Package) CreatePart(uri, contentType string, compressionOption CompressionOption) (*Part, error) {
	upperURI := strings.ToUpper(uri)
	if _, ok := p.parts[upperURI]; ok {
		return nil, errors.New("OPC: packages shall not contain equivalent part names, and package implementers shall neither create nor recognize packages with equivalent part names [M1.12]")
	}

	if p.checkPrefixCollision(upperURI) {
		return nil, errors.New("OPC: a package implementer shall neither create nor recognize a part with a part name derived from another part name by appending segments to it [M1.11]")
	}

	part, err := newPart(uri, contentType, compressionOption)
	if err != nil {
		return nil, err
	}
	p.parts[upperURI] = part
	return part, nil
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
