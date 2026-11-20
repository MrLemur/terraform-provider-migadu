package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestIdentityResourcePasswordUseSchema(t *testing.T) {
	schemaResp := mustResourceSchema(t, NewIdentityResource())

	attr, ok := schemaResp.Schema.Attributes["password_use"].(schema.StringAttribute)
	if !ok {
		t.Fatal("expected password_use to be a string attribute")
	}
	if !attr.Optional || !attr.Computed {
		t.Fatalf("expected password_use to be optional and computed, got optional=%t computed=%t", attr.Optional, attr.Computed)
	}
	if attr.Default != nil {
		t.Fatal("expected password_use to have no default")
	}

	validate := func(v string) bool {
		for _, val := range attr.Validators {
			resp := validator.StringResponse{}
			val.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(v)}, &resp)
			if resp.Diagnostics.HasError() {
				return false
			}
		}
		return true
	}
	for _, v := range []string{"none", "mailbox", "custom"} {
		if !validate(v) {
			t.Errorf("expected %q to be accepted", v)
		}
	}
	if validate("bogus") {
		t.Error("expected \"bogus\" to be rejected")
	}
}
