package common

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestRawJSONToNormalizedNull(t *testing.T) {
	for _, raw := range []json.RawMessage{nil, json.RawMessage("null")} {
		if !RawJSONToNormalized(raw).IsNull() {
			t.Errorf("expected null for %q", raw)
		}
	}
}

func TestNormalizedToRawJSONOmitsNull(t *testing.T) {
	payload := struct {
		Derived json.RawMessage `json:"derived,omitempty"`
	}{Derived: NormalizedToRawJSON(jsontypes.NewNormalizedNull())}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "{}" {
		t.Errorf("expected derived to be omitted, got %s", body)
	}
}

// Key order from the API differs from jsonencode, but `=case` branch order is meaningful.
func TestDerivedSemanticEquality(t *testing.T) {
	ctx := context.Background()
	config := jsontypes.NewNormalizedValue(`{"expr":{"=case":[{"=then":"a","=when":true},{"=then":"b","=when":true}]}}`)

	reorderedKeys := RawJSONToNormalized(json.RawMessage(`{"expr": {"=case": [{"=when": true, "=then": "a"}, {"=when": true, "=then": "b"}]}}`))
	if equal, _ := config.StringSemanticEquals(ctx, reorderedKeys); !equal {
		t.Error("expected reordered keys to be equal")
	}

	reorderedBranches := RawJSONToNormalized(json.RawMessage(`{"expr":{"=case":[{"=then":"b","=when":true},{"=then":"a","=when":true}]}}`))
	if equal, _ := config.StringSemanticEquals(ctx, reorderedBranches); equal {
		t.Error("expected reordered branches to differ")
	}
}
