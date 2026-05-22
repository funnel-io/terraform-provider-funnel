# Google Ads data source with campaign report type
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

# Facebook Ads data source with ad level report
resource "funnel_data_source" "facebook_ads" {
  workspace     = var.workspace_id
  type          = "facebookads"
  name          = "Facebook Ads - Main Account"
  report_type   = "ad"
  remote_id     = "act_123456789"
  credential_id = var.facebook_credential_id
}

# TikTok Ads data source with audience report
resource "funnel_data_source" "tiktok_audience" {
  workspace     = var.workspace_id
  type          = "tiktok"
  name          = "TikTok Ads - Audience Report"
  report_type   = "audience"
  remote_id     = "987654321"
  credential_id = var.tiktok_credential_id
}

# LinkedIn Ads data source (no report type needed)
resource "funnel_data_source" "linkedin" {
  workspace     = var.workspace_id
  type          = "linkedin_api"
  name          = "LinkedIn Ads"
  credential_id = var.linkedin_credential_id
}
