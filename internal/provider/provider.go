package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Sonny00/terraform-provider-thehive/internal/client"
)

// Compile-time interface check.
var _ provider.Provider = &TheHiveProvider{}

// TheHiveProvider implements the top-level provider.
type TheHiveProvider struct {
	version string
}

// theHiveProviderModel maps the HCL provider block to Go types.
type theHiveProviderModel struct {
	URL    types.String `tfsdk:"url"`
	APIKey types.String `tfsdk:"api_key"`
}

// New returns a factory function — the signature expected by providerserver.Serve.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TheHiveProvider{
			version: version,
		}
	}
}

func (p *TheHiveProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "thehive"
	resp.Version = p.version
}

func (p *TheHiveProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage TheHive 5 configuration resources (organisations, users, profiles, custom fields, …) as code.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Base URL of the TheHive instance (e.g. https://thehive.example.com). " +
					"Falls back to the THEHIVE_URL environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key used to authenticate against TheHive. " +
					"Falls back to the THEHIVE_API_KEY environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure is called by Terraform to set up the provider client.
// The resulting *client.Client is injected into every resource and data source
// via resp.ResourceData / resp.DataSourceData.
func (p *TheHiveProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "configuring TheHive provider")

	var config theHiveProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// --- Resolve URL ---
	url := os.Getenv("THEHIVE_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if url == "" {
		resp.Diagnostics.AddError(
			"Missing TheHive URL",
			"Set \"url\" in the provider block or export THEHIVE_URL.",
		)
		return
	}

	// --- Resolve API key ---
	apiKey := os.Getenv("THEHIVE_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing TheHive API key",
			"Set \"api_key\" in the provider block or export THEHIVE_API_KEY.",
		)
		return
	}

	// --- Build HTTP client ---
	c, err := client.NewClient(url, apiKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create TheHive client", err.Error())
		return
	}

	tflog.Debug(ctx, "TheHive client configured", map[string]any{"url": url})

	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns the list of managed resources the provider exposes.
func (p *TheHiveProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganisationResource,
		NewUserResource,
		NewProfileResource,
		NewCustomFieldResource,
		NewObservableTypeResource,
		NewCaseTemplateResource,
		NewPageTemplateResource,
		NewCortexConfigResource,
		NewMISPConfigResource,
		NewCaseStatusResource,
		NewAlertStatusResource,
		NewEmailIntakeResource,
		NewFunctionResource,
		NewTaxonomyResource,
		NewAlertFeederResource,
	}
}

// DataSources returns the list of read-only data sources.
func (p *TheHiveProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProfileDataSource,
		NewOrganisationDataSource,
		NewUserDataSource,
		NewCustomFieldDataSource,
		NewObservableTypeDataSource,
		NewCaseTemplateDataSource,
	}
}
