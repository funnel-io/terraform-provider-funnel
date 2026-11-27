# Look up the fields to export
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

# Basic Snowflake export
resource "funnel_snowflake_export" "basic" {
  workspace = var.workspace_id
  name      = "Daily Marketing Data to Snowflake"
  enabled   = true
  schedule  = "CRON_TZ=UTC 0 4 * * *" # Daily at 4 AM (UTC)

  destination = {
    account_locator = "myorg-myaccount"
    database        = "MARKETING_DB"
    schema_name     = "FUNNEL_DATA"
    table_name      = "DAILY_PERFORMANCE"
    credential_id   = var.snowflake_credential_id
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
