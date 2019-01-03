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
		parts:         parts,
		relationships: nil,
	}
}

func TestPackage_CreatePart(t *testing.T) {
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
		{"errorPart", NewPackage(), args{"a.xml", "a/b", CompressionNone}, nil, true},
		{"base", NewPackage(), args{"/a.xml", "a/b", CompressionNone}, &Part{"/a.xml", "a/b", CompressionNone}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.CreatePart(tt.args.uri, tt.args.contentType, tt.args.compressionOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Package.CreatePart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Package.CreatePart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPackage(t *testing.T) {
	tests := []struct {
		name string
		want *Package
	}{
		{"base", &Package{
			parts:         make(map[string]*Part, 0),
			relationships: make(map[string]*Relationship, 0),
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPackage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPackage() = %v, want %v", got, tt.want)
			}
		})
	}
}
