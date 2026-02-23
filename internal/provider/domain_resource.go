package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DomainResource{}
var _ resource.ResourceWithImportState = &DomainResource{}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	client *migadu.Client
}

type DomainResourceModel struct {
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

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Migadu domain.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Domain state (computed).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Domain description.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Domain tags.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"spam_aggressiveness": schema.Int64Attribute{
				MarkdownDescription: "Spam filter aggressiveness level for the domain (integer). " +
					"Valid values: `-3` (paranoid/most aggressive), `-2` (aggressive), " +
					"`0` (moderate, default), `2` (suspicious), `3` (permissive/least aggressive).",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"greylisting_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether greylisting is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mx_proxy_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether MX proxy is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"hosted_dns": schema.BoolAttribute{
				MarkdownDescription: "Whether DNS is hosted by Migadu. Setting this to `true` is not supported — Migadu plans to discontinue this service and the API will reject it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sender_denylist": schema.ListAttribute{
				MarkdownDescription: "List of denied sender addresses.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"recipient_denylist": schema.ListAttribute{
				MarkdownDescription: "List of denied recipient addresses.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"catchall_destinations": schema.ListAttribute{
				MarkdownDescription: "Catchall email destinations.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags, senderDenylist, recipientDenylistSlice, catchallDestinations []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	resp.Diagnostics.Append(data.SenderDenylist.ElementsAs(ctx, &senderDenylist, false)...)
	resp.Diagnostics.Append(data.RecipientDenylist.ElementsAs(ctx, &recipientDenylistSlice, false)...)
	resp.Diagnostics.Append(data.CatchallDestinations.ElementsAs(ctx, &catchallDestinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		Tags:                 tags,
		SpamAggressiveness:   int(data.SpamAggressiveness.ValueInt64()),
		GreylistingEnabled:   data.GreylistingEnabled.ValueBool(),
		MXProxyEnabled:       data.MXProxyEnabled.ValueBool(),
		HostedDNS:            data.HostedDNS.ValueBool(),
		SenderDenylist:       senderDenylist,
		RecipientDenylist:    recipientDenylistSlice,
		CatchallDestinations: catchallDestinations,
	}

	created, err := r.client.NewDomain(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create domain, got error: %s", err))
		return
	}

	data.State = types.StringValue(created.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.Name.ValueString()}
	retrieved, err := r.client.GetDomain(ctx, domain)
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

	tags, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(retrieved.Tags))
	resp.Diagnostics.Append(diags...)
	data.Tags = tags

	senderDenylist, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(retrieved.SenderDenylist))
	resp.Diagnostics.Append(diags...)
	data.SenderDenylist = senderDenylist

	recipientDenylist, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(retrieved.RecipientDenylist))
	resp.Diagnostics.Append(diags...)
	data.RecipientDenylist = recipientDenylist

	catchallDestinations, diags := types.ListValueFrom(ctx, types.StringType, normalizeStringSlice(retrieved.CatchallDestinations))
	resp.Diagnostics.Append(diags...)
	data.CatchallDestinations = catchallDestinations

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags, senderDenylist, recipientDenylistSlice, catchallDestinations []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	resp.Diagnostics.Append(data.SenderDenylist.ElementsAs(ctx, &senderDenylist, false)...)
	resp.Diagnostics.Append(data.RecipientDenylist.ElementsAs(ctx, &recipientDenylistSlice, false)...)
	resp.Diagnostics.Append(data.CatchallDestinations.ElementsAs(ctx, &catchallDestinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		Tags:                 tags,
		SpamAggressiveness:   int(data.SpamAggressiveness.ValueInt64()),
		GreylistingEnabled:   data.GreylistingEnabled.ValueBool(),
		MXProxyEnabled:       data.MXProxyEnabled.ValueBool(),
		HostedDNS:            data.HostedDNS.ValueBool(),
		SenderDenylist:       senderDenylist,
		RecipientDenylist:    recipientDenylistSlice,
		CatchallDestinations: catchallDestinations,
	}

	_, err := r.client.UpdateDomain(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update domain, got error: %s", err))
		return
	}

	// Note: state is intentionally not updated as it is a computed field that causes
	// provider consistency errors. It will refresh on the next read.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: The migadu-go library doesn't have a DeleteDomain method
	// Domains typically can't be deleted via API, only deactivated
	resp.Diagnostics.AddWarning(
		"Domain Deletion Not Supported",
		"The Migadu API does not support domain deletion. Remove this resource from state manually or contact Migadu support to delete the domain.",
	)
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func normalizeStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

