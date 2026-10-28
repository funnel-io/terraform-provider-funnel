package common

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// Converts a JSON attribute to raw JSON for the API. Null or unknown becomes nil so it is omitted from the request.
func NormalizedToRawJSON(v jsontypes.Normalized) json.RawMessage {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return json.RawMessage(v.ValueString())
}

// Converts raw JSON from the API to a JSON attribute. Missing or `null` becomes a null attribute.
func RawJSONToNormalized(raw json.RawMessage) jsontypes.Normalized {
	if len(raw) == 0 || string(raw) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}
