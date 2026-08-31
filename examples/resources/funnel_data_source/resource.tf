# Google analytics data source with remoteStruct contract
resource "funnel_data_source" "google_analytics" {
  workspace     = var.workspace_id
  type          = "googleanalytics"
  name          = "Google Analytics - With parent"
  template_id   = "funnel:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  credential_id = var.google_analytics_credential_id
  remote_struct = jsonencode({
    remote_id        = "123-456-7890"
    remote_parent_id = "098-765-4321"
  })

  # Default values
  download_disabled        = false
  exclude_data_from_funnel = false
}

# TikTok Ads data source with remoteId contract
resource "funnel_data_source" "tiktok_ads" {
  workspace     = var.workspace_id
  type          = "tiktok"
  name          = "TikTok Ads - Campaign"
  template_id   = "funnel:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5"
  remote_id     = "7012345678901234567"
  credential_id = var.tiktok_credential_id
}

# Spotify Ads data source with authOnly contract (no remote identity)
resource "funnel_data_source" "spotify" {
  workspace     = var.workspace_id
  type          = "spotifyads"
  name          = "Spotify Ads"
  template_id   = "funnel:c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6"
  credential_id = var.spotify_credential_id
}

# Custom template example
resource "funnel_data_source" "tiktok_custom" {
  workspace     = var.workspace_id
  type          = "tiktok"
  name          = "TikTok Ads - Custom Template"
  template_id   = "tiktok-d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1"
  remote_id     = "7012345678901234567"
  credential_id = var.tiktok_credential_id
}
