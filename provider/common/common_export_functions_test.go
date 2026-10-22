package common

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeEmptyString(t *testing.T) {
	tests := []struct {
		name     string
		notes    types.String
		prior    types.String
		expected types.String
	}{
		{"empty with null prior", types.StringValue(""), types.StringNull(), types.StringNull()},
		{"empty with empty prior", types.StringValue(""), types.StringValue(""), types.StringValue("")},
		{"empty with removed notes", types.StringValue(""), types.StringValue("old"), types.StringNull()},
		{"value is kept", types.StringValue("hello"), types.StringNull(), types.StringValue("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeEmptyString(tt.notes, tt.prior); !got.Equal(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestNormalizeFilters(t *testing.T) {
	filter := []ExportFilter{{FieldId: types.StringValue("f")}}

	tests := []struct {
		name      string
		filters   []ExportFilter
		prior     []ExportFilter
		expectNil bool
	}{
		{"empty with null prior", []ExportFilter{}, nil, true},
		{"empty with empty prior", []ExportFilter{}, []ExportFilter{}, false},
		{"empty with removed filters", []ExportFilter{}, filter, true},
		{"value is kept", filter, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeFilters(tt.filters, tt.prior)
			if (got == nil) != tt.expectNil {
				t.Errorf("expected nil=%v, got %v", tt.expectNil, got)
			}
		})
	}
}
