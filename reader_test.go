package gopc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

func Test_newReader(t *testing.T) {
	p1 := newPackage()
	p1.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml"}
	p1.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p1.parts["/FILES.XML"] = &Part{Name: "/files.xml", ContentType: "application/xml"}
	p1.contentTypes.addOverride("/DOCPROPS/APP.XML", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p1.contentTypes.addDefault("png", "image/png")
	p1.contentTypes.addDefault("xml", "application/xml")

	p2 := newPackage()
	p2.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
			&Relationship{ID: "rel-2", Type: "text/txt", TargetURI: "/", TargetMode: ModeExternal},
		},
	}
	p2.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p2.parts["/FILES.XML"] = &Part{Name: "/files.xml", ContentType: "application/xml"}
	p2.contentTypes.addOverride("/DOCPROPS/APP.XML", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p2.contentTypes.addDefault("xml", "application/xml")
	p2.contentTypes.addDefault("png", "image/png")

	tests := []struct {
		name    string
		files   []archiveFile
		want    *pkg
		wantErr bool
	}{
		{"baseWithEmptyDirectory", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("3D/", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("files.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p1, false},
		{"baseWithRels", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile(
				"docProps/_rels/app.xml.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rel-1", "text/txt", "/").withRelMode("rel-2", "text/txt", "/", "External").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("files.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p2, false},
		{"baseWithoutRelationships", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("files.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.p, tt.want) {
				t.Errorf("newReader() = %v, want %v", got.p, tt.want)
			}
		})
	}
}

func Test_newReader_ContentType(t *testing.T) {
	invalidType := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Fake Extension="" ContentType=""/>
</Types>`

	incorrectOverrideXML := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

	incorrectDefaultXML := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml" PartName="/docProps/app.xml">
</Types>`

	tests := []struct {
		name    string
		files   []archiveFile
		want    *pkg
		wantErr bool
	}{
		{"openError", []archiveFile{
			newMockFile("a.xml", nil, nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},

		{"duplicatedExtensionDefault", []archiveFile{
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withDefault("image/png", "png").withDefault("image/png2", "png").String())),
				nil,
			),
		}, nil, true},

		{"duplicatedPartNameOverride", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").String())),
				nil,
			),
		}, nil, true},

		{"emptyExtension", []archiveFile{
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withDefault("image/png", "png").withDefault("image/png2", "").String())),
				nil,
			),
		}, nil, true},

		{"invalidType", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(invalidType)), nil),
		}, nil, true},

		{"incorrectDefaultXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectDefaultXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"incorrectOverrideXML", []archiveFile{
			newMockFile("[Content_Types].xml", ioutil.NopCloser(bytes.NewBufferString(incorrectOverrideXML)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"partWithoutContentType", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo2.jpg", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, nil, true},

		{"noContentType", []archiveFile{
			newMockFile("docProps/app.xml", nil, nil),
			newMockFile("pictures/photo2.jpg", nil, nil),
		}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.p, tt.want) {
				t.Errorf("newReader() = %v, want %v", got.p, tt.want)
			}
		})
	}
}

