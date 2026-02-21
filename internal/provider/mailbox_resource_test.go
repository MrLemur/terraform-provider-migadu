package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewMailboxResourceMetadata(t *testing.T) {
	r := NewMailboxResource()

	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_mailbox" {
		t.Fatalf("expected type name %q, got %q", "migadu_mailbox", resp.TypeName)
	}
}

func TestNewMailboxResourceSchemaHasAttributes(t *testing.T) {
	r := NewMailboxResource()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestMailboxResourceImportState(t *testing.T) {
	r := NewMailboxResource()
	importer, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Fatal("expected mailbox resource to implement ResourceWithImportState")
	}
	schemaResp := mustResourceSchema(t, r)

	t.Run("valid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "example.com/admin"}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected import errors: %v", resp.Diagnostics)
		}

		if got := getStateStringAttribute(t, resp.State, "domain_name"); got != "example.com" {
			t.Fatalf("expected domain_name to be %q, got %q", "example.com", got)
		}

		if got := getStateStringAttribute(t, resp.State, "local_part"); got != "admin" {
			t.Fatalf("expected local_part to be %q, got %q", "admin", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		resp := resource.ImportStateResponse{
			State: newStateForSchema(schemaResp.Schema),
		}

		importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "invalid"}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("expected import parsing error for invalid id")
		}

		assertHasDiagnosticSummary(t, resp.Diagnostics, "Invalid Import ID")
	})
}

func TestMailboxResourcePasswordValidation(t *testing.T) {
	r := NewMailboxResource().(*MailboxResource)
	schemaResp := mustResourceSchema(t, r)

	testCases := map[string]func(t *testing.T, plan tfsdk.Plan) diag.Diagnostics{
		"create": func(t *testing.T, plan tfsdk.Plan) diag.Diagnostics {
			req := resource.CreateRequest{Plan: plan}
			resp := resource.CreateResponse{State: newStateForSchema(schemaResp.Schema)}
			r.Create(context.Background(), req, &resp)
			return resp.Diagnostics
		},
		"update": func(t *testing.T, plan tfsdk.Plan) diag.Diagnostics {
			req := resource.UpdateRequest{Plan: plan}
			resp := resource.UpdateResponse{State: newStateForSchema(schemaResp.Schema)}
			r.Update(context.Background(), req, &resp)
			return resp.Diagnostics
		},
	}

	for name, run := range testCases {
		t.Run(name, func(t *testing.T) {
			plan := newPlanForSchema(schemaResp.Schema)
			diags := plan.Set(context.Background(), &MailboxResourceModel{
				DomainName:            types.StringValue("example.com"),
				LocalPart:             types.StringValue("admin"),
				Name:                  types.StringValue("Admin"),
				PasswordMethod:        types.StringValue("password"),
				Password:              types.StringNull(),
				PasswordRecoveryEmail: types.StringNull(),
				MaySend:               types.BoolValue(true),
				MayReceive:            types.BoolValue(true),
				MayAccessImap:         types.BoolValue(true),
				MayAccessPop3:         types.BoolValue(true),
				MayAccessManageSieve:  types.BoolValue(true),
				SpamAction:            types.StringValue("folder"),
				SpamAggressiveness:    types.StringValue("moderate"),
				FooterActive:          types.BoolValue(false),
				FooterPlainBody:       types.StringNull(),
				FooterHTMLBody:        types.StringNull(),
				Address:               types.StringNull(),
				IsInternal:            types.BoolNull(),
				StorageUsage:          types.Int64Null(),
				ChangedAt:             types.StringNull(),
				LastLoginAt:           types.StringNull(),
			})
			if diags.HasError() {
				t.Fatalf("failed preparing test plan: %v", diags)
			}

			diags = run(t, plan)
			if !diags.HasError() {
				t.Fatal("expected password validation error")
			}

			assertHasDiagnosticSummary(t, diags, "Missing Password")
		})
	}
}
