package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MailboxDataSource{}

func NewMailboxDataSource() datasource.DataSource {
	return &MailboxDataSource{}
}

type MailboxDataSource struct {
	client *migadu.Client
}

type MailboxDataSourceModel struct {
	DomainName           types.String  `tfsdk:"domain_name"`
	LocalPart            types.String  `tfsdk:"local_part"`
	Name                 types.String  `tfsdk:"name"`
	Address              types.String  `tfsdk:"address"`
	IsInternal           types.Bool    `tfsdk:"is_internal"`
	MaySend              types.Bool    `tfsdk:"may_send"`
	MayReceive           types.Bool    `tfsdk:"may_receive"`
	MayAccessImap        types.Bool    `tfsdk:"may_access_imap"`
	MayAccessPop3        types.Bool    `tfsdk:"may_access_pop3"`
	MayAccessManageSieve types.Bool    `tfsdk:"may_access_managesieve"`
	StorageUsage         types.Float64 `tfsdk:"storage_usage"`
	ChangedAt            types.String  `tfsdk:"changed_at"`
	LastLoginAt          types.String  `tfsdk:"last_login_at"`
}

func (d *MailboxDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox"
}

func (d *MailboxDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Migadu mailbox.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"local_part": schema.StringAttribute{
				MarkdownDescription: "The local part of the email address.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name.",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Full email address.",
				Computed:            true,
			},
			"is_internal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an internal mailbox.",
				Computed:            true,
			},
			"may_send": schema.BoolAttribute{
				MarkdownDescription: "Whether the mailbox can send emails.",
				Computed:            true,
			},
			"may_receive": schema.BoolAttribute{
				MarkdownDescription: "Whether the mailbox can receive emails.",
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
			"storage_usage": schema.Float64Attribute{
				MarkdownDescription: "Storage usage in bytes.",
				Computed:            true,
			},
			"changed_at": schema.StringAttribute{
				MarkdownDescription: "Last modification timestamp.",
				Computed:            true,
			},
			"last_login_at": schema.StringAttribute{
				MarkdownDescription: "Last login timestamp.",
				Computed:            true,
			},
		},
	}
}

func (d *MailboxDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MailboxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MailboxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	localPart := data.LocalPart.ValueString()

	mailbox, err := d.client.GetMailbox(ctx, domain, &migadu.Mailbox{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read mailbox, got error: %s", err))
		return
	}

	data.Name = types.StringValue(mailbox.Name)
	data.Address = types.StringValue(mailbox.Address)
	data.IsInternal = types.BoolValue(mailbox.IsInternal)
	data.MaySend = types.BoolValue(mailbox.MaySend)
	data.MayReceive = types.BoolValue(mailbox.MayReceive)
	data.MayAccessImap = types.BoolValue(mailbox.MayAccessImap)
	data.MayAccessPop3 = types.BoolValue(mailbox.MayAccessPop3)
	data.MayAccessManageSieve = types.BoolValue(mailbox.MayAccessManagesieve)
	data.StorageUsage = types.Float64Value(mailbox.StorageUsage)
	data.ChangedAt = types.StringValue(mailbox.ChangedAt)
	data.LastLoginAt = types.StringValue(mailbox.LastLoginAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
