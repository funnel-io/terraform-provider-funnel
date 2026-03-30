# Contributing

## Commits

This project follows [Conventional Commits](https://www.conventionalcommits.org/). Each commit message should be structured as:

```
<type>[optional scope]: <description>

[optional body]
```

Common types:

| Type | Purpose |
|------|---------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks (deps, CI, tooling) |

Examples:

```
feat(data_source): add connection data source
fix: handle nil response from API
chore: bump Go to 1.23
```

## Documentation

To preview the documentation for the Terraform registry use this [site](https://registry.terraform.io/tools/doc-preview).

Add example HCL in the `examples/` folder for every data source and resource.

## Logging

Use the "tflog” package to write logs for your provider.

To follow the HTTP requests from the provider run Terraform with the environment variable `TF_LOG_PROVIDER=INFO`. See: https://developer.hashicorp.com/terraform/plugin/log/managing.
