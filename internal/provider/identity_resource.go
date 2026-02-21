package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IdentityResource{}
var _ resource.ResourceWithImportState = &IdentityResource{}

func NewIdentityResource() resource.Resource {
	return &IdentityResource{}
}

type IdentityResource struct {
	client *migadu.Client
}

type IdentityResourceModel struct {
	DomainName           types.String `tfsdk:"domain_name"`
	Mailbox              types.String `tfsdk:"mailbox"`
	LocalPart            types.String `tfsdk:"local_part"`
	Name                 types.String `tfsdk:"name"`
	Password             types.String `tfsdk:"password"`
	MaySend              types.Bool   `tfsdk:"may_send"`
	MayReceive           types.Bool   `tfsdk:"may_receive"`
	MayAccessImap        types.Bool   `tfsdk:"may_access_imap"`
	MayAccessPop3        types.Bool   `tfsdk:"may_access_pop3"`
	MayAccessManageSieve types.Bool   `tfsdk:"may_access_managesieve"`
	FooterActive         types.Bool   `tfsdk:"footer_active"`
	FooterPlainBody      types.String `tfsdk:"footer_plain_body"`
	FooterHTMLBody       types.String `tfsdk:"footer_html_body"`
	Address              types.String `tfsdk:"address"`
}

func (r *IdentityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (r *IdentityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Migadu mailbox identity (sender address).",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "The mailbox local part this identity belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_part": schema.StringAttribute{
				MarkdownDescription: "The local part of the identity address.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the identity.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the identity.",
				Optional:            true,
				Sensitive:           true,
			},
			"may_send": schema.BoolAttribute{
				MarkdownDescription: "Whether the identity can send emails.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"may_receive": schema.BoolAttribute{
				MarkdownDescription: "Whether the identity can receive emails.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"may_access_imap": schema.BoolAttribute{
				MarkdownDescription: "Whether IMAP access is allowed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"may_access_pop3": schema.BoolAttribute{
				MarkdownDescription: "Whether POP3 access is allowed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"may_access_managesieve": schema.BoolAttribute{
				MarkdownDescription: "Whether ManageSieve access is allowed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"footer_active": schema.BoolAttribute{
				MarkdownDescription: "Whether email footer is active.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"footer_plain_body": schema.StringAttribute{
				MarkdownDescription: "Plain text email footer.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"footer_html_body": schema.StringAttribute{
				MarkdownDescription: "HTML email footer.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Full email address (computed).",
				Computed:            true,
			},
		},
	}
}

func (r *IdentityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*migadu.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *migadu.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *IdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := &migadu.Identity{
		LocalPart:            data.LocalPart.ValueString(),
		Name:                 data.Name.ValueString(),
		Password:             data.Password.ValueString(),
		MaySend:              data.MaySend.ValueBool(),
		MayReceive:           data.MayReceive.ValueBool(),
		MayAccessImap:        data.MayAccessImap.ValueBool(),
		MayAccessPop3:        data.MayAccessPop3.ValueBool(),
		MayAccessManagesieve: data.MayAccessManageSieve.ValueBool(),
		FooterActive:         data.FooterActive.ValueBool(),
		FooterPlainBody:      data.FooterPlainBody.ValueString(),
		FooterHTMLBody:       data.FooterHTMLBody.ValueString(),
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxStr := data.Mailbox.ValueString()

	created, err := r.client.NewIdentity(ctx, domain, &migadu.Mailbox{LocalPart: mailboxStr}, identity)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create identity, got error: %s", err))
		return
	}

	data.Address = types.StringValue(created.Address)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxStr := data.Mailbox.ValueString()
	localPart := data.LocalPart.ValueString()

	identity, err := r.client.GetIdentity(ctx, domain, &migadu.Mailbox{LocalPart: mailboxStr}, &migadu.Identity{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read identity, got error: %s", err))
		return
	}

	data.Name = types.StringValue(identity.Name)
	data.MaySend = types.BoolValue(identity.MaySend)
	data.MayReceive = types.BoolValue(identity.MayReceive)
	data.MayAccessImap = types.BoolValue(identity.MayAccessImap)
	data.MayAccessPop3 = types.BoolValue(identity.MayAccessPop3)
	data.MayAccessManageSieve = types.BoolValue(identity.MayAccessManagesieve)
	data.FooterActive = types.BoolValue(identity.FooterActive)
	data.FooterPlainBody = types.StringValue(identity.FooterPlainBody)
	data.FooterHTMLBody = types.StringValue(identity.FooterHTMLBody)
	data.Address = types.StringValue(identity.Address)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := &migadu.Identity{
		LocalPart:            data.LocalPart.ValueString(),
		Name:                 data.Name.ValueString(),
		Password:             data.Password.ValueString(),
		MaySend:              data.MaySend.ValueBool(),
		MayReceive:           data.MayReceive.ValueBool(),
		MayAccessImap:        data.MayAccessImap.ValueBool(),
		MayAccessPop3:        data.MayAccessPop3.ValueBool(),
		MayAccessManagesieve: data.MayAccessManageSieve.ValueBool(),
		FooterActive:         data.FooterActive.ValueBool(),
		FooterPlainBody:      data.FooterPlainBody.ValueString(),
		FooterHTMLBody:       data.FooterHTMLBody.ValueString(),
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxStr := data.Mailbox.ValueString()

	updated, err := r.client.UpdateIdentity(ctx, domain, &migadu.Mailbox{LocalPart: mailboxStr}, identity)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update identity, got error: %s", err))
		return
	}

	data.Address = types.StringValue(updated.Address)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailboxStr := data.Mailbox.ValueString()
	identity := &migadu.Identity{
		LocalPart: data.LocalPart.ValueString(),
	}

	err := r.client.DeleteIdentity(ctx, domain, &migadu.Mailbox{LocalPart: mailboxStr}, identity)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete identity, got error: %s", err))
		return
	}
}

func (r *IdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: domain_name/mailbox/local_part
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: domain_name/mailbox/local_part",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mailbox"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("local_part"), parts[2])...)
}
