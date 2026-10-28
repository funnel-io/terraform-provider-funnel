# Number type custom metric with SUM aggregation
resource "funnel_custom_metric" "total_leads" {
  workspace   = var.workspace_id
  name        = "Total Leads"
  description = "Sum of all leads generated across campaigns"
  unit        = "number"
  aggregation = "SUM"
  precision   = 0
}

# Monetary type custom metric with precision
resource "funnel_custom_metric" "revenue_per_click" {
  workspace   = var.workspace_id
  name        = "Revenue Per Click"
  description = "Average revenue generated per click"
  unit        = "monetary"
  aggregation = "NONE"
  precision   = 2
}

# Percent type custom metric
resource "funnel_custom_metric" "conversion_rate" {
  workspace   = var.workspace_id
  name        = "Conversion Rate"
  description = "Percentage of visitors who convert"
  unit        = "percent"
  aggregation = "NONE"
  precision   = 2
}

# Duration type custom metric with MAX aggregation
resource "funnel_custom_metric" "max_session_duration" {
  workspace   = var.workspace_id
  name        = "Maximum Session Duration"
  description = "Longest user session duration in seconds"
  unit        = "duration"
  aggregation = "MAX"
  precision   = 0
}

# COUNT aggregation example
resource "funnel_custom_metric" "unique_campaigns" {
  workspace   = var.workspace_id
  name        = "Unique Campaign Count"
  description = "Count of unique campaigns"
  unit        = "number"
  aggregation = "COUNT"
  precision   = 0
}

# Metric with rules, here copied from the Funnel UI. Built-in field IDs like "adwords-conversions" work in any workspace.
resource "funnel_custom_metric" "conversions_by_source" {
  workspace   = var.workspace_id
  name        = "Conversions By Source"
  description = "Google Ads conversions for Google Ads, regular conversions for everything else"
  unit        = "number"
  aggregation = "SUM"
  derived = jsonencode({
    expr = {
      "=case" = [
        {
          "=when" = { sourceType = { "=eq" = "adwords" } }
          "=then" = {
            "=case" = [{
              "=when" = true
              "=then" = { "=value" = [{ "=get" = "adwords-conversions" }] }
            }]
          }
        },
        {
          "=when" = true
          "=then" = {
            "=case" = [{
              "=when" = true
              "=then" = { "=value" = [{ "=get" = "conversions" }] }
            }]
          }
        },
      ]
    }
  })
}

# Rules exported from the Funnel UI and kept in a file
resource "funnel_custom_metric" "from_file" {
  workspace   = var.workspace_id
  name        = "Blended Conversions"
  description = "Rules maintained in the Funnel UI"
  unit        = "number"
  aggregation = "SUM"
  derived     = file("${path.module}/blended_conversions.json")
}
