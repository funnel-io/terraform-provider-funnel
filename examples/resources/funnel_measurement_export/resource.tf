# Fetch export fields
data "funnel_export_field" "date" {
  workspace = var.workspace_id
  id        = "date"
}

data "funnel_export_field" "campaign_name" {
  workspace = var.workspace_id
  id        = "campaign_name"
}

data "funnel_export_field" "impressions" {
  workspace = var.workspace_id
  id        = "impressions"
}

data "funnel_export_field" "cost" {
  workspace = var.workspace_id
  id        = "cost"
}

# Basic Measurement export (without snapshot)
# Note: This resource is for internal Funnel Measurement Consultants only
resource "funnel_measurement_export" "basic" {
  workspace = var.workspace_id
  name      = "Daily Measurement Data"
  enabled   = true
  schedule  = "0 5 * * *" # Daily at 5 AM

  destination {
    table_name = "measurement_daily_performance"
  }

  fields = [
    data.funnel_export_field.date,
    data.funnel_export_field.campaign_name,
    data.funnel_export_field.impressions,
    data.funnel_export_field.cost
  ]

  format {
    type    = "parquet"
    metrics = "export"
  }

  range {
    rolling_start {
      period  = "days"
      periods = -7
    }
    rolling_end {
      period  = "days"
      periods = -1
    }
  }
}
