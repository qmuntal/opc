package gopc

import (
	"reflect"
	"strings"
	"testing"
)

var fakeURLUpper = "/DOC/A.XML"

func createFakePackage(m ...string) *Package {
	parts := make(map[string]*Part, len(m))
	for _, s := range m {
		parts[strings.ToUpper(s)] = new(Part)
	}
	return &Package{
		relationer: relationer{"/", make(map[string]*Relationship, 0)},
		parts:      parts,
	}
}

func TestPackage_create(t *testing.T) {
	type args struct {
		uri               string
		contentType       string
		compressionOption CompressionOption
	}
	tests := []struct {
		name    string
		p       *Package
		args    args
		want    *Part
		wantErr bool
	}{
		{"duplicated", createFakePackage(fakeURL), args{fakeURL, "a/b", CompressionNone}, nil, true},
		{"collision1", createFakePackage("/abc.xml", "/xyz/PQR/A.JPG"), args{"/abc.xml/b.xml", "a/b", CompressionNone}, nil, true},
		{"collision2", createFakePackage("/ABC.XML", "/XYZ/PQR/A.JPG"), args{"/xyz/pqr", "a/b", CompressionNone}, nil, true},
		{"errorPart", newPackage(), args{"a.xml", "a/b", CompressionNone}, nil, true},
		{"base", newPackage(), args{"/a.xml", "a/b", CompressionNone}, createFakePart("/a.xml", "a/b"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.create(tt.args.uri, tt.args.contentType, tt.args.compressionOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Package.create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Package.create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newPackage(t *testing.T) {
	tests := []struct {
		name string
		want *Package
	}{
		{"base", &Package{
			relationer: relationer{"/", make(map[string]*Relationship, 0)},
			parts:      make(map[string]*Part, 0),
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
		{"empty", newPackage(), args{fakeURL}},
		{"existing", createFakePackage(fakeURL), args{fakeURL}},
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
