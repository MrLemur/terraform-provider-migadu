package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNewRewriteResourceMetadata(t *testing.T) {
	r := NewRewriteResource()

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_rewrite" {
		t.Fatalf("expected type name %q, got %q", "migadu_rewrite", resp.TypeName)
	}
}

func TestNewRewriteResourceSchemaHasAttributes(t *testing.T) {
	r := NewRewriteResource()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestRewriteResourceImportState(t *testing.T) {
	r := NewRewriteResource()
	importer, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Fatal("expected rewrite resource to implement ResourceWithImportState")
	}
	schemaResp := mustResourceSchema(t, r)

	t.Run("valid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com/rule-a"}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected import errors: %v", resp.Diagnostics)
		}

		if got := getStateStringAttribute(t, resp.State, "domain_name"); got != "example.com" {
			t.Fatalf("expected domain_name to be %q, got %q", "example.com", got)
		}

		if got := getStateStringAttribute(t, resp.State, "name"); got != "rule-a" {
			t.Fatalf("expected name to be %q, got %q", "rule-a", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com"}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("expected import parsing error for invalid id")
		}

		assertHasDiagnosticSummary(t, resp.Diagnostics, "Invalid Import ID")
	})
}
