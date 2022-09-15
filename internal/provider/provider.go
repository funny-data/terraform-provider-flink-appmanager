package provider

import (
	"context"
	"git.sofunny.io/data-analysis-public/flink-appmanager-sdk/go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"time"
)

const (
	DefaultWaitInterval = 3
	DefaultWaitTimeout  = 180
)

// Ensure FlinkAppManagerProvider satisfies various provider interfaces.
var _ provider.Provider = &FlinkAppManagerProvider{}
var _ provider.ProviderWithMetadata = &FlinkAppManagerProvider{}

// FlinkAppManagerProvider defines the provider implementation.
type FlinkAppManagerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// FlinkAppManagerProviderModel describes the provider data model.
type FlinkAppManagerProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	WaitTimeout  types.Int64  `tfsdk:"wait_timeout"`
	WaitInterval types.Int64  `tfsdk:"wait_interval"`
}

func (p *FlinkAppManagerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "flink_appmanager"
	resp.Version = p.version
}

func (p *FlinkAppManagerProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				MarkdownDescription: "Flink AppManager Endpoint",
				Type:                types.StringType,
				Optional:            true,
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

func (p *FlinkAppManagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FlinkAppManagerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

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

	c := client.SetUp(client.Config{
		Endpoint: endpoint,
		Interval: time.Duration(waitInterval) * time.Second,
		Timeout:  time.Duration(waitTimeout) * time.Second,
	})

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *FlinkAppManagerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeploymentTargetResource,
		NewNamespaceResource,
		NewSessionClusterResource,
	}
}

func (p *FlinkAppManagerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FlinkAppManagerProvider{
			version: version,
		}
	}
}
