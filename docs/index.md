---
layout: ""
page_title: "Provider: Funnel"
description: |-
  The Funnel provider provides configuration management resources for Funnel.
---

# Funnel Provider

~> **Beta:** The Funnel provider is in beta. Resources and schemas may change between releases, so pin the provider version and read the [changelog](https://github.com/funnel-io/terraform-provider-funnel/blob/main/CHANGELOG.md) before upgrading.

-> **Funnel as Code:** Read the [Funnel as Code](https://help.funnel.io/en/articles/14884773-funnel-as-code) knowledge base article for an introduction to managing Funnel with Terraform.

## Example Usage

### Configuring the provider

```terraform
provider "funnel" {
  environment     = "us"
  subscription_id = "your_subscription_id_here"
  client_id       = var.client_id
  client_secret   = var.client_secret
}

resource "funnel_workspace" "example" {
  name = "My Workspace"
}
```

## Authentication

The provider authenticates with the `client_id` and `client_secret` of a Funnel System user.
In Funnel go to "Subscription overview" in the right top corner, select "Authentication" and create a new System user at the bottom of the page.

## Subscriptions, workspaces and resource IDs

The Funnel provider requires a `subscription_id` to be set in the provider configuration. Yours can be found in the URL if you go into the Subscription overview in the Funnel app.
The URL will look something like this: `https://app.funnel.io/subscription/SUBSCRIPTION_ID/overview`.

Funnel workspaces are managed with the `funnel_workspace` resource.
To bring an existing workspace under Terraform management, get its ID from the URL and import it with `terraform import funnel_workspace.<name> <workspace_id>`. The URL looks something like this: `https://app.funnel.io/workspaces/WORKSPACE_ID`.

Get the IDs of existing Funnel objects from the Funnel app as well. Above the table listing of resources there is a "Choose column" button where you can add the column "ID" that contains the resource IDs.
You can also open the object you want to manage with Terraform and get the ID from the URL. The URL will look something like this: `https://app.funnel.io/workspaces/WORKSPACE_ID/<object path>/OBJECT_ID`. For a BigQuery export it's `https://app.funnel.io/workspaces/WORKSPACE_ID/destinations/bigquery/EXPORT_ID`.

## Exports

Each export destination has its own resource type: `funnel_bigquery_export`, `funnel_snowflake_export`, and `funnel_gcs_export`. All arguments except the `destination` attribute are shared between the export resources.

## Fields

Use the `funnel_export_field` data source to look up a field for an export. Reference these data sources in the `fields` attribute of the export resources.

If `export_name` is not set, it defaults to the field's name in Funnel, here `Date`:
```hcl
data "funnel_export_field" "date_field" {
  workspace   = local.workspace_id
  id          = local.field_date_id
}
```

Set `export_name` to override the column name in the export:
```hcl
data "funnel_export_field" "adtriba_brand_field" {
  workspace   = local.workspace_id
  id          = local.field_adtriba_brand_id
  export_name = "TF_ADTRIBA_BRAND"
}
```

Set both `export_name` and `export_type` to override the column name and type in the export:
```hcl
data "funnel_export_field" "clicks_field" {
  workspace   = local.workspace_id
  id          = local.field_clicks_id
  export_name = "TF_CLICKS"
  export_type = "BIGINT"
}
```

## Filters

Filters are used to filter the data that is exported. They are set in the `filters` attribute of the export resources.
Example of an AND filter of three filters on three fields:

```hcl
filters = [
    {
      field_id = local.field_adtriba_brand_id
      or = [
        {
          operation = "contains"
          value     = "company_a"
        },
        {
          operation = "contains"
          value     = "company_b"
      }]
    },
    {
      field_id  = local.field_clicks_id
      operation = "gt"
      value     = "100"
    },
    {
      field_id  = "sourceType"
      operation = "eq"
      value     = "adwords"
    }
  ]
```

All entries in `filters` must match for a row to be exported. The example above exports rows where:

1. The Adtriba brand contains `company_a` **or** `company_b`.
2. Clicks are greater than `100`.
3. The source type is `adwords` (Google Ads).

## Potential errors

### Conflicting resources

If `terraform apply` fails with a `409 Conflict` error, the export conflicts with an existing one.
The error message says which export it conflicts with. For a BigQuery export an error would be caused by two exports targeting the same dataset and table.

### Flexpoints needed

If `terraform apply` fails with a `402 Payment Required` error, your subscription does not have enough Flexpoints to create the resource.
Go into the Funnel app and into the Flexpoints Usage in the Subscription overview to see how many Flexpoints you have and what resources consume Flexpoints.

### Outdated provider version

If you receive an error like the one below when running `terraform plan` or `terraform apply`, upgrade the provider to the latest version by updating its version constraint and running `terraform init -upgrade`.
`Unable to read workspaces: terraform-provider-funnel version is too old. Minimum required: 0.1.0`

### 429 Too Many Requests error

If `terraform plan` or `terraform apply` fails with a `429 Too Many Requests` error, the provider has exceeded the Funnel API rate limit.
If it happens frequently, consider contacting Funnel support to discuss your use case.

## Import command

Workspaces are imported by `<workspace_id>`. All other resources are imported by `<workspace_id>/<resource_id>`.

To import an existing export into your Terraform state, run:

```bash
terraform import funnel_<type>_export.<name> <workspace_id>/<export_id>
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_id` (String) Funnel System user client ID.
- `client_secret` (String, Sensitive) Funnel System user client secret.
- `subscription_id` (String) Funnel subscription ID. E.g. `fsXXXXXXXXXXX`.

### Optional

- `environment` (String) Funnel environment to manage. Default `us`. One of `us`, `eu`, `stage`, or `dev`.
