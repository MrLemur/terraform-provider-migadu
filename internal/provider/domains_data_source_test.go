package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewDomainsDataSourceMetadata(t *testing.T) {
	d := NewDomainsDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_domains" {
		t.Fatalf("expected type name %q, got %q", "migadu_domains", resp.TypeName)
	}
}

func TestNewDomainsDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewDomainsDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewDomainsDataSourceSchemaExpectations(t *testing.T) {
	d := NewDomainsDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	domains := requireDataSourceListNestedAttributeComputed(t, attrs, "domains")
	nestedAttrs := domains.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "name", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "state", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "description", false, true)
}
