package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainDiagnosticsDataSource{}

func NewDomainDiagnosticsDataSource() datasource.DataSource {
	return &DomainDiagnosticsDataSource{}
}

type DomainDiagnosticsDataSource struct {
	client *migadu.Client
}

type DomainDiagnosticsDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	MX         types.List   `tfsdk:"mx"`
	SPF        types.List   `tfsdk:"spf"`
	DKIM       types.List   `tfsdk:"dkim"`
	DMARC      types.List   `tfsdk:"dmarc"`
}

func (d *DomainDiagnosticsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_diagnostics"
}

func (d *DomainDiagnosticsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs DNS diagnostics for a Migadu domain.\n\n" +
			"Each attribute returns a list of validation error messages for that record type. " +
			"An empty list means the record is correctly configured.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name to diagnose.",
				Required:            true,
			},
			"mx": schema.ListAttribute{
				MarkdownDescription: "MX record validation errors. Empty if correctly configured.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"spf": schema.ListAttribute{
				MarkdownDescription: "SPF record validation errors. Empty if correctly configured.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"dkim": schema.ListAttribute{
				MarkdownDescription: "DKIM record validation errors. Empty if correctly configured.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"dmarc": schema.ListAttribute{
				MarkdownDescription: "DMARC record validation errors. Empty if correctly configured.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *DomainDiagnosticsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*migadu.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *migadu.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DomainDiagnosticsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDiagnosticsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	diagnostics, err := d.client.GetDomainDiagnostics(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get domain diagnostics, got error: %s", err))
		return
	}

	mx, diags := types.ListValueFrom(ctx, types.StringType, diagnostics.MX)
	resp.Diagnostics.Append(diags...)
	data.MX = mx

	spf, diags := types.ListValueFrom(ctx, types.StringType, diagnostics.SPF)
	resp.Diagnostics.Append(diags...)
	data.SPF = spf

	dkim, diags := types.ListValueFrom(ctx, types.StringType, diagnostics.DKIM)
	resp.Diagnostics.Append(diags...)
	data.DKIM = dkim

	dmarc, diags := types.ListValueFrom(ctx, types.StringType, diagnostics.DMARC)
	resp.Diagnostics.Append(diags...)
	data.DMARC = dmarc

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
