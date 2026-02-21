package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewMailboxDataSourceMetadata(t *testing.T) {
	d := NewMailboxDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_mailbox" {
		t.Fatalf("expected type name %q, got %q", "migadu_mailbox", resp.TypeName)
	}
}

func TestNewMailboxDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewMailboxDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewMailboxDataSourceSchemaExpectations(t *testing.T) {
	d := NewMailboxDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "local_part", true, false)
	requireDataSourceStringAttribute(t, attrs, "address", false, true)
	requireDataSourceBoolAttributeComputed(t, attrs, "is_internal")

	storageAttr, ok := attrs["storage_usage"]
	if !ok {
		t.Fatal("expected schema attribute storage_usage")
	}
	if floatAttr, ok := storageAttr.(datasourceschema.Float64Attribute); !ok || !floatAttr.Computed {
		t.Fatal("expected storage_usage to be a computed Float64Attribute")
	}
}
