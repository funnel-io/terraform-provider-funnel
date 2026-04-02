package resources

import (
	"testing"
)

func TestJSONSemanticEquals(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical JSON",
			a:        `{"key": "value"}`,
			b:        `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "different key order",
			a:        `{"a": 1, "b": 2, "c": 3}`,
			b:        `{"c": 3, "a": 1, "b": 2}`,
			expected: true,
		},
		{
			name:     "different array order",
			a:        `{"items": [1, 2, 3]}`,
			b:        `{"items": [3, 2, 1]}`,
			expected: true,
		},
		{
			name:     "nested objects with different key order",
			a:        `{"outer": {"a": 1, "b": 2}}`,
			b:        `{"outer": {"b": 2, "a": 1}}`,
			expected: true,
		},
		{
			name:     "array of objects with different order",
			a:        `{"list": [{"id": 1, "name": "a"}, {"id": 2, "name": "b"}]}`,
			b:        `{"list": [{"id": 2, "name": "b"}, {"id": 1, "name": "a"}]}`,
			expected: true,
		},
		{
			name:     "different values",
			a:        `{"key": "value1"}`,
			b:        `{"key": "value2"}`,
			expected: false,
		},
		{
			name:     "different structure",
			a:        `{"key": "value"}`,
			b:        `{"key": ["value"]}`,
			expected: false,
		},
		{
			name:     "different array length",
			a:        `{"items": [1, 2, 3]}`,
			b:        `{"items": [1, 2]}`,
			expected: false,
		},
		{
			name:     "missing key",
			a:        `{"a": 1, "b": 2}`,
			b:        `{"a": 1}`,
			expected: false,
		},
		{
			name:     "complex nested example",
			a:        `{"config": {"fields": [{"name": "field1", "type": "string"}, {"name": "field2", "type": "int"}], "enabled": true}}`,
			b:        `{"config": {"enabled": true, "fields": [{"type": "int", "name": "field2"}, {"type": "string", "name": "field1"}]}}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonSemanticEquals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("jsonSemanticEquals(%q, %q) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
