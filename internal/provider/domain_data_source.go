package provider

import (
	"context"
	"fmt"
	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainDataSource{}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

type DomainDataSource struct {
	client *migadu.Client
}

type DomainDataSourceModel struct {
	Name                 types.String `tfsdk:"name"`
	State                types.String `tfsdk:"state"`
	Description          types.String `tfsdk:"description"`
	Tags                 types.List   `tfsdk:"tags"`
	SpamAggressiveness   types.Int64  `tfsdk:"spam_aggressiveness"`
	GreylistingEnabled   types.Bool   `tfsdk:"greylisting_enabled"`
	MXProxyEnabled       types.Bool   `tfsdk:"mx_proxy_enabled"`
	HostedDNS            types.Bool   `tfsdk:"hosted_dns"`
	SenderDenylist       types.List   `tfsdk:"sender_denylist"`
	RecipientDenylist    types.List   `tfsdk:"recipient_denylist"`
	CatchallDestinations types.List   `tfsdk:"catchall_destinations"`
}

func (d *DomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Migadu domain.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Domain state.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Domain description.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Domain tags.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"spam_aggressiveness": schema.Int64Attribute{
				MarkdownDescription: "Spam filter aggressiveness level (integer). See `migadu_domain` resource for valid values.",
				Computed:            true,
			},
			"greylisting_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether greylisting is enabled.",
				Computed:            true,
			},
			"mx_proxy_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether MX proxy is enabled.",
				Computed:            true,
			},
			"hosted_dns": schema.BoolAttribute{
				MarkdownDescription: "Whether DNS is hosted by Migadu.",
				Computed:            true,
			},
			"sender_denylist": schema.ListAttribute{
				MarkdownDescription: "List of denied sender addresses.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"recipient_denylist": schema.ListAttribute{
				MarkdownDescription: "List of denied recipient addresses.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"catchall_destinations": schema.ListAttribute{
				MarkdownDescription: "Catchall email destinations.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *DomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.Name.ValueString()}
	retrieved, err := d.client.GetDomain(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domain, got error: %s", err))
		return
	}

	data.State = types.StringValue(retrieved.State)
	data.Description = types.StringValue(retrieved.Description)
	data.SpamAggressiveness = types.Int64Value(int64(retrieved.SpamAggressiveness))
	data.GreylistingEnabled = types.BoolValue(retrieved.GreylistingEnabled)
	data.MXProxyEnabled = types.BoolValue(retrieved.MXProxyEnabled)
	data.HostedDNS = types.BoolValue(retrieved.HostedDNS)

	tags, diags := types.ListValueFrom(ctx, types.StringType, retrieved.Tags)
	resp.Diagnostics.Append(diags...)
	data.Tags = tags

	senderDenylist, diags := types.ListValueFrom(ctx, types.StringType, retrieved.SenderDenylist)
	resp.Diagnostics.Append(diags...)
	data.SenderDenylist = senderDenylist

	recipientDenylist, diags := types.ListValueFrom(ctx, types.StringType, splitRecipientDenylist(retrieved.RecipientDenylist))
	resp.Diagnostics.Append(diags...)
	data.RecipientDenylist = recipientDenylist

	catchallDestinations, diags := types.ListValueFrom(ctx, types.StringType, retrieved.CatchallDestinations)
	resp.Diagnostics.Append(diags...)
	data.CatchallDestinations = catchallDestinations

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
