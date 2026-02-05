package common

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TFStruct struct {
	Name      types.String `tfsdk:"name"`
	Boolean   types.Bool   `tfsdk:"boolean"`
	BigNumber types.Int64  `tfsdk:"big_number"`
	Object    TFFormat     `tfsdk:"object"`
	Array     []TFFormat   `tfsdk:"array"`
}

type TFFormat struct {
	Type    types.String `tfsdk:"type"`
	Metrics types.String `tfsdk:"metrics"`
}

type JSONStruct struct {
	Name      string       `json:"name"`
	Boolean   bool         `json:"boolean"`
	BigNumber int64        `json:"big_number"`
	Object    JSONFormat   `json:"object"`
	Array     []JSONFormat `json:"array"`
}

type JSONFormat struct {
	Type    string `json:"type"`
	Metrics string `json:"metrics"`
}

func TestConvertTFToJSON(t *testing.T) {
	tfFields := TFStruct{
		Name:      types.StringValue("generic_field"),
		Boolean:   types.BoolValue(true),
		BigNumber: types.Int64Value(123),
		Object: TFFormat{
			Type:    types.StringValue("json"),
			Metrics: types.StringValue("raw"),
		},
		Array: []TFFormat{
			{
				Type:    types.StringValue("json"),
				Metrics: types.StringValue("raw"),
			},
		},
	}

	output, err := ConvertTFToJSON[TFStruct, JSONStruct](tfFields)
	if err != nil {
		t.Errorf("Coud not marshal to other struct: %v", err)
	}

	body, err := json.Marshal(output)
	if err != nil {
		t.Errorf("Coud not marshal to JSON: %v", err)
	}

	if string(body) != "{\"name\":\"generic_field\",\"boolean\":true,\"big_number\":123,\"object\":{\"type\":\"json\",\"metrics\":\"raw\"},\"array\":[{\"type\":\"json\",\"metrics\":\"raw\"}]}" {
		t.Errorf("Unexpected JSON output: %s", string(body))
	}
}

func TestConvertJSONToTF(t *testing.T) {
	jsonInput := "{\"name\":\"temp_field\",\"boolean\":false,\"big_number\":321,\"object\":{\"type\":\"csv\",\"metrics\":\"pure\"},\"array\":[{\"type\":\"tsv\",\"metrics\":\"rawer\"}]}"

	var output JSONStruct
	err := json.Unmarshal([]byte(jsonInput), &output)
	if err != nil {
		t.Errorf("Coud not unmarshal JSON: %v", err)
	}

	tfFields, err := ConvertJSONToTF[JSONStruct, TFStruct](output)
	if err != nil {
		t.Errorf("Coud not marshal to other struct: %v", err)
	}

	if !reflect.DeepEqual(tfFields, TFStruct{
		Name:      types.StringValue("temp_field"),
		Boolean:   types.BoolValue(false),
		BigNumber: types.Int64Value(321),
		Object: TFFormat{
			Type:    types.StringValue("csv"),
			Metrics: types.StringValue("pure"),
		},
		Array: []TFFormat{
			{
				Type:    types.StringValue("tsv"),
				Metrics: types.StringValue("rawer"),
			},
		},
	}) {
		t.Errorf("Unexpected TF struct: %+v", tfFields)
	}
}
