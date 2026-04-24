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

# Basic BigQuery export
resource "funnel_bigquery_export" "basic" {
  workspace = var.workspace_id
  name      = "Daily Marketing Data to BigQuery"
  enabled   = true
  schedule  = "0 3 * * *" # Daily at 3 AM

  destination {
    project_id         = "my-gcp-project"
    dataset_id         = "funnel_marketing_data"
    output_id_template = "daily_export_{date}"
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
