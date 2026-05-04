# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

- Funnel Measurment exports as a resource.

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
