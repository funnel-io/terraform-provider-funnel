# Import a metric set up in the Funnel UI. Run `terraform plan -generate-config-out=generated.tf` to have Terraform write its configuration, including `derived`.
import {
  to = funnel_custom_metric.example
  id = "<workspace_id>/<custom_metric_id>"
}
