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

# Dimension with rules, here copied from the Funnel UI. `=case` branches are evaluated in order.
resource "funnel_custom_dimension" "brand" {
  workspace   = var.workspace_id
  name        = "Brand"
  description = "Brand name for all data"
  unit        = "string"
  derived = jsonencode({
    expr = {
      "=case" = [{
        "=when" = true
        "=then" = {
          "=case" = [{
            "=when" = true
            "=then" = { "=value" = [{ "=string" = "Test Brand" }] }
          }]
        }
      }]
    }
  })
}

# Rules exported from the Funnel UI and kept in a file
resource "funnel_custom_dimension" "from_file" {
  workspace   = var.workspace_id
  name        = "Market"
  description = "Rules maintained in the Funnel UI"
  unit        = "string"
  derived     = file("${path.module}/market.json")
}
