# Basic field lookup without overrides
data "funnel_export_field" "date" {
  workspace = var.workspace_id
  id        = "date"
}

# Field lookup for common metrics
data "funnel_export_field" "impressions" {
  workspace = var.workspace_id
  id        = "impressions"
}

# Field with custom export name override
data "funnel_export_field" "campaign_name_custom" {
  workspace   = var.workspace_id
  id          = "campaign_name"
  export_name = "campaign"
}

# Field with custom export type override
data "funnel_export_field" "impressions_as_int" {
  workspace   = var.workspace_id
  id          = "impressions"
  export_name = "impression_count"
  export_type = "INTEGER"
}

# Field with both name and type overrides
data "funnel_export_field" "cost_custom" {
  workspace   = var.workspace_id
  id          = "cost"
  export_name = "total_spend"
  export_type = "FLOAT"
}
