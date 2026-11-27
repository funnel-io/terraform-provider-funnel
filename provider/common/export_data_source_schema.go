package common

import (
	"fmt"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Builds an export data source schema from the export resource schema, so both share attribute names and descriptions.
// Every attribute is computed except id and workspace, which identify the export to read.
func ExportDataSourceSchema(rs schema.Schema, description string) dsschema.Schema {
	attributes := toComputedAttributes(rs.Attributes)
	attributes["id"] = dsschema.StringAttribute{
		MarkdownDescription: "Export identifier.",
		Required:            true,
	}
	attributes["workspace"] = dsschema.StringAttribute{
		MarkdownDescription: "Funnel workspace ID the export belongs to.",
		Required:            true,
		Validators:          []validator.String{WorkspaceIdValidator()},
	}
	return dsschema.Schema{MarkdownDescription: description, Attributes: attributes}
}

func toComputedAttributes(attributes map[string]schema.Attribute) map[string]dsschema.Attribute {
	result := map[string]dsschema.Attribute{}
	for name, attribute := range attributes {
		result[name] = toComputedAttribute(attribute)
	}
	return result
}

func toComputedAttribute(attribute schema.Attribute) dsschema.Attribute {
	switch a := attribute.(type) {
	case schema.StringAttribute:
		return dsschema.StringAttribute{MarkdownDescription: a.MarkdownDescription, Sensitive: a.Sensitive, Computed: true}
	case schema.BoolAttribute:
		return dsschema.BoolAttribute{MarkdownDescription: a.MarkdownDescription, Sensitive: a.Sensitive, Computed: true}
	case schema.Int64Attribute:
		return dsschema.Int64Attribute{MarkdownDescription: a.MarkdownDescription, Sensitive: a.Sensitive, Computed: true}
	case schema.SingleNestedAttribute:
		return dsschema.SingleNestedAttribute{
			MarkdownDescription: a.MarkdownDescription,
			Attributes:          toComputedAttributes(a.Attributes),
			Computed:            true,
		}
	case schema.ListNestedAttribute:
		return dsschema.ListNestedAttribute{
			MarkdownDescription: a.MarkdownDescription,
			NestedObject:        dsschema.NestedAttributeObject{Attributes: toComputedAttributes(a.NestedObject.Attributes)},
			Computed:            true,
		}
	}
	panic(fmt.Sprintf("unsupported export attribute type %T", attribute))
}
