# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - Unreleased

### Changed

- Replaced `partition_schema` on exports with `partition_by`. It takes the period, such as `month`, or `all` for a single part. Use `none` for a single part with only data that has no values for Funnel date, which used to be `by = "none"`.
- Removed `format` from exports. `metrics` is always `export` and headers are always `safename`. `funnel_bigquery_export` and `funnel_snowflake_export` always export CSV, so existing exports set to TSV or Parquet switch to CSV on the next apply. `funnel_gcs_export` takes the file format as `destination.file_format`.
- Renamed `destination.output_id_template` on `funnel_bigquery_export` to `destination.table_name_template`, matching the name in the Funnel UI.
- `funnel_gcs_export` no longer gzips Parquet files, which are already compressed. Setting `gzip = true` with `file_format = "parquet"` is now an error.

### Fixed

- `partition_by` no longer accepts `week`, which the API doesn't support.
- `gzip = false` on `funnel_gcs_export` is now respected. Before, `gzip` was always sent as `true`.
- Removing `start`, `end`, `rolling_start` or `rolling_end` from an export's `range` now clears it in Funnel. Before, the old value was kept and every plan showed a diff.

### Removed

- `funnel_measurement_export`, deprecated since 0.2.4. Remove it from your configuration and state before upgrading.
- The `snapshot` value for `partition_schema.by`. It only applied to measurement exports.

## [0.2.7] - 2026-11-18

### Added

- `derived` on `funnel_custom_metric` and `funnel_custom_dimension`, so custom fields with rules can be managed. It takes the rules as JSON, and it is easiest to build them in the Funnel UI and import the field.
- `funnel_custom_metric` and `funnel_custom_dimension` data sources, to copy a field and its rules from one workspace to another.
- `funnel_data_source` data source, to copy a data source from one workspace to another.

### Fixed

- Exports deleted outside Terraform are now removed from state and recreated on the next apply. Before, every plan failed with a read error.
- Exports no longer show a perpetual diff when `enabled` is not set. It now defaults to `false`, which is what the API already received.
- `funnel_gcs_export` now reports a clear error when an export with the same destination already exists, like the other exports.
- `funnel_custom_metric` no longer fails after apply when `precision` is not set. It now defaults to `0`.

## [0.2.6] - 2026-10-27

### Added

- Plan-time validation of `workspace` in `funnel_custom_metric` and `funnel_custom_dimension`. It must be exactly 20 characters, the same as for exports and `funnel_data_source`.

### Changed

- Fixed attribute descriptions that were wrong or unclear, such as the export `id`, the custom dimension `description`, `format.type` and the provider's `client_id` and `client_secret`.
- Fixed examples with invalid or misleading values for the workspace ID, the GA4 `remote_struct`, the Snowflake `account_locator` and the filters.
- Every resource page now documents how to import it. Workspaces are imported by `<workspace_id>`, and all other resources by `<workspace_id>/<resource_id>`.

### Removed

- `destination.username` from `funnel_snowflake_export`. It only applied to V1 Snowflake exports, and the provider creates V2 exports. Remove it from your configuration before upgrading.

## [0.2.5] - 2026-10-21

### Added

- Plan-time validation for export attributes, so invalid values fail at `terraform plan` instead of at the API.

### Changed

- Attribute documentation now describes the allowed values. Examples use `CRON_TZ=` schedules, and the filter example in the provider docs explains how multiple filters combine.

### Fixed

- Exports no longer show a perpetual diff when `notes`, `currency`, `filters`, `filters[].or`, `filters[].operation`, `filters[].value`, `fields[].type`, `fields[].export_name`, `fields[].export_type`, `range.start` or `range.end` are not set. Existing state may need one `terraform apply` to settle.

## [0.2.4] - 2026-10-21

### Added

- Add the ability to provide a `credential_id` to `funnel_bigquery_export`.
- Add `credential_id` to `funnel_snowflake_export`. It is required and replaces `personal_access_token` and `private_key`.

### Changed

- Updated Go to 1.27.0 and bumped all dependencies.
- `username` is now optional in `funnel_snowflake_export`.

### Deprecated

- `funnel_measurement_export` (Iceberg export) is deprecated and will be removed in a future release.

### Removed

- **Breaking:** `personal_access_token` and `private_key` have been removed from `funnel_snowflake_export`. Use `credential_id` instead.

## [0.2.3] - 2026-08-31

### Added

- Add the option to create a demo data source that produces demo (test) data in Funnel instead of connecting to the external platform by setting `is_demo` to true for `funnel_data_source`.

### Changed

- Updated the Funnel data source `template_id` validation to accept the slug based template IDs.
- Expand the Funnel data source documentation.
- Add a guide in the documentation for the resource data source `remote_struct` field where users can look up the structure of the `remote_struct`.

## [0.2.2] - 2026-08-06

### Changed

- Improved the error messages around issues with the Funnel CLIENT_ID and CLIENT_SECRET authentication.
- Updated the examples in the export resources documentation.

## [0.2.1] - 2026-06-16

### Added

- Support for `template_id` in data source resource to replace deprecated `report_type`.
- Support for `remote_struct` configuration in data source resource.

### Fixed

- Made `partition_schema` required for export resources (BigQuery, GCS, Snowflake).
- Updated partition schema validation to support `snapshot` partitioning (measurement exports only).
- Expanded `per` field validation to include `week` and `quarter` options.
- Fixed incorrect export examples in documentation and example files.
- Fixed various typos in provider and resource documentation.

### Changed

- Replaced `report_type` with `template_id` in data source resource (breaking change).
- Enhanced export resource documentation with recommended partition schema settings.
- Updated data source documentation to remove report type references.

## [0.2.0] - 2026-04-24

### Added

- Resource for Funnel data sources (`funnel_data_source`) for a limited set of connectors.
- Examples for all resources and data sources.
- Support for Snowflake exports with private key authentication.

### Changed

- Improved error handling for bad requests in Funnel API client.
- Enhanced Snowflake export documentation.

## [0.1.6] - 2026-03-27

### Added

- A first iteration of Funnel custom dimensions and metrics resources.
- A data source for workspaces

### Changed

- Updated the terraform-plugin-framework to v1.19.0.

## [0.1.5]

### Changed

- Switch the Auth0 URLs.

### Added

- Import functionality for the Workspace resource.

## [0.1.4]

### Added

- Funnel Measurement exports as a resource.

### Removed

- The old workspace data source.

## [0.1.3]

### Added

- Workspace as a resource.

## [0.1.2]

### Changed

- Modernized the Go typing and updated Go to 1.26.0.
- Switched provider name to the published name.

## [0.1.1]

### Added

- Initial release of the Funnel Terraform Provider.
- Support for Funnel data sources `workspace` and `export_field`.
- Support for Funnel export resources BigQuery, Google Cloud Storage (GCS) and Snowflake.
