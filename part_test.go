package gopc

import (
	"testing"
)

func TestNormalizePartName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"base", args{"/a.xml"}, "/a.xml"},
		{"onedot", args{"/./a.xml"}, "/a.xml"},
		{"doubledot", args{"/../a.xml"}, "/a.xml"},
		{"noslash", args{"a.xml"}, "/a.xml"},
		{"folder", args{"/docs/a.xml"}, "/docs/a.xml"},
		{"noext", args{"/docs"}, "/docs"},
		{"win", args{"\\docs\\a.xml"}, "/docs/a.xml"},
		{"winnoslash", args{"docs\\a.xml"}, "/docs/a.xml"},
		{"fragment", args{"/docs/a.xml#a"}, "/docs/a.xml"},
		{"twoslash", args{"//docs/a.xml"}, "/docs/a.xml"},
		{"necessaryEscaped", args{"//docs/!\".xml"}, "/docs/%21%22.xml"},
		{"unecessaryEscaped", args{"//docs/%41.xml"}, "/docs/A.xml"},
		{"endslash", args{"/docs/a.xml/"}, "/docs/a.xml"},
		{"empty", args{""}, ""},
		{"onlyslash", args{"/"}, "/"},
		{"invalidURL", args{"/docs%/a.xml"}, "/docs%/a.xml"},
		{"abs", args{"http://a.com/docs/a.xml"}, "http://a.com/docs/a.xml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePartName(tt.args.name)
			if got != tt.want {
				t.Errorf("NormalizePartName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *Part
		wantErr bool
	}{
		{"base", &Part{"/docs/a.xml", "a/b", nil}, false},
		{"emptyName", &Part{"", "a/b", nil}, true},
		{"onlyspaces", &Part{"  ", "a/b", nil}, true},
		{"onlyslash", &Part{"/", "a/b", nil}, true},
		{"invalidURL", &Part{"/docs%/a.xml", "a/b", nil}, true},
		{"emptySegment", &Part{"/doc//a.xml", "a/b", nil}, true},
		{"abs uri", &Part{"http://docs//a.xml", "a/b", nil}, true},
		{"not rel uri", &Part{"docs/a.xml", "a/b", nil}, true},
		{"endSlash", &Part{"/docs/a.xml/", "a/b", nil}, true},
		{"endDot", &Part{"/docs/a.xml.", "a/b", nil}, true},
		{"dot", &Part{"/docs/./a.xml", "a/b", nil}, true},
		{"twoDots", &Part{"/docs/../a.xml", "a/b", nil}, true},
		{"reserved", &Part{"/docs/%7E/a.xml", "a/b", nil}, true},
		{"withQuery", &Part{"/docs/a.xml?a=2", "a/b", nil}, true},
		{"notencodechar", &Part{"/€/a.xml", "a/b", nil}, true},
		{"encodedBSlash", &Part{"/%5C/a.xml", "a/b", nil}, true},
		{"encodedBSlash", &Part{"/%2F/a.xml", "a/b", nil}, true},
		{"encodechar", &Part{"/%E2%82%AC/a.xml", "a/b", nil}, false},
		{"invalidMediaParams", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4 q=2", nil}, true},
		{"mediaParamNoName", &Part{"/a.txt", "TEXT/html; =ISO-8859-4", nil}, true},
		{"duplicateParamName", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4; charset=ISO-8859-4", nil}, true},
		{"linearSpace", &Part{"/a.txt", "TEXT /html; charset=ISO-8859-4;q=2", nil}, true},
		{"noSlash", &Part{"/a.txt", "application", nil}, true},
		{"unexpectedContent", &Part{"/a.txt", "application/html/html", nil}, true},
		{"noMediaType", &Part{"/a.txt", "/html", nil}, true},
		{"unexpectedToken", &Part{"/a.txt", "application/", nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Part.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
