package provider

import (
	"context"
	"os"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure MigaduProvider satisfies various provider interfaces.
var _ provider.Provider = &MigaduProvider{}

// MigaduProvider defines the provider implementation.
type MigaduProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MigaduProviderModel describes the provider data model.
type MigaduProviderModel struct {
	Username types.String `tfsdk:"username"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *MigaduProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "migadu"
	resp.Version = p.version
}

func (p *MigaduProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Migadu provider enables Terraform to manage Migadu email service resources such as mailboxes and aliases.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Migadu admin username (email address). Can also be set via the MIGADU_USERNAME environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Migadu API key. Can also be set via the MIGADU_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *MigaduProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MigaduProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	username := config.Username.ValueString()
	apiKey := config.APIKey.ValueString()

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Migadu Username",
			"The provider cannot create the Migadu API client as there is an unknown configuration value for the Migadu username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MIGADU_USERNAME environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Migadu API Key",
			"The provider cannot create the Migadu API client as there is an unknown configuration value for the Migadu API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MIGADU_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	if username == "" {
		username = os.Getenv("MIGADU_USERNAME")
	}

	if apiKey == "" {
		apiKey = os.Getenv("MIGADU_API_KEY")
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Migadu Username",
			"The provider cannot create the Migadu API client as there is a missing or empty value for the Migadu username. "+
				"Set the username value in the configuration or use the MIGADU_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Migadu API Key",
			"The provider cannot create the Migadu API client as there is a missing or empty value for the Migadu API key. "+
				"Set the api_key value in the configuration or use the MIGADU_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Migadu client using the configuration values
	client, err := migadu.New(username, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Migadu API Client",
			"An unexpected error occurred when creating the Migadu API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Migadu Client Error: "+err.Error(),
		)
		return
	}

	// Make the Migadu client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MigaduProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewDomainActivationResource,
		NewMailboxResource,
		NewAliasResource,
		NewIdentityResource,
		NewRewriteResource,
	}
}

func (p *MigaduProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
		NewDomainsDataSource,
		NewMailboxDataSource,
		NewMailboxesDataSource,
		NewAliasDataSource,
		NewAliasesDataSource,
		NewIdentityDataSource,
		NewIdentitiesDataSource,
		NewRewriteDataSource,
		NewRewritesDataSource,
		NewForwardingsDataSource,
		NewDomainDiagnosticsDataSource,
		NewDomainDNSRecordsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MigaduProvider{
			version: version,
		}
	}
}
