package gopc

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func createFakePackage(m ...string) *Package {
	parts := make(map[string]*Part, len(m))
	for _, s := range m {
		parts[strings.ToUpper(s)] = new(Part)
	}
	return &Package{
		parts: parts,
	}
}

func Test_newPackage(t *testing.T) {
	tests := []struct {
		name string
		want *Package
	}{
		{"base", &Package{
			parts: make(map[string]*Part, 0),
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPackage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPackage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackage_deletePart(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name string
		p    *Package
		args args
	}{
		{"empty", newPackage(), args{"/a.xml"}},
		{"existing", createFakePackage("/a.xml"), args{"/a.xml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.deletePart(tt.args.uri)
			if _, ok := tt.p.parts[strings.ToUpper(tt.args.uri)]; ok {
				t.Error("Package.deletePart() should have deleted the part")
			}
		})
	}
}

func TestPackage_add(t *testing.T) {
	type args struct {
		part *Part
	}
	tests := []struct {
		name             string
		p                *Package
		args             args
		wantContentTypes contentTypes
		wantErr          bool
	}{
		{"base", createFakePackage("/b.xml"), args{&Part{"/A.xml", "a/b", nil}}, contentTypes{map[string]string{"xml": "a/b"}, nil}, false},
		{"emptyContentType", createFakePackage(), args{&Part{"/A.xml", "", nil}}, contentTypes{}, false},
		{"noExtension", createFakePackage(), args{&Part{"/A", "a/b", nil}}, contentTypes{nil, map[string]string{"/A": "a/b"}}, false},
		{"duplicated", createFakePackage("/a.xml"), args{&Part{"/A.xml", "a/b", nil}}, contentTypes{}, true},
		{"collision1", createFakePackage("/abc.xml", "/xyz/PQR/A.JPG"), args{&Part{"/abc.xml/b.xml", "a/b", nil}}, contentTypes{}, true},
		{"collision2", createFakePackage("/abc.xml", "/xyz/PQR/A.JPG"), args{&Part{"/xyz/pqr", "a/b", nil}}, contentTypes{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.add(tt.args.part); (err != nil) != tt.wantErr {
				t.Errorf("Package.add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.p.contentTypes, tt.wantContentTypes) {
				t.Errorf("Package.add() = %v, want %v", tt.p.contentTypes, tt.wantContentTypes)
			}
		})
	}
}

func buildCoreString(content string) string {
	s := `<?xml version="1.0" encoding="UTF-8"?>`
	s += `<coreProperties xmlns="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"`
	s += ` xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dc="http://purl.org/dc/elements/1.1/">`
	return s + content + "</coreProperties>"
}

func TestCoreProperties_encode(t *testing.T) {
	tests := []struct {
		name    string
		c       *CoreProperties
		wantW   string
		wantErr bool
	}{
		{"empty", &CoreProperties{}, buildCoreString(""), false},
		{"some", &CoreProperties{Category: "A", LastPrinted: "b"}, buildCoreString("<category>A</category><lastPrinted>b</lastPrinted>"), false},
		{"all", &CoreProperties{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o"},
			buildCoreString("<category>a</category><contentStatus>b</contentStatus><dcterms:created>c</dcterms:created><dc:creator>d</dc:creator><dc:description>e</dc:description><dc:identifier>f</dc:identifier><keywords>g</keywords><dc:language>h</dc:language><lastModifiedBy>i</lastModifiedBy><lastPrinted>j</lastPrinted><dcterms:modified>k</dcterms:modified><revision>l</revision><dc:subject>m</dc:subject><dc:title>n</dc:title><version>o</version>"),
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := tt.c.encode(w); (err != nil) != tt.wantErr {
				t.Errorf("CoreProperties.encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("CoreProperties.encode() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
