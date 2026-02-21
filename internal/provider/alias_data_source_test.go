package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewAliasDataSourceMetadata(t *testing.T) {
	d := NewAliasDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_alias" {
		t.Fatalf("expected type name %q, got %q", "migadu_alias", resp.TypeName)
	}
}

func TestNewAliasDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewAliasDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewAliasDataSourceSchemaExpectations(t *testing.T) {
	d := NewAliasDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "local_part", true, false)
	requireDataSourceStringAttribute(t, attrs, "address", false, true)
	requireDataSourceListAttribute(t, attrs, "destinations", false, true)
	requireDataSourceBoolAttributeComputed(t, attrs, "is_internal")
}
