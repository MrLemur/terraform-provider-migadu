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

var _ resource.Resource = &DomainActivationResource{}
var _ resource.ResourceWithImportState = &DomainActivationResource{}

func NewDomainActivationResource() resource.Resource {
	return &DomainActivationResource{}
}

type DomainActivationResource struct {
	client *migadu.Client
}

type DomainActivationResourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	State      types.String `tfsdk:"state"`
}

func (r *DomainActivationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_activation"
}

func (r *DomainActivationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Activates a Migadu domain once DNS records are in place.\n\n" +
			"~> **Note:** DNS records (MX, SPF, DKIM, DMARC) must be valid before this resource will apply successfully.\n\n" +
			"-> **Note:** Destroying this resource does not deactivate the domain.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name to activate.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Domain state after activation.",
				Computed:            true,
			},
		},
	}
}

func (r *DomainActivationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainActivationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainActivationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.DomainName.ValueString()

	diag, err := r.client.GetDomainDiagnostics(ctx, &migadu.Domain{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get domain diagnostics, got error: %s", err))
		return
	}

	var issues []string
	if len(diag.MX) > 0 {
		issues = append(issues, fmt.Sprintf("MX: %s", strings.Join(diag.MX, "; ")))
	}
	if len(diag.SPF) > 0 {
		issues = append(issues, fmt.Sprintf("SPF: %s", strings.Join(diag.SPF, "; ")))
	}
	if len(diag.DKIM) > 0 {
		issues = append(issues, fmt.Sprintf("DKIM: %s", strings.Join(diag.DKIM, "; ")))
	}
	if len(diag.DMARC) > 0 {
		issues = append(issues, fmt.Sprintf("DMARC: %s", strings.Join(diag.DMARC, "; ")))
	}
	if len(issues) > 0 {
		resp.Diagnostics.AddError(
			"DNS Validation Failed",
			"DNS records are not valid:\n"+strings.Join(issues, "\n"),
		)
		return
	}

	domain, err := r.client.ActivateDomain(ctx, &migadu.Domain{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to activate domain, got error: %s", err))
		return
	}

	data.State = types.StringValue(domain.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainActivationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(ctx, &migadu.Domain{Name: data.DomainName.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domain, got error: %s", err))
		return
	}

	data.State = types.StringValue(domain.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainActivationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// domain_name is RequiresReplace and state is Computed — no update path needed
}

func (r *DomainActivationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deactivation is not supported by the Migadu API; remove from state only
}

func (r *DomainActivationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_name"), req, resp)
}
