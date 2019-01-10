// Package gopc implements the ISO/IEC 29500-2, also known as the "Open Packaging Convention".
//
// The Open Packaging specification describes an abstract model and physical format conventions for the use of
// XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and
// resources of a document within a package.
//
// The OPC is the foundation technology for many new file formats: .docx, .pptx, .xlsx, .3mf, .dwfx, ...
package gopc

import (
	"encoding/xml"
	"errors"
	"mime"
	"path/filepath"
	"sort"
	"strings"
)

// A Package is a container that holds a collection of parts. The purpose of the package is to aggregate constituent
// components of a document (or other type of content) into a single object.
// The package is also capable of storing relationships between parts.
// Defined in ISO/IEC 29500-2 §9.
type Package struct {
	parts        map[string]*Part
	contentTypes contentTypes
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
	// ISO/IEC 29500-2 M1.12
	if _, ok := p.parts[upperURI]; ok {
		return errors.New("OPC: packages shall not contain equivalent part names")
	}
	// ISO/IEC 29500-2 M1.11
	if p.checkPrefixCollision(upperURI) {
		return errors.New("OPC: a package shall not contain a part with a part name derived from another part name by appending segments to it")
	}
	p.contentTypes.add(part.Name, part.ContentType)
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

type contentTpesXML struct {
	XMLName xml.Name `xml:"Types"`
	Xmlns   string   `xml:"xmlns,attr"`
	Types   []interface{}
}

type defaultContentTypeXML struct {
	XMLName     xml.Name `xml:"Default"`
	Extension   string   `xml:"Extension,attr"`
	ContentType string   `xml:"ContentType,attr"`
}

type overrideContentTypeXML struct {
	XMLName     xml.Name `xml:"Override"`
	PartName    string   `xml:"PartName,attr"`
	ContentType string   `xml:"ContentType,attr"`
}

type contentTypes struct {
	defaults  map[string]string // extenstion:contenttype
	overrides map[string]string // partname:contenttype
}

func (c *contentTypes) toXML() *contentTpesXML {
	cx := &contentTpesXML{Xmlns: "http://schemas.openxmlformats.org/package/2006/content-types"}
	if c.defaults != nil {
		for e, ct := range c.defaults {
			cx.Types = append(cx.Types, &defaultContentTypeXML{Extension: e, ContentType: ct})
		}
	}
	if c.overrides != nil {
		for pn, ct := range c.overrides {
			cx.Types = append(cx.Types, &overrideContentTypeXML{PartName: pn, ContentType: ct})
		}
	}
	return cx
}

func (c *contentTypes) ensureDefaultsMap() {
	if c.defaults == nil {
		c.defaults = make(map[string]string, 0)
	}
}

func (c *contentTypes) ensureOverridesMap() {
	if c.overrides == nil {
		c.overrides = make(map[string]string, 0)
	}
}

// Add needs a valid content type, else the behaviour is undefined
func (c *contentTypes) add(partName, contentType string) error {
	// Process descrived in ISO/IEC 29500-2 §10.1.2.3
	if len(contentType) == 0 {
		return nil
	}
	t, params, _ := mime.ParseMediaType(contentType)
	contentType = mime.FormatMediaType(t, params)

	ext := strings.ToLower(filepath.Ext(partName))
	if len(ext) == 0 {
		c.addOverride(partName, contentType)
		return nil
	}
	ext = ext[1:] // remove dot
	c.ensureDefaultsMap()
	currentType, ok := c.defaults[ext]
	if ok {
		if currentType != contentType {
			c.addOverride(partName, contentType)
		}
	} else {
		c.addDefault(ext, contentType)
	}

	return nil
}

func (c *contentTypes) addOverride(partName, contentType string) {
	c.ensureOverridesMap()
	// ISO/IEC 29500-2 M2.5
	c.overrides[partName] = contentType
}

func (c *contentTypes) addDefault(extension, contentType string) {
	c.ensureDefaultsMap()
	// ISO/IEC 29500-2 M2.5
	c.defaults[extension] = contentType
}
