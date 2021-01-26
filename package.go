// Package opc implements the ISO/IEC 29500-2, also known as the "Open Packaging Convention".
//
// The Open Packaging specification describes an abstract model and physical format conventions for the use of
// XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and
// resources of a document within a package.
//
// The OPC is the foundation technology for many new file formats: .docx, .pptx, .xlsx, .3mf, .dwfx, ...
package opc

import (
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"path"
	"sort"
	"strings"
)

const (
	corePropsRel            = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	corePropsContentType    = "application/vnd.openxmlformats-package.core-properties+xml"
	corePropsDefaultName    = "/props/core.xml"
	contentTypesName        = "/[Content_Types].xml"
	relationshipContentType = "application/vnd.openxmlformats-package.relationships+xml"
	packageRelName          = "/_rels/.rels"
)

type pkg struct {
	parts        map[string]struct{}
	contentTypes contentTypes
}

func newPackage() *pkg {
	return &pkg{
		parts: make(map[string]struct{}, 0),
	}
}

func (p *pkg) partExists(partName string) bool {
	_, ok := p.parts[partName]
	return ok
}

func (p *pkg) add(part *Part) error {
	if err := part.validate(); err != nil {
		return err
	}
	name := strings.ToUpper(NormalizePartName(part.Name))
	if p.partExists(name) {
		return newError(112, part.Name)
	}
	if p.checkPrefixCollision(name) {
		return newError(111, part.Name)
	}
	p.contentTypes.add(name, part.ContentType)
	p.parts[name] = struct{}{}
	return nil
}

func (p *pkg) deletePart(uri string) {
	delete(p.parts, strings.ToUpper(uri))
}

