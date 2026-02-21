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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MailboxResource{}
var _ resource.ResourceWithImportState = &MailboxResource{}

func NewMailboxResource() resource.Resource {
	return &MailboxResource{}
}

// MailboxResource defines the resource implementation.
type MailboxResource struct {
	client *migadu.Client
}

// MailboxResourceModel describes the resource data model.
type MailboxResourceModel struct {
	DomainName            types.String `tfsdk:"domain_name"`
	LocalPart             types.String `tfsdk:"local_part"`
	Name                  types.String `tfsdk:"name"`
	PasswordMethod        types.String `tfsdk:"password_method"`
	Password              types.String `tfsdk:"password"`
	PasswordRecoveryEmail types.String `tfsdk:"password_recovery_email"`
	MaySend               types.Bool   `tfsdk:"may_send"`
	MayReceive            types.Bool   `tfsdk:"may_receive"`
	MayAccessImap         types.Bool   `tfsdk:"may_access_imap"`
	MayAccessPop3         types.Bool   `tfsdk:"may_access_pop3"`
	MayAccessManageSieve  types.Bool   `tfsdk:"may_access_managesieve"`
	SpamAction            types.String `tfsdk:"spam_action"`
	SpamAggressiveness    types.String `tfsdk:"spam_aggressiveness"`
	FooterActive          types.Bool   `tfsdk:"footer_active"`
	FooterPlainBody       types.String `tfsdk:"footer_plain_body"`
	FooterHTMLBody        types.String `tfsdk:"footer_html_body"`
	Address               types.String `tfsdk:"address"`
	IsInternal            types.Bool   `tfsdk:"is_internal"`
	StorageUsage          types.Int64  `tfsdk:"storage_usage"`
	ChangedAt             types.String `tfsdk:"changed_at"`
	LastLoginAt           types.String `tfsdk:"last_login_at"`
}

func (r *MailboxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox"
}

