package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RewriteResource{}
var _ resource.ResourceWithImportState = &RewriteResource{}

func NewRewriteResource() resource.Resource {
	return &RewriteResource{}
}

type RewriteResource struct {
	client *migadu.Client
}

type RewriteResourceModel struct {
	DomainName    types.String `tfsdk:"domain_name"`
	Name          types.String `tfsdk:"name"`
	LocalPartRule types.String `tfsdk:"local_part_rule"`
	OrderNum      types.Int64  `tfsdk:"order_num"`
	Destinations  types.List   `tfsdk:"destinations"`
}

func (r *RewriteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite"
}

func (r *RewriteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Migadu rewrite rule for address rewriting.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the rewrite rule.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_part_rule": schema.StringAttribute{
				MarkdownDescription: "The local part matching rule (supports wildcards).",
				Required:            true,
			},
			"order_num": schema.Int64Attribute{
				MarkdownDescription: "Order number for rule processing (lower numbers processed first).",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"destinations": schema.ListAttribute{
				MarkdownDescription: "List of destination email addresses. All destinations must be on the same domain.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *RewriteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RewriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RewriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var destinations []string
	resp.Diagnostics.Append(data.Destinations.ElementsAs(ctx, &destinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rewrite := &migadu.Rewrite{
		Name:          data.Name.ValueString(),
		LocalPartRule: data.LocalPartRule.ValueString(),
		OrderNum:      int(data.OrderNum.ValueInt64()),
		Destinations:  destinations,
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	_, err := r.client.NewRewrite(ctx, domain, rewrite)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rewrite, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RewriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RewriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	name := data.Name.ValueString()

	rewrite, err := r.client.GetRewrite(ctx, domain, &migadu.Rewrite{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rewrite, got error: %s", err))
		return
	}

	data.LocalPartRule = types.StringValue(rewrite.LocalPartRule)
	data.OrderNum = types.Int64Value(int64(rewrite.OrderNum))

	destinations, diags := types.ListValueFrom(ctx, types.StringType, rewrite.Destinations)
	resp.Diagnostics.Append(diags...)
	data.Destinations = destinations

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RewriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RewriteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var destinations []string
	resp.Diagnostics.Append(data.Destinations.ElementsAs(ctx, &destinations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rewrite := &migadu.Rewrite{
		Name:          data.Name.ValueString(),
		LocalPartRule: data.LocalPartRule.ValueString(),
		OrderNum:      int(data.OrderNum.ValueInt64()),
		Destinations:  destinations,
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}

	_, err := r.client.UpdateRewrite(ctx, domain, rewrite)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rewrite, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RewriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RewriteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	rewrite := &migadu.Rewrite{
		Name: data.Name.ValueString(),
	}

	err := r.client.DeleteRewrite(ctx, domain, rewrite)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rewrite, got error: %s", err))
		return
	}
}

func (r *RewriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: domain_name/name
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: domain_name/name",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
