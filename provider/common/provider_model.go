package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FunnelProviderModel is the provider configuration model and is used across the provider
type FunnelProviderModel struct {
	Environment    types.String `tfsdk:"environment"`
	SubscriptionId types.String `tfsdk:"subscription_id"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	Token          string       `tfsdk:"-"`
}
