package common

import (
	"reflect"
	"testing"
)

func TestConvertFiltersToMeld_WithMixedFilters(t *testing.T) {
	filters := []ExportFilterJSON{
		{
			FieldId: "brand_field",
			Or: []ExportFilterOrJSON{
				{Operation: "contains", Value: "burger king"},
				{Operation: "notcontains", Value: "wendys"},
			},
		},
		{
			FieldId:   "date_field",
			Operation: "after",
			Value:     "2025",
		},
	}

	result := ConvertFiltersToMeld(filters)

	expected := map[string]any{
		"=and": []map[string]any{
			{
				"brand_field": map[string]any{
					"=or": []map[string]any{
						{"=contains": "burger king"},
						{"=notcontains": "wendys"},
					},
				},
			},
			{
				"date_field": map[string]any{
					"=after": "2025",
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestConvertFiltersToMeld_SingleFilter(t *testing.T) {
	filters := []ExportFilterJSON{
		{
			FieldId:   "status_field",
			Operation: "equals",
			Value:     "active",
		},
	}

	result := ConvertFiltersToMeld(filters)

	expected := map[string]any{
		"=and": []map[string]any{
			{
				"status_field": map[string]any{
					"=equals": "active",
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestConvertFiltersToMeld_EmptyFilters(t *testing.T) {
	filters := []ExportFilterJSON{}

	result := ConvertFiltersToMeld(filters)

	if result != nil {
		t.Errorf("Expected nil for empty filters, got %v", result)
	}
}

func TestConvertFiltersFromMeld_WithMixedFilters(t *testing.T) {
	filters := map[string]any{
		"=and": []any{
			map[string]any{
				"brand_field": map[string]any{
					"=or": []any{
						map[string]any{"=contains": "burger king"},
						map[string]any{"=notcontains": "wendys"},
					},
				},
			},
			map[string]any{
				"date_field": map[string]any{
					"=after": "2025",
				},
			},
		},
	}

	result := ConvertFiltersFromMeld(filters)

	expected := []ExportFilterJSON{
		{
			FieldId: "brand_field",
			Or: []ExportFilterOrJSON{
				{Operation: "contains", Value: "burger king"},
				{Operation: "notcontains", Value: "wendys"},
			},
		},
		{
			FieldId:   "date_field",
			Operation: "after",
			Value:     "2025",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestConvertFiltersFromMeld_SingleFilter(t *testing.T) {
	filters := map[string]any{
		"sourceKey": map[string]any{
			"=eq": "adwords:a0da9cc9-4140-4af6-9656-b6933aa5e52d",
		},
	}

	result := ConvertFiltersFromMeld(filters)

	expected := []ExportFilterJSON{
		{
			FieldId:   "sourceKey",
			Operation: "eq",
			Value:     "adwords:a0da9cc9-4140-4af6-9656-b6933aa5e52d",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestConvertFiltersFromMeld_EmptyFilters(t *testing.T) {
	filters := map[string]any{}

	result := ConvertFiltersFromMeld(filters)

	if len(result) != 0 {
		t.Errorf("Expected empty slice for empty filters, got %v", result)
	}
}
