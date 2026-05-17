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

var _ datasource.DataSource = &DomainsDataSource{}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

type DomainsDataSource struct {
	client *migadu.Client
}

type DomainsDataSourceModel struct {
	Domains types.List `tfsdk:"domains"`
}

type DomainListItemModel struct {
	Name                 types.String `tfsdk:"name"`
	State                types.String `tfsdk:"state"`
	Description          types.String `tfsdk:"description"`
	SpamAggressiveness   types.String `tfsdk:"spam_aggressiveness"`
	GreylistingEnabled   types.Bool   `tfsdk:"greylisting_enabled"`
	MXProxyEnabled       types.Bool   `tfsdk:"mx_proxy_enabled"`
	HostedDNS            types.Bool   `tfsdk:"hosted_dns"`
	SenderAllowlist      types.List   `tfsdk:"sender_allowlist"`
	SenderDenylist       types.List   `tfsdk:"sender_denylist"`
	RecipientDenylist    types.List   `tfsdk:"recipient_denylist"`
	CatchallDestinations types.List   `tfsdk:"catchall_destinations"`
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all Migadu domains (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				MarkdownDescription: "List of domains in the account.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":        schema.StringAttribute{MarkdownDescription: "The domain name.", Computed: true},
						"state":       schema.StringAttribute{MarkdownDescription: "Domain state.", Computed: true},
						"description": schema.StringAttribute{MarkdownDescription: "Domain description.", Computed: true},
						"spam_aggressiveness": schema.StringAttribute{
							MarkdownDescription: "Spam filter aggressiveness level.",
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
						"sender_allowlist": schema.ListAttribute{
							MarkdownDescription: "List of allowed sender addresses.",
							Computed:            true,
							ElementType:         types.StringType,
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
				},
			},
		},
	}
}

func (d *DomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domains, err := d.client.ListDomains(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list domains, got error: %s", err))
		return
	}

	items := make([]DomainListItemModel, 0, len(domains))
	for _, domain := range domains {
		senderAllowlist, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(domain.SenderAllowlist))
		resp.Diagnostics.Append(diags...)
		senderDenylist, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(domain.SenderDenylist))
		resp.Diagnostics.Append(diags...)
		recipientDenylist, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(domain.RecipientDenylist))
		resp.Diagnostics.Append(diags...)
		catchallDestinations, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(domain.CatchallDestinations))
		resp.Diagnostics.Append(diags...)

		items = append(items, DomainListItemModel{
			Name:                 types.StringValue(domain.Name),
			State:                types.StringValue(domain.State),
			Description:          types.StringValue(domain.Description),
			SpamAggressiveness:   types.StringValue(domain.SpamAggressiveness),
			GreylistingEnabled:   types.BoolValue(domain.GreylistingEnabled),
			MXProxyEnabled:       types.BoolValue(domain.MXProxyEnabled),
			HostedDNS:            types.BoolValue(domain.HostedDNS),
			SenderAllowlist:      senderAllowlist,
			SenderDenylist:       senderDenylist,
			RecipientDenylist:    recipientDenylist,
			CatchallDestinations: catchallDestinations,
		})
	}

	domainsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                  types.StringType,
			"state":                 types.StringType,
			"description":           types.StringType,
			"spam_aggressiveness":   types.StringType,
			"greylisting_enabled":   types.BoolType,
			"mx_proxy_enabled":      types.BoolType,
			"hosted_dns":            types.BoolType,
			"sender_allowlist":      types.ListType{ElemType: types.StringType},
			"sender_denylist":       types.ListType{ElemType: types.StringType},
			"recipient_denylist":    types.ListType{ElemType: types.StringType},
			"catchall_destinations": types.ListType{ElemType: types.StringType},
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Domains = domainsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
