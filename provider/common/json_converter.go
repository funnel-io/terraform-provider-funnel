package common

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jinzhu/copier"
)

// Converts a Terraform SDK struct to a JSON struct. Handles nested structures automatically by recursively copying fields.
// Only supports types.String, types.Int64 and types.Bool now.
func ConvertTFToJSON[TF any, JSON any](tf TF) (JSON, error) {
	var output JSON

	err := copier.CopyWithOption(&output, &tf, copier.Option{Converters: []copier.TypeConverter{
		{
			SrcType: types.String{},
			DstType: string(""),
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(types.String)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return s.ValueString(), nil
			},
		},
		{
			SrcType: types.Int64{},
			DstType: int64(0),
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(types.Int64)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return s.ValueInt64(), nil
			},
		},
		{
			SrcType: types.Bool{},
			DstType: bool(false),
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(types.Bool)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return s.ValueBool(), nil
			},
		},
	}})

	return output, err
}

// Converts a JSON struct to a Terraform SDK struct. Handles nested structures automatically by recursively copying fields.
// Only supports types.String, types.Int64 and types.Bool now.
func ConvertJSONToTF[JSON any, TF any](json JSON) (TF, error) {
	var output TF

	err := copier.CopyWithOption(&output, &json, copier.Option{Converters: []copier.TypeConverter{
		{
			SrcType: string(""),
			DstType: types.String{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(string)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return types.StringValue(s), nil
			},
		},
		{
			SrcType: int64(0),
			DstType: types.Int64{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(int64)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return types.Int64Value(s), nil
			},
		},
		{
			SrcType: bool(false),
			DstType: types.Bool{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(bool)

				if !ok {
					return nil, errors.New("src type not matching")
				}

				return types.BoolValue(s), nil
			},
		},
	}})

	return output, err
}