func Test_newReader_PartRelationships(t *testing.T) {
	p3 := newPackage()
	p3.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
			&Relationship{ID: "rel-2", Type: "text/txt", TargetURI: "/", TargetMode: ModeExternal},
		},
	}
	p3.parts["/PICTURES/PHOTO.PNG"] = &Part{Name: "/pictures/photo.png", ContentType: "image/png"}
	p3.parts["/FILES.XML"] = &Part{Name: "/files.xml", ContentType: "application/xml"}
	p3.contentTypes.addOverride("/DOCPROPS/APP.XML", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p3.contentTypes.addDefault("xml", "application/xml")
	p3.contentTypes.addDefault("png", "image/png")

	p4 := newPackage()
	p4.parts["/DOCPROPS/APP.XML"] = &Part{Name: "/docProps/app.xml", ContentType: "application/vnd.openxmlformats-officedocument.extended-properties+xml",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-1", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
			&Relationship{ID: "rel-2", Type: "text/txt", TargetURI: "/", TargetMode: ModeExternal},
		},
	}
	p4.parts["/PICTURES/SEASON/SUMMER/PHOTO.PNG"] = &Part{Name: "/pictures/season/summer/photo.png", ContentType: "image/png",
		Relationships: []*Relationship{
			&Relationship{ID: "rel-3", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
			&Relationship{ID: "rel-4", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
			&Relationship{ID: "rel-5", Type: "text/txt", TargetURI: "/", TargetMode: ModeInternal},
		},
	}
	p4.parts["/PICTURES/SUMMER/PHOTO2.PNG"] = &Part{Name: "/pictures/summer/photo2.png", ContentType: "image/png"}
	p4.parts["/FILES.XML"] = &Part{Name: "/files.xml", ContentType: "application/xml"}
	p4.contentTypes.addOverride("/DOCPROPS/APP.XML", "application/vnd.openxmlformats-officedocument.extended-properties+xml")
	p4.contentTypes.addDefault("xml", "application/xml")
	p4.contentTypes.addDefault("png", "image/png")

	tests := []struct {
		name    string
		files   []archiveFile
		want    *pkg
		wantErr bool
	}{

		{"complexRelationships", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile(
				"_rels/.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rel-1", "text/txt", "/docProps/app.xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"docProps/_rels/app.xml.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rel-1", "text/txt", "/").withRelMode("rel-2", "text/txt", "/", "External").String())),
				nil,
			),
			newMockFile("files.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, p3, false},

		{"ComplexRoute", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"docProps/_rels/app.xml.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rel-1", "text/txt", "/").withRelMode("rel-2", "text/txt", "/", "External").String())),
				nil,
			),
			newMockFile("files.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/summer/photo2.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("pictures/season/summer/photo.png", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"pictures/season/summer/_rels/photo.png.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rel-3", "text/txt", "/").withRel("rel-4", "text/txt", "/").withRel("rel-5", "text/txt", "/").String())),
				nil,
			),
		}, p4, false},

		{"openEmptyXML", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},

		{"decodeMalformedXML", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("docProps/_rels/app.xml.rels", ioutil.NopCloser(bytes.NewBufferString("relations")), nil),
		}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.p, tt.want) {
				t.Errorf("newReader() = %v, want %v", got.p, tt.want)
			}
		})
	}
}

func Test_newReader_CoreProperties(t *testing.T) {
	coreFile := `<?xml version="1.0" encoding="UTF-8" standalone="true"?>
	<cp:coreProperties xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties">
	<dc:creator/>
	<cp:lastModifiedBy/>
	<dcterms:created xsi:type="dcterms:W3CDTF">2015-06-05T18:19:34Z</dcterms:created>
	<dcterms:modified xsi:type="dcterms:W3CDTF">2019-01-24T19:58:26Z</dcterms:modified>
	</cp:coreProperties>`

	cp := &CoreProperties{Created: "2015-06-05T18:19:34Z", Modified: "2019-01-24T19:58:26Z"}

	tests := []struct {
		name    string
		files   []archiveFile
		want    CoreProperties
		wantErr bool
	}{

		{"base", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").withOverride("application/vnd.openxmlformats-package.core-properties+xml", "/docProps/core.xml").String())),
				nil,
			),
			newMockFile(
				"_rels/.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rId3", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", "docProps/app.xml").withRel("rId2", "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", "docProps/core.xml").withRel("rId1", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument", "xl/workbook.xml").String())),
				nil,
			),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString(coreFile)), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, *cp, false},
		{"decodeError", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").withOverride("application/vnd.openxmlformats-package.core-properties+xml", "/docProps/core.xml").String())),
				nil,
			),
			newMockFile(
				"_rels/.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rId3", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", "docProps/app.xml").withRel("rId2", "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", "docProps/core.xml").withRel("rId1", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument", "xl/workbook.xml").String())),
				nil,
			),
			newMockFile("docProps/core.xml", ioutil.NopCloser(bytes.NewBufferString("{a : 2}")), nil),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, *cp, true},
		{"openError", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").withOverride("application/vnd.openxmlformats-package.core-properties+xml", "/docProps/core.xml").String())),
				nil,
			),
			newMockFile(
				"_rels/.rels",
				ioutil.NopCloser(bytes.NewBufferString(new(relsBuilder).withRel("rId3", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", "docProps/app.xml").withRel("rId2", "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", "docProps/core.xml").withRel("rId1", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument", "xl/workbook.xml").String())),
				nil,
			),
			newMockFile("docProps/core.xml", ioutil.NopCloser(nil), errors.New("")),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, *cp, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Properties, tt.want) {
				t.Errorf("newReader() = %v, want %v", got.Properties, tt.want)
			}
		})
	}
}

