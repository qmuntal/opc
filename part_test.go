package gopc

import (
	"reflect"
	"testing"
)

func Test_newPart(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		args    args
		want    *Part
		wantErr bool
	}{
		{"newPart", args{"fakeUri"}, &Part{"fakeUri", nil}, false},
		{"incorrectURI", args{""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPart(tt.args.uri)
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

func TestPart_HasRelationship(t *testing.T) {
	tests := []struct {
		name string
		p    *Part
		want bool
	}{
		{"partRelationshipTrue", &Part{"fakeUri", []*Relationship{&Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}}}, true},
		{"partRelationshipFalse", &Part{"fakeUri", nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.HasRelationship(); got != tt.want {
				t.Errorf("Part.HasRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_Relationships(t *testing.T) {
	tests := []struct {
		name string
		p    *Part
		want []*Relationship
	}{
		{"partRelationship", &Part{"fakeUri", []*Relationship{&Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}}}, []*Relationship{&Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Relationships(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Part.Relationships() = %v, want %v", got, tt.want)
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
		{"partURI", &Part{"fakeUri", nil}, "fakeUri"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.URI(); got != tt.want {
				t.Errorf("Part.URI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPart_AddRelationship(t *testing.T) {
	type args struct {
		id      string
		reltype string
		uri     string
	}
	tests := []struct {
		name    string
		p       *Part
		args    args
		want    *Part
		wantErr bool
	}{
		{"newRelationship", &Part{"fakeUri", nil}, args{"fakeId", "fakeType", "fakeTarget"}, &Part{"fakeUri", []*Relationship{&Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}}}, false},
		{"existingID", &Part{"fakeUri", []*Relationship{&Relationship{"fakeId", "fakeType", "fakeTarget", ModeInternal}}}, args{"fakeId", "fakeType", "fakeTarget"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.AddRelationship(tt.args.id, tt.args.reltype, tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("Part.AddRelationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Part.AddRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}
