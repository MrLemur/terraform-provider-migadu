package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewMailboxesDataSourceMetadata(t *testing.T) {
	d := NewMailboxesDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_mailboxes" {
		t.Fatalf("expected type name %q, got %q", "migadu_mailboxes", resp.TypeName)
	}
}

func TestNewMailboxesDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewMailboxesDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewMailboxesDataSourceSchemaExpectations(t *testing.T) {
	d := NewMailboxesDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	mailboxes := requireDataSourceListNestedAttributeComputed(t, attrs, "mailboxes")

	nestedAttrs := mailboxes.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "local_part", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "address", false, true)

	storageAttr, ok := nestedAttrs["storage_usage"]
	if !ok {
		t.Fatal("expected nested attribute mailboxes.storage_usage")
	}
	if floatAttr, ok := storageAttr.(datasourceschema.Float64Attribute); !ok || !floatAttr.Computed {
		t.Fatal("expected mailboxes.storage_usage to be a computed Float64Attribute")
	}
}