func Test_newReader_PackageRelationships(t *testing.T) {
	validPackageRelationships := `<?xml version="1.0" encoding="UTF-8" standalone="true"?>
	<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
	<Relationship Target="http://www.custom.com/images/pic1.jpg" Type="http://www.custom.com/external-resource" Id="rId3" TargetMode="External"/>
	<Relationship Target="DOCPROPS/app.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Id="rId2" TargetMode="Internal"/>
	<Relationship Target="xl/workbook.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Id="rId1"/>
	<Relationship Target="./xl/other.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Id="rId4"/>
	</Relationships>`

	r := []*Relationship{
		{ID: "rId3", Type: "http://www.custom.com/external-resource", TargetURI: "http://www.custom.com/images/pic1.jpg", TargetMode: ModeExternal},
		{ID: "rId2", Type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", TargetURI: "/DOCPROPS/app.xml", TargetMode: ModeInternal},
		{ID: "rId1", Type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument", TargetURI: "/xl/workbook.xml", TargetMode: ModeInternal},
		{ID: "rId4", Type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument", TargetURI: "./xl/other.xml", TargetMode: ModeInternal},
	}
	tests := []struct {
		name    string
		files   []archiveFile
		want    []*Relationship
		wantErr bool
	}{
		{"base", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/app.xml").withOverride("application/vnd.openxmlformats-package.core-properties+xml", "/docProps/core.xml").String())),
				nil,
			),
			newMockFile("_rels/.rels", ioutil.NopCloser(bytes.NewBufferString(validPackageRelationships)), nil),
			newMockFile("docprops/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
		}, r, false},

		{"openEmptyXMLPackage", []archiveFile{
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("_rels/.rels", ioutil.NopCloser(nil), errors.New("")),
		}, nil, true},

		{"decodeMalformedXMLPackage", []archiveFile{
			newMockFile(
				"[Content_Types].xml",
				ioutil.NopCloser(bytes.NewBufferString(new(cTypeBuilder).withOverride("application/vnd.openxmlformats-officedocument.extended-properties+xml", "/docProps/APP.xml").withDefault("image/png", "png").withDefault("application/xml", "xml").String())),
				nil,
			),
			newMockFile("docProps/app.xml", ioutil.NopCloser(bytes.NewBufferString("")), nil),
			newMockFile("_rels/.rels", ioutil.NopCloser(bytes.NewBufferString("relations")), nil),
		}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := new(mockArchive)
			a.On("Files").Return(tt.files)
			got, err := newReader(a)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Relationships, tt.want) {
				t.Errorf("newReader() = %v, want %v", got.Relationships, tt.want)
			}
		})
	}
}

type mockFile struct {
	mock.Mock
}

func (m *mockFile) Open() (io.ReadCloser, error) {
	args := m.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFile) Name() string {
	args := m.Called()
	return args.String(0)
}

type mockArchive struct {
	mock.Mock
}

func (m *mockArchive) Files() []archiveFile {
	args := m.Called()
	return args.Get(0).([]archiveFile)
}

func newMockFile(name string, r io.ReadCloser, e error) *mockFile {
	f := new(mockFile)
	f.On("Name").Return(name)
	if r != nil {
		f.On("Open").Return(r, e)
	}
	return f
}

type relsBuilder struct {
	rels strings.Builder
}

func (r *relsBuilder) withRel(id, relType, target string) *relsBuilder {
	r.rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"/>`, id, relType, target))
	return r
}

func (r *relsBuilder) withRelMode(id, relType, target, mode string) *relsBuilder {
	r.rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s" TargetMode="%s"/>`, id, relType, target, mode))
	return r
}

func (r *relsBuilder) String() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">%s</Relationships>`, r.rels.String())
}

type cTypeBuilder struct {
	ctype strings.Builder
}

func (ct *cTypeBuilder) withOverride(cType, pName string) *cTypeBuilder {
	ct.ctype.WriteString(fmt.Sprintf(`<Override ContentType="%s" PartName="%s"/>`, cType, pName))
	return ct
}

func (ct *cTypeBuilder) withDefault(cType, ext string) *cTypeBuilder {
	ct.ctype.WriteString(fmt.Sprintf(`<Default Extension="%s" ContentType="%s"/>`, ext, cType))
	return ct
}

func (ct *cTypeBuilder) String() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">%s</Types>`, ct.ctype.String())
}
