resource "funnel_snowflake_export" "example" {
  name = "example_snowflake_export"
  workspace = "workspace_id"
  schedule = "0 12 * * *"

  destination {
    account_locator = "account_locator"
    table_name = "table_name"
    database = "database"
    schema_name = "schema_name"
    username = "username"
    personal_access_token = var.personal_access_token
  }
}
