package planmodifiers

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type jsonSemanticEqualModifier struct{}

func JSONSemanticEqual() planmodifier.String {
	return jsonSemanticEqualModifier{}
}

func (m jsonSemanticEqualModifier) Description(ctx context.Context) string {
	return "Suppresses diff when JSON is semantically equivalent (ignores key and array order)"
}

func (m jsonSemanticEqualModifier) MarkdownDescription(ctx context.Context) string {
	return "Suppresses diff when JSON is semantically equivalent (ignores key and array order)"
}

func (m jsonSemanticEqualModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	if req.StateValue.IsUnknown() || req.PlanValue.IsUnknown() {
		return
	}

	stateJSON := req.StateValue.ValueString()
	planJSON := req.PlanValue.ValueString()

	if stateJSON == "" || planJSON == "" {
		return
	}

	if jsonSemanticEquals(stateJSON, planJSON) {
		resp.PlanValue = req.StateValue
	}
}

func jsonSemanticEquals(a, b string) bool {
	var objA, objB interface{} // nosemgrep: go.lang.security.deserialization.unsafe-deserialization-interface.go-unsafe-deserialization-interface

	err := json.Unmarshal([]byte(a), &objA)
	if err != nil {
		return false
	}

	err = json.Unmarshal([]byte(b), &objB)
	if err != nil {
		return false
	}

	return deepEqualIgnoreOrder(objA, objB)
}

func deepEqualIgnoreOrder(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		for key, aItem := range aVal {
			bItem, exists := bVal[key]
			if !exists {
				return false
			}
			if !deepEqualIgnoreOrder(aItem, bItem) {
				return false
			}
		}
		return true

	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}

		aSorted := sortArrayForComparison(aVal)
		bSorted := sortArrayForComparison(bVal)

		for i := range aSorted {
			if !deepEqualIgnoreOrder(aSorted[i], bSorted[i]) {
				return false
			}
		}
		return true

	default:
		return reflect.DeepEqual(a, b)
	}
}

func sortArrayForComparison(arr []interface{}) []interface{} {
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	sort.SliceStable(sorted, func(i, j int) bool {
		iJSON, _ := json.Marshal(sorted[i])
		jJSON, _ := json.Marshal(sorted[j])
		return string(iJSON) < string(jJSON)
	})

	return sorted
}
