package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewRewriteDataSourceMetadata(t *testing.T) {
	d := NewRewriteDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_rewrite" {
		t.Fatalf("expected type name %q, got %q", "migadu_rewrite", resp.TypeName)
	}
}

func TestNewRewriteDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewRewriteDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewRewriteDataSourceSchemaExpectations(t *testing.T) {
	d := NewRewriteDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceStringAttribute(t, attrs, "name", true, false)
	requireDataSourceStringAttribute(t, attrs, "local_part_rule", false, true)
	requireDataSourceListAttribute(t, attrs, "destinations", false, true)

	orderAttr, ok := attrs["order_num"]
	if !ok {
		t.Fatal("expected schema attribute order_num")
	}
	if intAttr, ok := orderAttr.(datasourceschema.Int64Attribute); !ok || !intAttr.Computed {
		t.Fatal("expected order_num to be a computed Int64Attribute")
	}
}
