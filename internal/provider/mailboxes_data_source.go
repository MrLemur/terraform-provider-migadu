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

var _ datasource.DataSource = &MailboxesDataSource{}

func NewMailboxesDataSource() datasource.DataSource {
	return &MailboxesDataSource{}
}

type MailboxesDataSource struct {
	client *migadu.Client
}

type MailboxesDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	Mailboxes  types.List   `tfsdk:"mailboxes"`
}

type MailboxListItemModel struct {
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

func (d *MailboxesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailboxes"
}

func (d *MailboxesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all mailboxes for a Migadu domain (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"mailboxes": schema.ListNestedAttribute{
				MarkdownDescription: "List of mailboxes for this domain.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"local_part": schema.StringAttribute{
							MarkdownDescription: "The local part of the email address.",
							Computed:            true,
						},
						"name":    schema.StringAttribute{MarkdownDescription: "Display name.", Computed: true},
						"address": schema.StringAttribute{MarkdownDescription: "Full email address.", Computed: true},
						"is_internal": schema.BoolAttribute{
							MarkdownDescription: "Whether this is an internal mailbox.",
							Computed:            true,
						},
						"may_send":    schema.BoolAttribute{MarkdownDescription: "Whether the mailbox can send emails.", Computed: true},
						"may_receive": schema.BoolAttribute{MarkdownDescription: "Whether the mailbox can receive emails.", Computed: true},
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
						"changed_at":    schema.StringAttribute{MarkdownDescription: "Last modification timestamp.", Computed: true},
						"last_login_at": schema.StringAttribute{MarkdownDescription: "Last login timestamp.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *MailboxesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MailboxesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MailboxesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxes, err := d.client.ListMailboxes(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list mailboxes, got error: %s", err))
		return
	}

	items := make([]MailboxListItemModel, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		items = append(items, MailboxListItemModel{
			LocalPart:            types.StringValue(mailbox.LocalPart),
			Name:                 types.StringValue(mailbox.Name),
			Address:              types.StringValue(mailbox.Address),
			IsInternal:           types.BoolValue(mailbox.IsInternal),
			MaySend:              types.BoolValue(mailbox.MaySend),
			MayReceive:           types.BoolValue(mailbox.MayReceive),
			MayAccessImap:        types.BoolValue(mailbox.MayAccessImap),
			MayAccessPop3:        types.BoolValue(mailbox.MayAccessPop3),
			MayAccessManageSieve: types.BoolValue(mailbox.MayAccessManagesieve),
			StorageUsage:         types.Float64Value(mailbox.StorageUsage),
			ChangedAt:            types.StringValue(mailbox.ChangedAt),
			LastLoginAt:          types.StringValue(mailbox.LastLoginAt),
		})
	}

	mailboxesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"local_part":             types.StringType,
			"name":                   types.StringType,
			"address":                types.StringType,
			"is_internal":            types.BoolType,
			"may_send":               types.BoolType,
			"may_receive":            types.BoolType,
			"may_access_imap":        types.BoolType,
			"may_access_pop3":        types.BoolType,
			"may_access_managesieve": types.BoolType,
			"storage_usage":          types.Float64Type,
			"changed_at":             types.StringType,
			"last_login_at":          types.StringType,
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Mailboxes = mailboxesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
