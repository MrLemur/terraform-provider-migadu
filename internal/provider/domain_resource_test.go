package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewDomainResourceMetadata(t *testing.T) {
	r := NewDomainResource()

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_domain" {
		t.Fatalf("expected type name %q, got %q", "migadu_domain", resp.TypeName)
	}
}

func TestNewDomainResourceSchemaHasAttributes(t *testing.T) {
	r := NewDomainResource()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestDomainResourceImportState(t *testing.T) {
	r := NewDomainResource()
	importer, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Fatal("expected domain resource to implement ResourceWithImportState")
	}
	schemaResp := mustResourceSchema(t, r)

	resp := resource.ImportStateResponse{
		State: newStateForSchema(schemaResp.Schema),
	}

	importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com"}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected import errors: %v", resp.Diagnostics)
	}

	if got := getStateStringAttribute(t, resp.State, "name"); got != "example.com" {
		t.Fatalf("expected name to be %q, got %q", "example.com", got)
	}
}

func TestDomainResourceDeleteAddsWarning(t *testing.T) {
	r := NewDomainResource()
	schemaResp := mustResourceSchema(t, r)

	state := newStateForSchema(schemaResp.Schema)
	diags := state.Set(context.Background(), &DomainResourceModel{
		Name:                 types.StringValue("example.com"),
		State:                types.StringNull(),
		Description:          types.StringNull(),
		Tags:                 types.ListNull(types.StringType),
		SpamAggressiveness:   types.Int64Null(),
		GreylistingEnabled:   types.BoolNull(),
		MXProxyEnabled:       types.BoolNull(),
		HostedDNS:            types.BoolNull(),
		SenderDenylist:       types.ListNull(types.StringType),
		RecipientDenylist:    types.ListNull(types.StringType),
		CatchallDestinations: types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("failed preparing delete state: %v", diags)
	}

	req := resource.DeleteRequest{State: state}
	var resp resource.DeleteResponse

	r.Delete(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected delete errors: %v", resp.Diagnostics)
	}

	if resp.Diagnostics.WarningsCount() != 1 {
		t.Fatalf("expected exactly 1 warning, got %d", resp.Diagnostics.WarningsCount())
	}

	assertHasDiagnosticSummary(t, resp.Diagnostics, "Domain Deletion Not Supported")
}
