# String type custom dimension
resource "funnel_custom_dimension" "campaign_type" {
  workspace   = var.workspace_id
  name        = "Campaign Type"
  description = "Categorizes campaigns by their marketing strategy type"
  unit        = "string"
}

# Date type custom dimension
resource "funnel_custom_dimension" "launch_date" {
  workspace   = var.workspace_id
  name        = "Campaign Launch Date"
  description = "The date when the campaign was first launched"
  unit        = "date"
}

# DateTime type custom dimension
resource "funnel_custom_dimension" "last_updated" {
  workspace   = var.workspace_id
  name        = "Last Updated Timestamp"
  description = "Timestamp of the last data update"
  unit        = "datetime"
}
