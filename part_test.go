package gopc

import (
	"reflect"
	"testing"
)

var fakeURL = "/doc/a.xml"

func createFakePart(uri, contentType string) *Part {
	return &Part{
		relationable:      relationable{uri, make(map[string]*Relationship, 0)},
		uri:               uri,
		contentType:       contentType,
		compressionOption: CompressionNone}
}

func Test_newPart(t *testing.T) {
	type args struct {
		uri               string
		contentType       string
		compressionOption CompressionOption
	}
	tests := []struct {
		name    string
		args    args
		want    *Part
		wantErr bool
	}{
		{"base", args{fakeURL, "application/HTML", CompressionNone}, createFakePart(fakeURL, "application/html"), false},
		{"baseWithParameters", args{fakeURL, "TEXT/html; charset=ISO-8859-4", CompressionNone}, createFakePart(fakeURL, "text/html; charset=ISO-8859-4"), false},
		{"baseWithTwoParams", args{fakeURL, "TEXT/html; charset=ISO-8859-4;q=2", CompressionNone}, createFakePart(fakeURL, "text/html; charset=ISO-8859-4; q=2"), false},
		{"invalidMediaParams", args{fakeURL, "TEXT/html; charset=ISO-8859-4 q=2", CompressionNone}, nil, true},
		{"mediaParamNoName", args{fakeURL, "TEXT/html; =ISO-8859-4", CompressionNone}, nil, true},
		{"duplicateParamName", args{fakeURL, "TEXT/html; charset=ISO-8859-4; charset=ISO-8859-4", CompressionNone}, nil, true},
		{"linearSpace", args{fakeURL, "TEXT /html; charset=ISO-8859-4;q=2", CompressionNone}, nil, true},
		{"noSlash", args{fakeURL, "application", CompressionNone}, nil, true},
		{"unexpectedContent", args{fakeURL, "application/html/html", CompressionNone}, nil, true},
		{"noMediaType", args{fakeURL, "/html", CompressionNone}, nil, true},
		{"unexpectedToken", args{fakeURL, "application/", CompressionNone}, nil, true},
		{"incorrectURI", args{"", "fakeContentType", CompressionNone}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPart(tt.args.uri, tt.args.contentType, tt.args.compressionOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("newPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_URI(t *testing.T) {
	tests := []struct {
		name string
		p    *Part
		want string
	}{
		{"base", new(Part), ""},
		{"partURI", createFakePart(fakeURL, "fakeContentType"), fakeURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.URI(); got != tt.want {
				t.Errorf("Part.URI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_ContentType(t *testing.T) {
	tests := []struct {
		name string
		p    *Part
		want string
	}{
		{"base", new(Part), ""},
		{"partContentType", createFakePart(fakeURL, "fakeContentType"), "fakeContentType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ContentType(); got != tt.want {
				t.Errorf("Part.ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_CompressionOption(t *testing.T) {
	tests := []struct {
		name string
		p    *Part
		want CompressionOption
	}{
		{"base", new(Part), CompressionNormal},
		{"partCompressionOption", createFakePart(fakeURL, "fakeContentType"), CompressionNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.CompressionOption(); got != tt.want {
				t.Errorf("Part.CompressionOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePartName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{""}, true},
		{"onlyspaces", args{"  "}, true},
		{"invalidURL", args{"/docs%/a.xml"}, true},
		{"emptySegment", args{"/doc//a.xml"}, true},
		{"abs uri", args{"http://docs//a.xml"}, true},
		{"not rel uri", args{"docs/a.xml"}, true},
		{"endSlash", args{"/docs/a.xml/"}, true},
		{"endDot", args{"/docs/a.xml."}, true},
		{"dot", args{"/docs/./a.xml"}, true},
		{"twoDots", args{"/docs/../a.xml"}, true},
		{"reserved", args{"/docs/%7E/a.xml"}, true},
		{"withQuery", args{"/docs/a.xml?a=2"}, true},
		{"notencodechar", args{"/€/a.xml"}, true},
		{"encodedBSlash", args{"/%5C/a.xml"}, true},
		{"encodedBSlash", args{"/%2F/a.xml"}, true},
		{"encodechar", args{"/%E2%82%AC/a.xml"}, false},
		{"base", args{"/docs/a.xml"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePartName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePartName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
