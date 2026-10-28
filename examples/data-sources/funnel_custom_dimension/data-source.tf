# Copy a dimension built in the Funnel UI from one workspace to another
data "funnel_custom_dimension" "source" {
  workspace = var.source_workspace_id
  id        = var.source_dimension_id
}

resource "funnel_custom_dimension" "copy" {
  workspace   = var.target_workspace_id
  name        = data.funnel_custom_dimension.source.name
  description = data.funnel_custom_dimension.source.description
  unit        = data.funnel_custom_dimension.source.unit
  derived     = data.funnel_custom_dimension.source.derived
}
