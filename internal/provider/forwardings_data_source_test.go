package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewForwardingsDataSourceMetadata(t *testing.T) {
	d := NewForwardingsDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_forwardings" {
		t.Fatalf("expected type name %q, got %q", "migadu_forwardings", resp.TypeName)
	}
}

func TestNewForwardingsDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewForwardingsDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewForwardingsDataSourceSchemaExpectations(t *testing.T) {
	d := NewForwardingsDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "mailbox", true, false)
	forwardings := requireDataSourceListNestedAttributeComputed(t, attrs, "forwardings")

	nestedAttrs := forwardings.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "address", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "expires_on", false, true)
	requireDataSourceBoolAttributeComputed(t, nestedAttrs, "is_active")
}
