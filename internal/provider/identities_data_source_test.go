package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewIdentitiesDataSourceMetadata(t *testing.T) {
	d := NewIdentitiesDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_identities" {
		t.Fatalf("expected type name %q, got %q", "migadu_identities", resp.TypeName)
	}
}

func TestNewIdentitiesDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewIdentitiesDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewIdentitiesDataSourceSchemaExpectations(t *testing.T) {
	d := NewIdentitiesDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "mailbox", true, false)
	identities := requireDataSourceListNestedAttributeComputed(t, attrs, "identities")

	nestedAttrs := identities.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "local_part", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "address", false, true)
	requireDataSourceBoolAttributeComputed(t, nestedAttrs, "may_send")
}
