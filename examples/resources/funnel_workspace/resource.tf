# Create a new workspace
resource "funnel_workspace" "example" {
  name = "My Marketing Workspace"
}

# Output the workspace ID for use in other resources
output "workspace_id" {
  value = funnel_workspace.example.id
}
