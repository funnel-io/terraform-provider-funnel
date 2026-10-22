package common

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidators(t *testing.T) {
	cases := []struct {
		name    string
		v       validator.String
		valid   []string
		invalid []string
	}{
		{"workspace", WorkspaceIdValidator(), []string{"abcdefghij0123456789", "abc-_efghij012345678"}, []string{"short", "abcdefghij01234567890", "abcdefghij012345678!"}},
		{"currency", CurrencyValidator(), []string{"USD", "SEK"}, []string{"usd", "US", "USDX"}},
		{"period", PeriodValidator(), []string{"day", "year"}, []string{"days", "Day"}},
		{"bq table", BigqueryTableTemplateValidator(), []string{"daily_export", "daily_export_{date}", "{date}"}, []string{"daily-export", "a b", "export_{}", "export_{date"}},
		{"bq dataset", BigqueryDatasetIdValidator(), []string{"my_dataset", "Dataset1"}, []string{"my-dataset", "my.dataset", ""}},
		{"bq project", BigqueryProjectIdValidator(), []string{"my-project-123", "example.com:my-project"}, []string{"My-Project", "1project", "proj-", "abc"}},
		{"sf identifier", SnowflakeIdentifierValidator(), []string{"MARKETING_DB", "_tmp", "tbl$1"}, []string{"1DB", "MY-DB", "my db", "db.schema"}},
		{"sf account", SnowflakeAccountValidator(), []string{"xy12345.us-east-1", "funnel-snowflake"}, []string{"-acc", "acc/1", "acc name"}},
		{"gcs bucket", GcsBucketValidator(), []string{"my-funnel-exports", "my.bucket_1"}, []string{"My-Bucket", "ab", "-bucket", "bucket-"}},
	}

	for _, c := range cases {
		check := func(value string, wantErr bool) {
			resp := &validator.StringResponse{}
			c.v.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(value)}, resp)
			if resp.Diagnostics.HasError() != wantErr {
				t.Errorf("%s: %q, expected error=%v", c.name, value, wantErr)
			}
		}
		for _, v := range c.valid {
			check(v, false)
		}
		for _, v := range c.invalid {
			check(v, true)
		}
	}
}
