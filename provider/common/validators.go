package common

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func WorkspaceIdValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[A-Za-z0-9-_]{20}$`),
		"must be exactly 20 characters and contain only alphanumeric characters, hyphens, and underscores",
	)
}

func CurrencyValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[A-Z]{3}$`),
		"must be an uppercase ISO 4217 currency code, e.g. USD",
	)
}

func PeriodValidator() validator.String {
	return stringvalidator.OneOf("day", "week", "month", "quarter", "year")
}

// Alphanumerics and underscores, plus template placeholders such as {date}.
func BigqueryTableTemplateValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^([A-Za-z0-9_]|\{[A-Za-z]+\})+$`),
		"must contain only alphanumeric characters, underscores and placeholders such as {date}",
	)
}

func BigqueryDatasetIdValidator() validator.String {
	return stringvalidator.All(
		stringvalidator.LengthBetween(1, 1024),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9_]+$`),
			"must contain only alphanumeric characters and underscores",
		),
	)
}

// Allows the legacy domain-scoped form, e.g. example.com:my-project.
func BigqueryProjectIdValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^([a-z0-9.-]+:)?[a-z][a-z0-9-]{4,28}[a-z0-9]$`),
		"must be 6-30 lowercase letters, digits or hyphens, start with a letter and not end with a hyphen",
	)
}

// Unquoted Snowflake identifier rules.
func SnowflakeIdentifierValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]{0,254}$`),
		"must be 1-255 characters, start with a letter or underscore and contain only alphanumeric characters, underscores and $",
	)
}

func SnowflakeAccountValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`),
		"must contain only alphanumeric characters, dots, hyphens and underscores",
	)
}

func GcsBucketValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,220}[a-z0-9]$`),
		"must be 3-222 characters, contain only lowercase letters, digits, dots, hyphens and underscores, and start and end with a letter or digit",
	)
}
