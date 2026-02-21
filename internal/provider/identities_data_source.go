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

var _ datasource.DataSource = &IdentitiesDataSource{}

func NewIdentitiesDataSource() datasource.DataSource {
	return &IdentitiesDataSource{}
}

type IdentitiesDataSource struct {
	client *migadu.Client
}

type IdentitiesDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	Mailbox    types.String `tfsdk:"mailbox"`
	Identities types.List   `tfsdk:"identities"`
}

type IdentityListItemModel struct {
	LocalPart            types.String `tfsdk:"local_part"`
	Name                 types.String `tfsdk:"name"`
	Address              types.String `tfsdk:"address"`
	MaySend              types.Bool   `tfsdk:"may_send"`
	MayReceive           types.Bool   `tfsdk:"may_receive"`
	MayAccessImap        types.Bool   `tfsdk:"may_access_imap"`
	MayAccessPop3        types.Bool   `tfsdk:"may_access_pop3"`
	MayAccessManageSieve types.Bool   `tfsdk:"may_access_managesieve"`
}

func (d *IdentitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identities"
}

func (d *IdentitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all identities for a mailbox (INDEX operation).",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "The mailbox local part.",
				Required:            true,
			},
			"identities": schema.ListNestedAttribute{
				MarkdownDescription: "List of identities for this mailbox.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"local_part": schema.StringAttribute{
							MarkdownDescription: "The local part of the identity address.",
							Computed:            true,
						},
						"name":        schema.StringAttribute{MarkdownDescription: "Display name for the identity.", Computed: true},
						"address":     schema.StringAttribute{MarkdownDescription: "Full email address.", Computed: true},
						"may_send":    schema.BoolAttribute{MarkdownDescription: "Whether the identity can send emails.", Computed: true},
						"may_receive": schema.BoolAttribute{MarkdownDescription: "Whether the identity can receive emails.", Computed: true},
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
					},
				},
			},
		},
	}
}

func (d *IdentitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IdentitiesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	identities, err := d.client.ListIdentities(ctx, domain, &migadu.Mailbox{LocalPart: data.Mailbox.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list identities, got error: %s", err))
		return
	}

	items := make([]IdentityListItemModel, 0, len(identities))
	for _, identity := range identities {
		items = append(items, IdentityListItemModel{
			LocalPart:            types.StringValue(identity.LocalPart),
			Name:                 types.StringValue(identity.Name),
			Address:              types.StringValue(identity.Address),
			MaySend:              types.BoolValue(identity.MaySend),
			MayReceive:           types.BoolValue(identity.MayReceive),
			MayAccessImap:        types.BoolValue(identity.MayAccessImap),
			MayAccessPop3:        types.BoolValue(identity.MayAccessPop3),
			MayAccessManageSieve: types.BoolValue(identity.MayAccessManagesieve),
		})
	}

	identitiesList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"local_part":             types.StringType,
			"name":                   types.StringType,
			"address":                types.StringType,
			"may_send":               types.BoolType,
			"may_receive":            types.BoolType,
			"may_access_imap":        types.BoolType,
			"may_access_pop3":        types.BoolType,
			"may_access_managesieve": types.BoolType,
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Identities = identitiesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
