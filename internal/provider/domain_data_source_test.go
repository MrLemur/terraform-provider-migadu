package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewDomainDataSourceMetadata(t *testing.T) {
	d := NewDomainDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_domain" {
		t.Fatalf("expected type name %q, got %q", "migadu_domain", resp.TypeName)
	}
}

func TestNewDomainDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewDomainDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewDomainDataSourceSchemaExpectations(t *testing.T) {
	d := NewDomainDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "name", true, false)
	requireDataSourceStringAttribute(t, attrs, "state", false, true)
	requireDataSourceStringAttribute(t, attrs, "description", false, true)
	requireDataSourceListAttribute(t, attrs, "tags", false, true)
}
