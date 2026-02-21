package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewRewritesDataSourceMetadata(t *testing.T) {
	d := NewRewritesDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_rewrites" {
		t.Fatalf("expected type name %q, got %q", "migadu_rewrites", resp.TypeName)
	}
}

func TestNewRewritesDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewRewritesDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewRewritesDataSourceSchemaExpectations(t *testing.T) {
	d := NewRewritesDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	rewrites := requireDataSourceListNestedAttributeComputed(t, attrs, "rewrites")

	nestedAttrs := rewrites.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "name", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "local_part_rule", false, true)
	requireDataSourceListAttribute(t, nestedAttrs, "destinations", false, true)

	orderAttr, ok := nestedAttrs["order_num"]
	if !ok {
		t.Fatal("expected nested attribute rewrites.order_num")
	}
	if intAttr, ok := orderAttr.(datasourceschema.Int64Attribute); !ok || !intAttr.Computed {
		t.Fatal("expected rewrites.order_num to be a computed Int64Attribute")
	}
}
