package opc

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
		{"empty", args{"   "}, ""},
		{"noslash", args{"a.xml"}, "/a.xml"},
		{"endDot", args{"/a/."}, "/a"},
		{"folder", args{"/docs/a.xml"}, "/docs/a.xml"},
		{"noext", args{"/docs"}, "/docs"},
		{"win", args{"\\docs\\a.xml"}, "/docs/a.xml"},
		{"winnoslash", args{"docs\\a.xml"}, "/docs/a.xml"},
		{"fragment", args{"/docs/a.xml#a"}, "/docs/a.xml"},
		{"twoslash", args{"//docs/a.xml"}, "/docs/a.xml"},
		{"necessaryEscaped", args{"//docs/!\".xml"}, "/docs/!%22.xml"},
		{"unecessaryEscaped", args{"//docs/%41.xml"}, "/docs/A.xml"},
		{"endslash", args{"/docs/a.xml/"}, "/docs/a.xml"},
		{"empty", args{""}, ""},
		{"dot", args{"."}, ""},
		{"onlyslash", args{"/"}, ""},
		{"percentSign", args{"/docs%/a.xml"}, "/docs%25/a.xml"},
		{"percentSign2", args{"/docs%25/%41.xml"}, "/docs%25/A.xml"},
		{"percentSignEnd", args{"/docs/a.%"}, "/docs/a.%25"},
		{"pre-encoded", args{"/%3Aa.xml"}, "/%3Aa.xml"},
		{"pre-encodedMixedWithNecessaryEscaped", args{"/%28a a.xml"}, "/%28a%20a.xml"},
		{"chinese", args{"/传/傳.xml"}, "/%E4%BC%A0/%E5%82%B3.xml"},
		{"fromSpec1", args{"/a/b.xml"}, "/a/b.xml"},
		{"fromSpec2", args{"/a/ц.xml"}, "/a/%D1%86.xml"},
		{"fromSpec3", args{"/%41/%61.xml"}, "/A/a.xml"},
		{"fromSpec4", args{"/%25XY.xml"}, "/%25XY.xml"},
		{"fromSpec5", args{"/%XY.xml"}, "/%25XY.xml"},
		{"fromSpec6", args{"/%2541.xml"}, "/%2541.xml"},
		{"fromSpec7", args{"/../a.xml"}, "/a.xml"},
		{"fromSpec8", args{"/./ц.xml"}, "/%D1%86.xml"},
		{"fromSpec9", args{"/%2e/%2e/a.xml"}, "/a.xml"},
		{"fromSpec10", args{"\\a.xml"}, "/a.xml"},
		{"fromSpec11", args{"\\%41.xml"}, "/A.xml"},
		{"fromSpec12", args{"/%D1%86.xml"}, "/%D1%86.xml"},
		{"fromSpec13", args{"\\%2e/a.xml"}, "/a.xml"},
		{"unicode1", args{"/\uFFFDa.xml"}, "/%EF%BF%BDa.xml"},
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
		{"percentChar", &Part{"/docs%/a.xml", "a/b", nil}, false},
		{"ucschar", &Part{"/€/a.xml", "a/b", nil}, false},
		{"mediaEmpty", &Part{"/a.txt", "", nil}, true},
		{"emptyName", &Part{"", "a/b", nil}, true},
		{"onlyspaces", &Part{"  ", "a/b", nil}, true},
		{"onlyslash", &Part{"/", "a/b", nil}, true},
		{"emptySegment", &Part{"/doc//a.xml", "a/b", nil}, true},
		{"abs uri", &Part{"http://docs//a.xml", "a/b", nil}, true},
		{"not rel uri", &Part{"docs/a.xml", "a/b", nil}, true},
		{"encoded unreserved", &Part{"/%41.xml", "a/b", nil}, true},
		{"endSlash", &Part{"/docs/a.xml/", "a/b", nil}, true},
		{"endDot", &Part{"/docs/a.xml.", "a/b", nil}, true},
		{"dot", &Part{"/docs/./a.xml", "a/b", nil}, true},
		{"twoDots", &Part{"/docs/../a.xml", "a/b", nil}, true},
		{"reserved", &Part{"/docs/%7E/a.xml", "a/b", nil}, true},
		{"withQuery", &Part{"/docs/a.xml?a=2", "a/b", nil}, true},
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

func TestResolveRelationship(t *testing.T) {
	type args struct {
		source string
		rel    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"package", args{"/", "c.xml"}, "/c.xml"},
		{"packageWithSlash", args{"/", "/c.xml"}, "/c.xml"},
		{"packageWin", args{"\\", "c.xml"}, "/c.xml"},
		{"packageWinWithSlash", args{"\\", "\\c.xml"}, "/c.xml"},
		{"rel", args{"/3D/3dmodel.model", "c.xml"}, "/3D/c.xml"},
		{"rel", args{"/3D/3dmodel.model", "/3D/box1.model"}, "/3D/box1.model"},
		{"rel", args{"/3D/box3.model", "/2D/2dmodel.model"}, "/2D/2dmodel.model"},
		{"relChild", args{"/3D/box3.model", "2D/2dmodel.model"}, "/3D/2D/2dmodel.model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveRelationship(tt.args.source, tt.args.rel); got != tt.want {
				t.Errorf("ResolveRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}