func (p *pkg) checkPrefixCollision(uri string) bool {
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

func (p *pkg) encodeContentTypes(w io.Writer) error {
	w.Write(([]byte)(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(p.contentTypes.toXML())
}

func (p *pkg) checkStringsPrefixCollision(s1, s2 string) bool {
	return strings.HasPrefix(s1, s2) && len(s1) > len(s2) && s1[len(s2)] == '/'
}

type contentTypesXML struct {
	XMLName xml.Name      `xml:"Types"`
	XML     string        `xml:"xmlns,attr"`
	Types   []interface{} `xml:",any"`
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
	defaults  map[string]string // extension:contenttype
	overrides map[string]string // partname:contenttype
}

func (c *contentTypes) toXML() *contentTypesXML {
	cx := &contentTypesXML{XML: "http://schemas.openxmlformats.org/package/2006/content-types"}
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

// Add needs a valid content type, else the behavior is undefined
func (c *contentTypes) add(partName, contentType string) error {
	// Process descrived in ISO/IEC 29500-2 §10.1.2.3
	t, params, _ := mime.ParseMediaType(contentType)
	contentType = mime.FormatMediaType(t, params)

	ext := strings.ToLower(path.Ext(partName))
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

func (c *contentTypes) findType(name string) (string, error) {
	if t, ok := c.overrides[strings.ToUpper(name)]; ok {
		return t, nil
	}
	ext := path.Ext(name)
	if ext != "" {
		if t, ok := c.defaults[strings.ToLower(ext[1:])]; ok {
			return t, nil
		}
	}
	return "", newError(208, name)
}

type corePropertiesXMLMarshal struct {
	XMLName        xml.Name    `xml:"coreProperties"`
	XML            string      `xml:"xmlns,attr"`
	XMLDCTERMS     string      `xml:"xmlns:dcterms,attr"`
	XMLDC          string      `xml:"xmlns:dc,attr"`
	XMLXSI         string      `xml:"xmlns:xsi,attr"`
	Category       string      `xml:"category,omitempty"`
	ContentStatus  string      `xml:"contentStatus,omitempty"`
	Created        w3CDateTime `xml:"dcterms:created,omitempty"`
	Creator        string      `xml:"dc:creator,omitempty"`
	Description    string      `xml:"dc:description,omitempty"`
	Identifier     string      `xml:"dc:identifier,omitempty"`
	Keywords       string      `xml:"keywords,omitempty"`
	Language       string      `xml:"dc:language,omitempty"`
	LastModifiedBy string      `xml:"lastModifiedBy,omitempty"`
	LastPrinted    w3CDateTime `xml:"lastPrinted,omitempty"`
	Modified       w3CDateTime `xml:"dcterms:modified,omitempty"`
	Revision       string      `xml:"revision,omitempty"`
	Subject        string      `xml:"dc:subject,omitempty"`
	Title          string      `xml:"dc:title,omitempty"`
	Version        string      `xml:"version,omitempty"`
}

type corePropertiesXMLUnmarshal struct {
	XMLName        xml.Name `xml:"coreProperties"`
	XML            string   `xml:"xmlns,attr"`
	XMLDCTERMS     string   `xml:"dcterms,attr"`
	XMLDC          string   `xml:"dc,attr"`
	Category       string   `xml:"category,omitempty"`
	ContentStatus  string   `xml:"contentStatus,omitempty"`
	Created        string   `xml:"created,omitempty"`
	Creator        string   `xml:"creator,omitempty"`
	Description    string   `xml:"description,omitempty"`
	Identifier     string   `xml:"identifier,omitempty"`
	Keywords       string   `xml:"keywords,omitempty"`
	Language       string   `xml:"language,omitempty"`
	LastModifiedBy string   `xml:"lastModifiedBy,omitempty"`
	LastPrinted    string   `xml:"lastPrinted,omitempty"`
	Modified       string   `xml:"modified,omitempty"`
	Revision       string   `xml:"revision,omitempty"`
	Subject        string   `xml:"subject,omitempty"`
	Title          string   `xml:"title,omitempty"`
	Version        string   `xml:"version,omitempty"`
}

type w3CDateTime string

func (s w3CDateTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type xmlType struct {
		XSITYPE string `xml:"xsi:type,attr"`
		Value   string `xml:",chardata"`
	}
	return e.EncodeElement(xmlType{"dcterms:W3CDTF", string(s)}, start)
}

// CoreProperties enable users to get and set well-known and common sets of property metadata within packages.
type CoreProperties struct {
	PartName       string // Won't be written to the package, only used to indicate the location of the CoreProperties part. If empty the default location is "/props/core.xml".
	RelationshipID string // Won't be written to the package, only used to indicate the relationship ID for target "/props/core.xml".
	Category       string // A categorization of the content of this package.
	ContentStatus  string // The status of the content.
	Created        string // Date of creation of the resource.
	Creator        string // An entity primarily responsible for making the content of the resource.
	Description    string // An explanation of the content of the resource.
	Identifier     string // An unambiguous reference to the resource within a given context.
	Keywords       string // A delimited set of keywords to support searching and indexing.
	Language       string // The language of the intellectual content of the resource.
	LastModifiedBy string // The user who performed the last modification.
	LastPrinted    string // The date and time of the last printing.
	Modified       string // Date on which the resource was changed.
	Revision       string // The revision number.
	Subject        string // The topic of the content of the resource.
	Title          string // The name given to the resource.
	Version        string // The version number.
}

func (c *CoreProperties) encode(w io.Writer) error {
	w.Write(([]byte)(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(&corePropertiesXMLMarshal{
		xml.Name{Local: "coreProperties"},
		"http://schemas.openxmlformats.org/package/2006/metadata/core-properties",
		"http://purl.org/dc/terms/",
		"http://purl.org/dc/elements/1.1/",
		"http://www.w3.org/2001/XMLSchema-instance",
		c.Category, c.ContentStatus, w3CDateTime(c.Created),
		c.Creator, c.Description, c.Identifier,
		c.Keywords, c.Language, c.LastModifiedBy,
		w3CDateTime(c.LastPrinted), w3CDateTime(c.Modified), c.Revision,
		c.Subject, c.Title, c.Version,
	})
}

func decodeCoreProperties(r io.Reader, props *CoreProperties) error {
	propDecode := new(corePropertiesXMLUnmarshal)
	if err := xml.NewDecoder(r).Decode(propDecode); err != nil {
		return fmt.Errorf("opc: %s: cannot be decoded: %v", contentTypesName, err)
	}
	props.Category = propDecode.Category
	props.ContentStatus = propDecode.ContentStatus
	props.Created = propDecode.Created
	props.Creator = propDecode.Creator
	props.Description = propDecode.Description
	props.Identifier = propDecode.Identifier
	props.Keywords = propDecode.Keywords
	props.Language = propDecode.Language
	props.LastModifiedBy = propDecode.LastModifiedBy
	props.LastPrinted = propDecode.LastPrinted
	props.Modified = propDecode.Modified
	props.Revision = propDecode.Revision
	props.Subject = propDecode.Subject
	props.Title = propDecode.Title
	props.Version = propDecode.Version
	return nil
}
