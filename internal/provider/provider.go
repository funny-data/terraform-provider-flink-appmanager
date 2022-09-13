package provider

import (
	"context"
	"fmt"
	"git.sofunny.io/data-analysis-public/flink-appmanager-sdk/go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"time"
)

const (
	DefaultWaitInterval = 3
	DefaultWaitTimeout  = 180
)

var _ provider.Provider = &appManagerProvider{}

type appManagerProvider struct {
	configured bool
	client     *client.Client
}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &appManagerProvider{}
	}
}

func (p *appManagerProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				Type:     types.StringType,
				Optional: true,
			},
			"wait_timeout": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"wait_interval": {
				Type:     types.Int64Type,
				Optional: true,
			},
		},
	}, nil
}

type providerData struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	WaitTimeout  types.Int64  `tfsdk:"wait_timeout"`
	WaitInterval types.Int64  `tfsdk:"wait_interval"`
}

func (p *appManagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var endpoint string

	if config.Endpoint.Unknown {
		resp.Diagnostics.AddError("Unable to create client", "Cannot use unknown value as endpoint")
		return
	}

	if config.Endpoint.Null {
		endpoint = os.Getenv("FLINK_APPMANAGER_ENDPOINT")
	} else {
		endpoint = config.Endpoint.Value
	}

	if endpoint == "" {
		resp.Diagnostics.AddError("endpoint cannot be an empty string", "endpoint cannot be an empty string")
		return
	}

	waitInterval := config.WaitInterval.Value
	if waitInterval == 0 {
		waitInterval = DefaultWaitInterval
	}

	waitTimeout := config.WaitTimeout.Value
	if waitTimeout == 0 {
		waitTimeout = DefaultWaitTimeout
	}

	p.client = client.SetUp(client.Config{
		Endpoint: endpoint,
		Interval: time.Duration(waitInterval) * time.Second,
		Timeout:  time.Duration(waitTimeout) * time.Second,
	})
	p.configured = true
}

func (p *appManagerProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"flink_appmanager_namespace":         namespaceResourceType{},
		"flink_appmanager_deployment_target": deploymentTargetResourceType{},
		"flink_appmanager_session_cluster":   sessionClusterResourceType{},
	}, nil
}

func (p *appManagerProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}

func convertProviderType(in provider.Provider) (appManagerProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*appManagerProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return appManagerProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return appManagerProvider{}, diags
	}

	return *p, diags
}
