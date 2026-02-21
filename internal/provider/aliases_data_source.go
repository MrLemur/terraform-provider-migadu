package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AliasesDataSource{}

func NewAliasesDataSource() datasource.DataSource {
	return &AliasesDataSource{}
}

type AliasesDataSource struct {
	client *migadu.Client
}

type AliasesDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	Aliases    types.List   `tfsdk:"aliases"`
}

type AliasListItemModel struct {
	LocalPart    types.String `tfsdk:"local_part"`
	Address      types.String `tfsdk:"address"`
	Destinations types.List   `tfsdk:"destinations"`
	IsInternal   types.Bool   `tfsdk:"is_internal"`
}

func (d *AliasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aliases"
}

func (d *AliasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all aliases for a Migadu domain (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"aliases": schema.ListNestedAttribute{
				MarkdownDescription: "List of aliases for this domain.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"local_part": schema.StringAttribute{
							MarkdownDescription: "The local part of the email address.",
							Computed:            true,
						},
						"address": schema.StringAttribute{
							MarkdownDescription: "Full email address.",
							Computed:            true,
						},
						"destinations": schema.ListAttribute{
							MarkdownDescription: "List of destination email addresses.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"is_internal": schema.BoolAttribute{
							MarkdownDescription: "Whether this is an internal alias.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *AliasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AliasesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	aliases, err := d.client.ListAliases(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list aliases, got error: %s", err))
		return
	}

	items := make([]AliasListItemModel, 0, len(aliases))
	for _, alias := range aliases {
		destinations, diags := types.ListValueFrom(ctx, types.StringType, alias.Destinations)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		items = append(items, AliasListItemModel{
			LocalPart:    types.StringValue(alias.LocalPart),
			Address:      types.StringValue(alias.Address),
			Destinations: destinations,
			IsInternal:   types.BoolValue(alias.IsInternal),
		})
	}

	aliasesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"local_part":   types.StringType,
			"address":      types.StringType,
			"destinations": types.ListType{ElemType: types.StringType},
			"is_internal":  types.BoolType,
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Aliases = aliasesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
