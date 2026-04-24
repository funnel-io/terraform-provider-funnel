# Look up an existing workspace by ID
data "funnel_workspace" "existing" {
  id = "ws_abc123xyz"
}

# Output workspace information
output "workspace_info" {
  description = "Information about the existing workspace"
  value = {
    id          = data.funnel_workspace.existing.id
    name        = data.funnel_workspace.existing.name
    users_count = data.funnel_workspace.existing.users_count
    created_at  = data.funnel_workspace.existing.created_at
    updated_at  = data.funnel_workspace.existing.updated_at
  }
}
