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

# Basic GCS export with Parquet format
resource "funnel_gcs_export" "basic" {
  workspace = var.workspace_id
  name      = "Daily Marketing Data to GCS"
  enabled   = true
  schedule  = "CRON_TZ=UTC 0 2 * * *" # Daily at 2 AM (UTC)

  destination = {
    bucket             = "my-funnel-exports"
    path               = "marketing-data"
    output_id_template = "funnel_export_{date}"
    credentials_ref    = "gcs-service-account-key"
    file_format        = "parquet"
  }

  fields = [
    data.funnel_export_field.date,
    data.funnel_export_field.campaign_name,
    data.funnel_export_field.impressions,
    data.funnel_export_field.cost
  ]

  range = {
    end = "today"
    rolling_start = {
      period  = "month"
      periods = 3
    }
  }

  partition_by = "month"
}
