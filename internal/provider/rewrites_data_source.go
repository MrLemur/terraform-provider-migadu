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

var _ datasource.DataSource = &RewritesDataSource{}

func NewRewritesDataSource() datasource.DataSource {
	return &RewritesDataSource{}
}

type RewritesDataSource struct {
	client *migadu.Client
}

type RewritesDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	Rewrites   types.List   `tfsdk:"rewrites"`
}

type RewriteListItemModel struct {
	Name          types.String `tfsdk:"name"`
	LocalPartRule types.String `tfsdk:"local_part_rule"`
	OrderNum      types.Int64  `tfsdk:"order_num"`
	Destinations  types.List   `tfsdk:"destinations"`
}

func (d *RewritesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrites"
}

func (d *RewritesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all rewrite rules for a Migadu domain (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"rewrites": schema.ListNestedAttribute{
				MarkdownDescription: "List of rewrite rules for this domain.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the rewrite rule.",
							Computed:            true,
						},
						"local_part_rule": schema.StringAttribute{
							MarkdownDescription: "The local part matching rule (supports wildcards).",
							Computed:            true,
						},
						"order_num": schema.Int64Attribute{
							MarkdownDescription: "Order number for rule processing (lower numbers processed first).",
							Computed:            true,
						},
						"destinations": schema.ListAttribute{
							MarkdownDescription: "List of destination email addresses.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *RewritesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RewritesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RewritesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	rewrites, err := d.client.ListRewrites(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list rewrites, got error: %s", err))
		return
	}

	items := make([]RewriteListItemModel, 0, len(rewrites))
	for _, rewrite := range rewrites {
		destinations, diags := types.ListValueFrom(ctx, types.StringType, rewrite.Destinations)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		items = append(items, RewriteListItemModel{
			Name:          types.StringValue(rewrite.Name),
			LocalPartRule: types.StringValue(rewrite.LocalPartRule),
			OrderNum:      types.Int64Value(int64(rewrite.OrderNum)),
			Destinations:  destinations,
		})
	}

	rewritesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":            types.StringType,
			"local_part_rule": types.StringType,
			"order_num":       types.Int64Type,
			"destinations":    types.ListType{ElemType: types.StringType},
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Rewrites = rewritesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
