provider "funnel" {
  environment     = "us"
  subscription_id = "your_subscription_id_here"
  client_id       = var.client_id
  client_secret   = var.client_secret
}

resource "funnel_workspace" "example" {
  name = "My Workspace"
}
