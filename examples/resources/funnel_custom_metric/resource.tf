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
