package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNewIdentityResourceMetadata(t *testing.T) {
	r := NewIdentityResource()

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_identity" {
		t.Fatalf("expected type name %q, got %q", "migadu_identity", resp.TypeName)
	}
}

func TestNewIdentityResourceSchemaHasAttributes(t *testing.T) {
	r := NewIdentityResource()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestIdentityResourceImportState(t *testing.T) {
	r := NewIdentityResource()
	importer, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Fatal("expected identity resource to implement ResourceWithImportState")
	}
	schemaResp := mustResourceSchema(t, r)

	t.Run("valid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com/mailbox/admin"}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected import errors: %v", resp.Diagnostics)
		}

		if got := getStateStringAttribute(t, resp.State, "domain_name"); got != "example.com" {
			t.Fatalf("expected domain_name to be %q, got %q", "example.com", got)
		}

		if got := getStateStringAttribute(t, resp.State, "mailbox"); got != "mailbox" {
			t.Fatalf("expected mailbox to be %q, got %q", "mailbox", got)
		}

		if got := getStateStringAttribute(t, resp.State, "local_part"); got != "admin" {
			t.Fatalf("expected local_part to be %q, got %q", "admin", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com/admin"}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("expected import parsing error for invalid id")
		}

		assertHasDiagnosticSummary(t, resp.Diagnostics, "Invalid Import ID")
	})
}
