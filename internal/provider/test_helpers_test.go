package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func mustProviderSchema(t *testing.T, p provider.Provider) provider.SchemaResponse {
	t.Helper()

	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	return resp
}

func mustResourceSchema(t *testing.T, r resource.Resource) resource.SchemaResponse {
	t.Helper()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func mustDataSourceSchema(t *testing.T, d datasource.DataSource) datasource.SchemaResponse {
	t.Helper()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func newConfigFromSchema(schema providerschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	objectType := schema.Type().TerraformType(context.Background())
	raw := tftypes.NewValue(objectType, values)
	return tfsdk.Config{
		Schema: schema,
		Raw:    raw,
	}
}

func newPlanForSchema(schema resourceschema.Schema) tfsdk.Plan {
	objectType := schema.Type().TerraformType(context.Background())
	return tfsdk.Plan{
		Schema: schema,
		Raw:    tftypes.NewValue(objectType, nil),
	}
}

func newStateForSchema(schema resourceschema.Schema) tfsdk.State {
	objectType := schema.Type().TerraformType(context.Background())
	return tfsdk.State{
		Schema: schema,
		Raw:    tftypes.NewValue(objectType, nil),
	}
}

func assertHasDiagnosticSummary(t *testing.T, diags diag.Diagnostics, summary string) {
	t.Helper()

	for _, d := range diags {
		if d.Summary() == summary {
			return
		}
	}

	t.Fatalf("expected diagnostics to include summary %q, got %d diagnostics", summary, len(diags))
}

func getStateStringAttribute(t *testing.T, state tfsdk.State, attr string) string {
	t.Helper()

	var value string
	diags := state.GetAttribute(context.Background(), path.Root(attr), &value)
	if diags.HasError() {
		t.Fatalf("failed reading state attribute %q: %v", attr, diags)
	}

	return value
}

func requireDataSourceStringAttribute(t *testing.T, attrs map[string]datasourceschema.Attribute, name string, required bool, computed bool) {
	t.Helper()

	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected data source schema attribute %q", name)
	}

	stringAttr, ok := attr.(datasourceschema.StringAttribute)
	if !ok {
		t.Fatalf("expected %q to be a StringAttribute", name)
	}

	if stringAttr.Required != required {
		t.Fatalf("expected %q required=%t, got %t", name, required, stringAttr.Required)
	}

	if stringAttr.Computed != computed {
		t.Fatalf("expected %q computed=%t, got %t", name, computed, stringAttr.Computed)
	}
}

func requireDataSourceBoolAttributeComputed(t *testing.T, attrs map[string]datasourceschema.Attribute, name string) {
	t.Helper()

	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected data source schema attribute %q", name)
	}

	boolAttr, ok := attr.(datasourceschema.BoolAttribute)
	if !ok {
		t.Fatalf("expected %q to be a BoolAttribute", name)
	}

	if !boolAttr.Computed {
		t.Fatalf("expected %q to be computed", name)
	}
}

func requireDataSourceListAttribute(t *testing.T, attrs map[string]datasourceschema.Attribute, name string, required bool, computed bool) {
	t.Helper()

	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected data source schema attribute %q", name)
	}

	listAttr, ok := attr.(datasourceschema.ListAttribute)
	if !ok {
		t.Fatalf("expected %q to be a ListAttribute", name)
	}

	if listAttr.Required != required {
		t.Fatalf("expected %q required=%t, got %t", name, required, listAttr.Required)
	}

	if listAttr.Computed != computed {
		t.Fatalf("expected %q computed=%t, got %t", name, computed, listAttr.Computed)
	}
}

func requireDataSourceListNestedAttributeComputed(t *testing.T, attrs map[string]datasourceschema.Attribute, name string) datasourceschema.ListNestedAttribute {
	t.Helper()

	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected data source schema attribute %q", name)
	}

	listAttr, ok := attr.(datasourceschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("expected %q to be a ListNestedAttribute", name)
	}

	if !listAttr.Computed {
		t.Fatalf("expected %q to be computed", name)
	}

	return listAttr
}
