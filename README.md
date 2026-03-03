# Terraform Provider Funnel

A [Terraform](https://www.terraform.io) provider for managing [Funnel](https://funnel.io) resources. Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

> **Note:** This provider is in an early state. APIs and behaviour may change.

## Documentation

To use this provider in your Terraform module, follow the documentation on [Terraform Registry](https://registry.terraform.io/providers/funnel-io/funnel/latest/docs).

## Contributors

See the [contribution guide](CONTRIBUTING.md).

## Releases

See [CHANGELOG.md](CHANGELOG.md) for full details.

### Creating a release

Releases are automated via GitHub Actions and [GoReleaser](https://goreleaser.com). To publish a new version:

1. Ensure `CHANGELOG.md` is up to date on `main`.
2. Create a tag on `main` through GitHub matching the pattern `v*` (e.g. `v0.1.3`):
3. The [release workflow](.github/workflows/release.yaml) will build, sign, and publish the provider to GitHub Releases. The Terraform Registry picks up the new version automatically.

## License

Copyright (c) 2026 Funnel.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
