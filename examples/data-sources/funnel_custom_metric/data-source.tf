# Copy a metric built in the Funnel UI from one workspace to another
data "funnel_custom_metric" "source" {
  workspace = var.source_workspace_id
  id        = var.source_metric_id
}

resource "funnel_custom_metric" "copy" {
  workspace   = var.target_workspace_id
  name        = data.funnel_custom_metric.source.name
  description = data.funnel_custom_metric.source.description
  unit        = data.funnel_custom_metric.source.unit
  aggregation = data.funnel_custom_metric.source.aggregation
  precision   = data.funnel_custom_metric.source.precision
  derived     = data.funnel_custom_metric.source.derived
}
