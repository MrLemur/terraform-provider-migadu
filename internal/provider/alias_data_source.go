package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AliasDataSource{}

func NewAliasDataSource() datasource.DataSource {
	return &AliasDataSource{}
}

type AliasDataSource struct {
	client *migadu.Client
}

type AliasDataSourceModel struct {
	DomainName   types.String `tfsdk:"domain_name"`
	LocalPart    types.String `tfsdk:"local_part"`
	Address      types.String `tfsdk:"address"`
	Destinations types.List   `tfsdk:"destinations"`
	IsInternal   types.Bool   `tfsdk:"is_internal"`
}

func (d *AliasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alias"
}

func (d *AliasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Migadu alias.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"local_part": schema.StringAttribute{
				MarkdownDescription: "The local part of the email address.",
				Required:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Full email address.",
				Computed:            true,
			},
			"destinations": schema.ListAttribute{
				MarkdownDescription: "List of destination email addresses.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_internal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an internal alias.",
				Computed:            true,
			},
		},
	}
}

func (d *AliasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AliasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	localPart := data.LocalPart.ValueString()

	alias, err := d.client.GetAlias(ctx, domain, &migadu.Alias{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alias, got error: %s", err))
		return
	}

	data.Address = types.StringValue(alias.Address)
	data.IsInternal = types.BoolValue(alias.IsInternal)

	destinations, diags := types.ListValueFrom(ctx, types.StringType, alias.Destinations)
	resp.Diagnostics.Append(diags...)
	data.Destinations = destinations

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
