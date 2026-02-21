package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AliasResource{}
var _ resource.ResourceWithImportState = &AliasResource{}

func NewAliasResource() resource.Resource {
	return &AliasResource{}
}

// AliasResource defines the resource implementation.
type AliasResource struct {
	client *migadu.Client
}

// AliasResourceModel describes the resource data model.
type AliasResourceModel struct {
	DomainName   types.String `tfsdk:"domain_name"`
	LocalPart    types.String `tfsdk:"local_part"`
	Destinations types.List   `tfsdk:"destinations"`
	Address      types.String `tfsdk:"address"`
	IsInternal   types.Bool   `tfsdk:"is_internal"`
}

func (r *AliasResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alias"
}

func (r *AliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Migadu email alias.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name for this alias.",
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
			"destinations": schema.ListAttribute{
				MarkdownDescription: "List of destination email addresses for this alias. All destinations must be on the same domain as the alias.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Full email address (computed).",
				Computed:            true,
			},
			"is_internal": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an internal alias (computed).",
				Computed:            true,
			},
		},
	}
}

func (r *AliasResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AliasResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert destinations from types.List to []string
	var destinations []string
	resp.Diagnostics.Append(data.Destinations.ElementsAs(ctx, &destinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body
	alias := &migadu.Alias{
		LocalPart:    data.LocalPart.ValueString(),
		Destinations: destinations,
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	// Create the alias
	created, err := r.client.NewAlias(ctx, domain, alias)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alias, got error: %s", err))
		return
	}

	// Update the state with the created alias data
	data.Address = types.StringValue(created.Address)
	data.IsInternal = types.BoolValue(created.IsInternal)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AliasResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	localPart := data.LocalPart.ValueString()

	// Get current state from API
	alias, err := r.client.GetAlias(ctx, domain, &migadu.Alias{LocalPart: localPart})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alias, got error: %s", err))
		return
	}

	// Convert destinations to types.List
	destinations, diags := types.ListValueFrom(ctx, types.StringType, alias.Destinations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state with API data
	data.Destinations = destinations
	data.Address = types.StringValue(alias.Address)
	data.IsInternal = types.BoolValue(alias.IsInternal)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AliasResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert destinations from types.List to []string
	var destinations []string
	resp.Diagnostics.Append(data.Destinations.ElementsAs(ctx, &destinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body
	alias := &migadu.Alias{
		LocalPart:    data.LocalPart.ValueString(),
		Destinations: destinations,
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	// Update the alias
	updated, err := r.client.UpdateAlias(ctx, domain, alias)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alias, got error: %s", err))
		return
	}

	// Update the state with the updated alias data
	data.Address = types.StringValue(updated.Address)
	data.IsInternal = types.BoolValue(updated.IsInternal)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AliasResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	alias := &migadu.Alias{
		LocalPart: data.LocalPart.ValueString(),
	}

	// Delete the alias
	err := r.client.DeleteAlias(ctx, domain, alias)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alias, got error: %s", err))
		return
	}
}

func (r *AliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
