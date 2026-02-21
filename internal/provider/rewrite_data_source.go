package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RewriteDataSource{}

func NewRewriteDataSource() datasource.DataSource {
	return &RewriteDataSource{}
}

type RewriteDataSource struct {
	client *migadu.Client
}

type RewriteDataSourceModel struct {
	DomainName    types.String `tfsdk:"domain_name"`
	Name          types.String `tfsdk:"name"`
	LocalPartRule types.String `tfsdk:"local_part_rule"`
	OrderNum      types.Int64  `tfsdk:"order_num"`
	Destinations  types.List   `tfsdk:"destinations"`
}

func (d *RewriteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite"
}

func (d *RewriteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Migadu rewrite rule.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the rewrite rule.",
				Required:            true,
			},
			"local_part_rule": schema.StringAttribute{
				MarkdownDescription: "The local part matching rule.",
				Computed:            true,
			},
			"order_num": schema.Int64Attribute{
				MarkdownDescription: "Order number for rule processing.",
				Computed:            true,
			},
			"destinations": schema.ListAttribute{
				MarkdownDescription: "List of destination email addresses.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *RewriteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RewriteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RewriteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	name := data.Name.ValueString()

	rewrite, err := d.client.GetRewrite(ctx, domain, &migadu.Rewrite{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rewrite, got error: %s", err))
		return
	}

	data.LocalPartRule = types.StringValue(rewrite.LocalPartRule)
	data.OrderNum = types.Int64Value(int64(rewrite.OrderNum))

	destinations, diags := types.ListValueFrom(ctx, types.StringType, rewrite.Destinations)
	resp.Diagnostics.Append(diags...)
	data.Destinations = destinations

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
