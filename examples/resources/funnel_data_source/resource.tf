# Data source with a report type
resource "funnel_data_source" "adwords_campaign" {
  workspace     = var.workspace_id
  type          = "adwords"
  name          = "Google Ads - Main Account"
  report_type   = "campaign"
  remote_id     = "12345678"
  credential_id = var.google_ads_credential_id

  # Default values
  download_disabled        = false
  exclude_data_from_funnel = false
}
