package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IdentityDataSource{}

func NewIdentityDataSource() datasource.DataSource {
	return &IdentityDataSource{}
}

type IdentityDataSource struct {
	client *migadu.Client
}

type IdentityDataSourceModel struct {
	DomainName           types.String `tfsdk:"domain_name"`
	Mailbox              types.String `tfsdk:"mailbox"`
	LocalPart            types.String `tfsdk:"local_part"`
	Name                 types.String `tfsdk:"name"`
	Address              types.String `tfsdk:"address"`
	MaySend              types.Bool   `tfsdk:"may_send"`
	MayReceive           types.Bool   `tfsdk:"may_receive"`
	MayAccessImap        types.Bool   `tfsdk:"may_access_imap"`
	MayAccessPop3        types.Bool   `tfsdk:"may_access_pop3"`
	MayAccessManageSieve types.Bool   `tfsdk:"may_access_managesieve"`
	FooterActive         types.Bool   `tfsdk:"footer_active"`
	FooterPlainBody      types.String `tfsdk:"footer_plain_body"`
	FooterHTMLBody       types.String `tfsdk:"footer_html_body"`
}

func (d *IdentityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (d *IdentityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Migadu identity.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "The mailbox local part this identity belongs to.",
				Required:            true,
			},
			"local_part": schema.StringAttribute{
				MarkdownDescription: "The local part of the identity address.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the identity.",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Full email address.",
				Computed:            true,
			},
			"may_send": schema.BoolAttribute{
				MarkdownDescription: "Whether the identity can send emails.",
				Computed:            true,
			},
			"may_receive": schema.BoolAttribute{
				MarkdownDescription: "Whether the identity can receive emails.",
				Computed:            true,
			},
			"may_access_imap": schema.BoolAttribute{
				MarkdownDescription: "Whether IMAP access is allowed.",
				Computed:            true,
			},
			"may_access_pop3": schema.BoolAttribute{
				MarkdownDescription: "Whether POP3 access is allowed.",
				Computed:            true,
			},
			"may_access_managesieve": schema.BoolAttribute{
				MarkdownDescription: "Whether ManageSieve access is allowed.",
				Computed:            true,
			},
			"footer_active": schema.BoolAttribute{
				MarkdownDescription: "Whether email footer is active.",
				Computed:            true,
			},
			"footer_plain_body": schema.StringAttribute{
				MarkdownDescription: "Plain text email footer.",
				Computed:            true,
			},
			"footer_html_body": schema.StringAttribute{
				MarkdownDescription: "HTML email footer.",
				Computed:            true,
			},
		},
	}
}

func (d *IdentityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IdentityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxStr := data.Mailbox.ValueString()
	localPart := data.LocalPart.ValueString()

	identity, err := d.client.GetIdentity(ctx, domain, &migadu.Mailbox{LocalPart: mailboxStr}, &migadu.Identity{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read identity, got error: %s", err))
		return
	}

	data.Name = types.StringValue(identity.Name)
	data.Address = types.StringValue(identity.Address)
	data.MaySend = types.BoolValue(identity.MaySend)
	data.MayReceive = types.BoolValue(identity.MayReceive)
	data.MayAccessImap = types.BoolValue(identity.MayAccessImap)
	data.MayAccessPop3 = types.BoolValue(identity.MayAccessPop3)
	data.MayAccessManageSieve = types.BoolValue(identity.MayAccessManagesieve)
	data.FooterActive = types.BoolValue(identity.FooterActive)
	data.FooterPlainBody = types.StringValue(identity.FooterPlainBody)
	data.FooterHTMLBody = types.StringValue(identity.FooterHTMLBody)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
