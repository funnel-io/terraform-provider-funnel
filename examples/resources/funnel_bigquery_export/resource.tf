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
  schedule  = "CRON_TZ=UTC 0 3 * * *" # Daily at 3 AM (UTC)

  destination = {
    project_id          = "my-gcp-project"
    dataset_id          = "funnel_marketing_data"
    table_name_template = "funnel_export_{workspaceId}"
    # If you don't provide a "credential_id" a Funnel service account will be used.
    credential_id = var.bigquery_credential_id
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
