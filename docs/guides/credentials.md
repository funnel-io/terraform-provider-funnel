---
page_title: "Credentials"
subcategory: ""
---

# Credentials

Data sources and some exports connect to an external platform with a Funnel credential, set in the `credential_id` argument.
A credential is always owned by a regular Funnel user, never by a System user.
To use it with Terraform, the owner shares it with the System user the provider authenticates as.

## Share a credential with the System user

Only the owner of a credential can share it, and it can only be shared with a System user.

1. Log in to Funnel as the user who owns the credential.
2. Open the User menu and go to My credentials.
3. Click on the credential and press the Share button.
4. Choose the System user whose `client_id` and `client_secret` the provider is configured with.

The owner can revoke the sharing at any time.

## Find the credential ID

1. Go to the Subscription overview in Funnel and open the Users section.
2. Open the System users tab and click on the System user the provider is configured with. This lists the credentials shared with it.
3. Copy the Credential ID into the `credential_id` argument.

## Data sources

The `funnel_data_source` resource uses a credential for the platform it downloads data from.

## Exports

The `funnel_bigquery_export` and `funnel_snowflake_export` resources take a `credential_id` in their `destination` attribute.

- **BigQuery:** Optional. If not set, the export uses a Funnel service account, which needs write access to the BigQuery project.
- **Snowflake:** Required. Use either a Snowflake upload (PAT) or a Snowflake upload (Private Key) credential.
