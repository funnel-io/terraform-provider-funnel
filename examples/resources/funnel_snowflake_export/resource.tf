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

# Basic Snowflake export
resource "funnel_snowflake_export" "basic" {
  workspace = var.workspace_id
  name      = "Daily Marketing Data to Snowflake"
  enabled   = true
  schedule  = "0 4 * * *" # Daily at 4 AM

  destination {
    account_locator       = "xy12345.us-east-1"
    database              = "MARKETING_DB"
    schema_name           = "FUNNEL_DATA"
    table_name            = "DAILY_PERFORMANCE"
    username              = "funnel_user"
    personal_access_token = var.snowflake_pat
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