func (r *MailboxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Migadu mailbox.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name for this mailbox.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_part": schema.StringAttribute{
				MarkdownDescription: "The local part of the email address (before the @).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name for the mailbox.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"password_method": schema.StringAttribute{
				MarkdownDescription: "Password method: `password` or `invitation`. Defaults to `invitation` if omitted on create.\n\n" +
					"- `password`: `password` is required; `password_recovery_email` is ignored.\n" +
					"- `invitation`: `password_recovery_email` is required; `password` must not be set.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for the mailbox. Required if password_method is 'password'.",
				Optional:            true,
				Sensitive:           true,
			},
			"password_recovery_email": schema.StringAttribute{
				MarkdownDescription: "Recovery email address for password resets. Required when `password_method` is `invitation`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"may_send": schema.BoolAttribute{
				MarkdownDescription: "Whether the mailbox can send emails.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"may_receive": schema.BoolAttribute{
				MarkdownDescription: "Whether the mailbox can receive emails.",
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
			"spam_action": schema.StringAttribute{
				MarkdownDescription: "Action for spam emails: 'folder' or 'delete'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("folder"),
			},
			"spam_aggressiveness": schema.StringAttribute{
				MarkdownDescription: "Spam filter aggressiveness level for the mailbox. Valid values (most to least aggressive):\n\n" +
					"`strictest`, `stricter`, `strict`, `default` (use domain setting), `permissive`, `more permissive`, `most permissive`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("default"),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_internal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an internal mailbox (computed).",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_usage": schema.Int64Attribute{
				MarkdownDescription: "Storage usage in bytes (computed).",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"changed_at": schema.StringAttribute{
				MarkdownDescription: "Last modification timestamp (computed).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_login_at": schema.StringAttribute{
				MarkdownDescription: "Last login timestamp (computed).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MailboxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*migadu.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *migadu.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MailboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MailboxResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	passwordMethod := "invitation"
	if !data.PasswordMethod.IsNull() && !data.PasswordMethod.IsUnknown() {
		passwordMethod = data.PasswordMethod.ValueString()
	}
	if passwordMethod == "password" && data.Password.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Password",
			"When password_method is 'password', the password field must be provided.",
		)
		return
	}

	if err := validateMailboxSpam(data.SpamAggressiveness.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid spam_aggressiveness", err.Error())
		return
	}

	// Create API request body
	mailbox := &migadu.Mailbox{
		LocalPart:             data.LocalPart.ValueString(),
		Name:                  data.Name.ValueString(),
		PasswordMethod:        passwordMethod,
		PasswordRecoveryEmail: data.PasswordRecoveryEmail.ValueString(),
		MaySend:               data.MaySend.ValueBool(),
		MayReceive:            data.MayReceive.ValueBool(),
		MayAccessImap:         data.MayAccessImap.ValueBool(),
		MayAccessPop3:         data.MayAccessPop3.ValueBool(),
		MayAccessManagesieve:  data.MayAccessManageSieve.ValueBool(),
		SpamAction:            data.SpamAction.ValueString(),
		SpamAggressiveness:    data.SpamAggressiveness.ValueString(),
		FooterActive:          data.FooterActive.ValueBool(),
		FooterPlainBody:       data.FooterPlainBody.ValueString(),
		FooterHTMLBody:        data.FooterHTMLBody.ValueString(),
	}
	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		mailbox.Password = data.Password.ValueString()
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	// Create the mailbox
	created, err := r.client.NewMailbox(ctx, domain, mailbox)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create mailbox, got error: %s", err))
		return
	}

	// Update the state with the created mailbox data
	data.PasswordMethod = types.StringValue(passwordMethod)
	data.Address = types.StringValue(created.Address)
	data.IsInternal = types.BoolValue(created.IsInternal)
	data.StorageUsage = types.Int64Value(int64(created.StorageUsage))
	data.ChangedAt = types.StringValue(created.ChangedAt)
	data.LastLoginAt = types.StringValue(created.LastLoginAt)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MailboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MailboxResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	localPart := data.LocalPart.ValueString()

	// Get current state from API
	mailbox, err := r.client.GetMailbox(ctx, domain, &migadu.Mailbox{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read mailbox, got error: %s", err))
		return
	}

	// Update state with API data
	data.Name = types.StringValue(mailbox.Name)
	data.PasswordRecoveryEmail = types.StringValue(mailbox.PasswordRecoveryEmail)
	data.MaySend = types.BoolValue(mailbox.MaySend)
	data.MayReceive = types.BoolValue(mailbox.MayReceive)
	data.MayAccessImap = types.BoolValue(mailbox.MayAccessImap)
	data.MayAccessPop3 = types.BoolValue(mailbox.MayAccessPop3)
	data.MayAccessManageSieve = types.BoolValue(mailbox.MayAccessManagesieve)
	data.SpamAction = types.StringValue(mailbox.SpamAction)
	data.SpamAggressiveness = types.StringValue(mailbox.SpamAggressiveness)
	data.FooterActive = types.BoolValue(mailbox.FooterActive)
	data.FooterPlainBody = types.StringValue(mailbox.FooterPlainBody)
	data.FooterHTMLBody = types.StringValue(mailbox.FooterHTMLBody)
	data.Address = types.StringValue(mailbox.Address)
	data.IsInternal = types.BoolValue(mailbox.IsInternal)
	data.StorageUsage = types.Int64Value(int64(mailbox.StorageUsage))
	data.ChangedAt = types.StringValue(mailbox.ChangedAt)
	data.LastLoginAt = types.StringValue(mailbox.LastLoginAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MailboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MailboxResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	passwordMethodIsSet := !data.PasswordMethod.IsNull() && !data.PasswordMethod.IsUnknown()
	if passwordMethodIsSet && data.PasswordMethod.ValueString() == "password" && data.Password.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Password",
			"When password_method is 'password', the password field must be provided.",
		)
		return
	}

	if err := validateMailboxSpam(data.SpamAggressiveness.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid spam_aggressiveness", err.Error())
		return
	}

	// Create API request body
	mailbox := &migadu.Mailbox{
		LocalPart:             data.LocalPart.ValueString(),
		Name:                  data.Name.ValueString(),
		PasswordRecoveryEmail: data.PasswordRecoveryEmail.ValueString(),
		MaySend:               data.MaySend.ValueBool(),
		MayReceive:            data.MayReceive.ValueBool(),
		MayAccessImap:         data.MayAccessImap.ValueBool(),
		MayAccessPop3:         data.MayAccessPop3.ValueBool(),
		MayAccessManagesieve:  data.MayAccessManageSieve.ValueBool(),
		SpamAction:            data.SpamAction.ValueString(),
		SpamAggressiveness:    data.SpamAggressiveness.ValueString(),
		FooterActive:          data.FooterActive.ValueBool(),
		FooterPlainBody:       data.FooterPlainBody.ValueString(),
		FooterHTMLBody:        data.FooterHTMLBody.ValueString(),
	}
	if passwordMethodIsSet {
		mailbox.PasswordMethod = data.PasswordMethod.ValueString()
	}
	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		mailbox.Password = data.Password.ValueString()
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	// Update the mailbox
	updated, err := r.client.UpdateMailbox(ctx, domain, mailbox)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update mailbox, got error: %s", err))
		return
	}

	// Update the state with the updated mailbox data
	// Note: changed_at and last_login_at are intentionally not updated as they cause
	// provider consistency errors. These computed values will refresh on the next read.
	data.Address = types.StringValue(updated.Address)
	data.IsInternal = types.BoolValue(updated.IsInternal)
	data.StorageUsage = types.Int64Value(int64(updated.StorageUsage))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MailboxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MailboxResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	mailbox := &migadu.Mailbox{
		LocalPart: data.LocalPart.ValueString(),
	}

	// Delete the mailbox
	err := r.client.DeleteMailbox(ctx, domain, mailbox)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete mailbox, got error: %s", err))
		return
	}
}

// validateMailboxSpam checks that level is a valid Migadu mailbox spam aggressiveness string.
// See the Mailbox GoDoc for valid levels.
func validateMailboxSpam(level string) error {
	valid := map[string]bool{
		"strictest":       true,
		"stricter":        true,
		"strict":          true,
		"default":         true,
		"permissive":      true,
		"more permissive": true,
		"most permissive": true,
	}
	if !valid[level] {
		return fmt.Errorf("invalid mailbox spam_aggressiveness %q: must be one of strictest, stricter, strict, default, permissive, \"more permissive\", \"most permissive\"", level)
	}
	return nil
}

func (r *MailboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: domain_name/local_part
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: domain_name/local_part",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("local_part"), parts[1])...)
}
