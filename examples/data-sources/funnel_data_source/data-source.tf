# Copy a data source from one workspace to another
data "funnel_data_source" "source" {
  workspace = var.source_workspace_id
  id        = var.source_data_source_id
}

resource "funnel_data_source" "copy" {
  workspace     = var.target_workspace_id
  type          = data.funnel_data_source.source.type
  name          = data.funnel_data_source.source.name
  template_id   = data.funnel_data_source.source.template_id
  credential_id = data.funnel_data_source.source.credential_id
  remote_id     = data.funnel_data_source.source.remote_id
  remote_struct = data.funnel_data_source.source.remote_struct
}
