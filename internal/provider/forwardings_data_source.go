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

var _ datasource.DataSource = &ForwardingsDataSource{}

func NewForwardingsDataSource() datasource.DataSource {
	return &ForwardingsDataSource{}
}

type ForwardingsDataSource struct {
	client *migadu.Client
}

type ForwardingsDataSourceModel struct {
	DomainName  types.String `tfsdk:"domain_name"`
	Mailbox     types.String `tfsdk:"mailbox"`
	Forwardings types.List   `tfsdk:"forwardings"`
}

type ForwardingListItemModel struct {
	Address            types.String `tfsdk:"address"`
	BlockedAt          types.String `tfsdk:"blocked_at"`
	ConfirmationSentAt types.String `tfsdk:"confirmation_sent_at"`
	ConfirmedAt        types.String `tfsdk:"confirmed_at"`
	ExpiresOn          types.String `tfsdk:"expires_on"`
	IsActive           types.Bool   `tfsdk:"is_active"`
	RemoveUponExpiry   types.Bool   `tfsdk:"remove_upon_expiry"`
}

func (d *ForwardingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwardings"
}

func (d *ForwardingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all forwardings for a mailbox (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "The mailbox local part.",
				Required:            true,
			},
			"forwardings": schema.ListNestedAttribute{
				MarkdownDescription: "List of forwardings for this mailbox.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address":              schema.StringAttribute{MarkdownDescription: "The forwarding destination address.", Computed: true},
						"blocked_at":           schema.StringAttribute{MarkdownDescription: "Timestamp when the forwarding was blocked.", Computed: true},
						"confirmation_sent_at": schema.StringAttribute{MarkdownDescription: "Timestamp when confirmation was sent.", Computed: true},
						"confirmed_at":         schema.StringAttribute{MarkdownDescription: "Timestamp when the forwarding was confirmed.", Computed: true},
						"expires_on":           schema.StringAttribute{MarkdownDescription: "Expiry date of the forwarding.", Computed: true},
						"is_active":            schema.BoolAttribute{MarkdownDescription: "Whether the forwarding is active.", Computed: true},
						"remove_upon_expiry":   schema.BoolAttribute{MarkdownDescription: "Whether to remove the forwarding upon expiry.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *ForwardingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ForwardingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ForwardingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	forwardings, err := d.client.ListForwardings(ctx, domain, &migadu.Mailbox{LocalPart: data.Mailbox.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list forwardings, got error: %s", err))
		return
	}

	items := make([]ForwardingListItemModel, 0, len(forwardings))
	for _, forwarding := range forwardings {
		items = append(items, ForwardingListItemModel{
			Address:            types.StringValue(forwarding.Address),
			BlockedAt:          types.StringValue(forwarding.BlockedAt),
			ConfirmationSentAt: types.StringValue(forwarding.ConfirmationSentAt),
			ConfirmedAt:        types.StringValue(forwarding.ConfirmedAt),
			ExpiresOn:          types.StringValue(forwarding.ExpiresOn),
			IsActive:           types.BoolValue(forwarding.IsActive),
			RemoveUponExpiry:   types.BoolValue(forwarding.RemoveUponExpiry),
		})
	}

	forwardingsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address":              types.StringType,
			"blocked_at":           types.StringType,
			"confirmation_sent_at": types.StringType,
			"confirmed_at":         types.StringType,
			"expires_on":           types.StringType,
			"is_active":            types.BoolType,
			"remove_upon_expiry":   types.BoolType,
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Forwardings = forwardingsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
