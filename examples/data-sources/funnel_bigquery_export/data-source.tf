# Copy a BigQuery export from one workspace to another
data "funnel_bigquery_export" "source" {
  workspace = var.source_workspace_id
  id        = var.source_export_id
}

resource "funnel_bigquery_export" "copy" {
  workspace    = var.target_workspace_id
  name         = data.funnel_bigquery_export.source.name
  schedule     = data.funnel_bigquery_export.source.schedule
  notes        = data.funnel_bigquery_export.source.notes
  currency     = data.funnel_bigquery_export.source.currency
  fields       = data.funnel_bigquery_export.source.fields
  filters      = data.funnel_bigquery_export.source.filters
  range        = data.funnel_bigquery_export.source.range
  partition_by = data.funnel_bigquery_export.source.partition_by
  destination  = data.funnel_bigquery_export.source.destination
}
