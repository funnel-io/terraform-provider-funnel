# Import a dimension set up in the Funnel UI. Run `terraform plan -generate-config-out=generated.tf` to have Terraform write its configuration, including `derived`.
import {
  to = funnel_custom_dimension.example
  id = "<workspace_id>/<custom_dimension_id>"
}
