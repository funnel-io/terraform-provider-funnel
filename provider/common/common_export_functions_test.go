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

func TestNormalizeFiltersNested(t *testing.T) {
	read := func() []ExportFilter {
		return []ExportFilter{{
			FieldId:   types.StringValue("f"),
			Operation: types.StringValue(""),
			Value:     types.StringValue(""),
			Or: []struct {
				Operation types.String `tfsdk:"operation"`
				Value     types.String `tfsdk:"value"`
			}{},
		}}
	}

	got := NormalizeFilters(read(), []ExportFilter{{FieldId: types.StringValue("f")}})
	if got[0].Or != nil || !got[0].Operation.IsNull() || !got[0].Value.IsNull() {
		t.Errorf("expected nested empties to be null, got %+v", got[0])
	}

	prior := read()
	got = NormalizeFilters(read(), prior)
	if got[0].Or == nil || got[0].Operation.IsNull() || got[0].Value.IsNull() {
		t.Errorf("expected explicit empties in prior to be kept, got %+v", got[0])
	}
}

func TestNormalizeFields(t *testing.T) {
	read := []ExportField{{
		Id:         types.StringValue("date"),
		Type:       types.StringValue("dimension"),
		ExportName: types.StringValue(""),
		ExportType: types.StringValue(""),
	}}

	got := NormalizeFields(read, nil)
	if !got[0].ExportName.IsNull() || !got[0].ExportType.IsNull() {
		t.Errorf("expected empty export_name/export_type to be null, got %+v", got[0])
	}
	if !got[0].Id.Equal(types.StringValue("date")) || !got[0].Type.Equal(types.StringValue("dimension")) {
		t.Errorf("expected id/type to be kept, got %+v", got[0])
	}
}
