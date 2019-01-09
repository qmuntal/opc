package gopc

import (
	"reflect"
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
		{"mediaEmpty", &Part{"/a.txt", "", nil}, false},
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
		{"mediaSpaceStart", &Part{"/a.txt", " TEXT/html; charset=ISO-8859-4;q=2", nil}, true},
		{"mediaSpaceEnd", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4;q=2 ", nil}, true},
		{"mediaLinearStart", &Part{"/a.txt", "/tTEXT/html; charset=ISO-8859-4;q=2", nil}, true},
		{"mediaLinearEnd", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4;q=2/t", nil}, true},
		{"invalidMediaParams", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4 q=2", nil}, true},
		{"mediaParamNoName", &Part{"/a.txt", "TEXT/html; =ISO-8859-4", nil}, true},
		{"duplicateParamName", &Part{"/a.txt", "TEXT/html; charset=ISO-8859-4; charset=ISO-8859-4", nil}, true},
		{"linearSpace", &Part{"/a.txt", "TEXT/t/html; charset=ISO-8859-4;q=2", nil}, true},
		{"whiteSpace", &Part{"/a.txt", "TEXT /html; charset=ISO-8859-4;q=2", nil}, true},
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

func TestPart_CreateRelationship(t *testing.T) {
	type args struct {
		ID         string
		targetURI  string
		relType    string
		targetMode TargetMode
	}
	tests := []struct {
		name string
		p    *Part
		args args
		want *Relationship
	}{
		{"base", &Part{Name: "/a.xml"}, args{"rel0", "/b.xml", "fake", ModeInternal}, &Relationship{ID: "rel0", RelType: "fake", TargetURI: "/b.xml", sourceURI: "/a.xml", TargetMode: ModeInternal}},
		{"noid", &Part{Name: "/a.xml"}, args{"", "/b.xml", "fake", ModeInternal}, &Relationship{ID: "", RelType: "fake", TargetURI: "/b.xml", sourceURI: "/a.xml", TargetMode: ModeInternal}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.CreateRelationship(tt.args.ID, tt.args.targetURI, tt.args.relType, tt.args.targetMode)
			if tt.args.ID != "" {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Part.CreateRelationship() = %v, want %v", got, tt.want)
					return
				}
			} else {
				if got == nil {
					t.Error("Part.CreateRelationship() got nit relationship")
					return
				}
			}
			if !reflect.DeepEqual(got, tt.p.Relationships[0]) {
				t.Errorf("Part.CreateRelationship() = %v, want %v", got, tt.p.Relationships[0])
			}
		})
	}
}
