package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewIdentityDataSourceMetadata(t *testing.T) {
	d := NewIdentityDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_identity" {
		t.Fatalf("expected type name %q, got %q", "migadu_identity", resp.TypeName)
	}
}

func TestNewIdentityDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewIdentityDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewIdentityDataSourceSchemaExpectations(t *testing.T) {
	d := NewIdentityDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "mailbox", true, false)
	requireDataSourceStringAttribute(t, attrs, "local_part", true, false)
	requireDataSourceStringAttribute(t, attrs, "address", false, true)
	requireDataSourceBoolAttributeComputed(t, attrs, "may_send")
}
