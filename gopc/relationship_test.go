package gopc

import (
	"reflect"
	"testing"
)

func TestNewRelationship(t *testing.T) {
	tests := []struct {
		name string
		want *Relationship
	}{
		{"new", new(Relationship)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRelationship(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRelationship() = %v, want %v", got, tt.want)
			}
		})
	}
}
