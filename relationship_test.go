package gopc

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestNewRelationship(t *testing.T) {
	type args struct {
		id            string
		relType       string
		targetPartURI string
	}
	tests := []struct {
		name    string
		args    args
		want    *Relationship
		wantErr bool
	}{
		{"new", args{"fakeId", "fakeType", "fakeTarget"}, &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, false},
		{"invalidTarget", args{"fakeId", "fakeType", ""}, nil, true},
		{"invalidTarget2", args{"fakeId", "fakeType", "."}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRelationship(tt.args.id, tt.args.relType, tt.args.targetPartURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRelationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_ID(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.ID(); got != tt.want {
				t.Errorf("Relationship.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_Type(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Type(); got != tt.want {
				t.Errorf("Relationship.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_TargetPartURI(t *testing.T) {
	tests := []struct {
		name string
		r    *Relationship
		want string
	}{
		{"id", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, "fakeTarget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.TargetPartURI(); got != tt.want {
				t.Errorf("Relationship.TargetPartURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelationship_WriteToXML(t *testing.T) {
	tests := []struct {
		name    string
		r       *Relationship
		want    string
		wantErr bool
	}{
		{"xmlWriter", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
		{"emptyURI", &Relationship{"fakeId", "fakeType", "", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/"></Relationship>`, false},
		{"externalMode", &Relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, `<Relationship Id="fakeId" Type="fakeType" Target="fakeTarget" TargetMode="External"></Relationship>`, false},
		{"base", &Relationship{"fakeId", "fakeType", "/fakeTarget", ModeInternal}, `<Relationship Id="fakeId" Type="fakeType" Target="/fakeTarget"></Relationship>`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.NewBufferString("")
			encoder := xml.NewEncoder(buff)
			if err := tt.r.WriteToXML(encoder); (err != nil) != tt.wantErr {
				t.Errorf("Relationship.WriteToXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := buff.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Relationship.WriteToXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRelationshipMode(t *testing.T) {
	type args struct {
		id            string
		relType       string
		targetPartURI string
		mode          TargetMode
	}
	tests := []struct {
		name    string
		args    args
		want    *Relationship
		wantErr bool
	}{
		{"new", args{"fakeId", "fakeType", "fakeTarget", ModeExternal}, &Relationship{"fakeId", "fakeType", "fakeTarget", ModeExternal}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRelationshipMode(tt.args.id, tt.args.relType, tt.args.targetPartURI, tt.args.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRelationshipMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRelationshipMode() = %v, want %v", got, tt.want)
			}
		})
	}
}