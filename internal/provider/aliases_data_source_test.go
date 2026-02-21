package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewAliasesDataSourceMetadata(t *testing.T) {
	d := NewAliasesDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_aliases" {
		t.Fatalf("expected type name %q, got %q", "migadu_aliases", resp.TypeName)
	}
}

func TestNewAliasesDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewAliasesDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewAliasesDataSourceSchemaExpectations(t *testing.T) {
	d := NewAliasesDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	aliases := requireDataSourceListNestedAttributeComputed(t, attrs, "aliases")

	nestedAttrs := aliases.NestedObject.Attributes
	requireDataSourceStringAttribute(t, nestedAttrs, "local_part", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "address", false, true)
	requireDataSourceListAttribute(t, nestedAttrs, "destinations", false, true)

	attr, ok := nestedAttrs["is_internal"]
	if !ok {
		t.Fatal("expected nested attribute aliases.is_internal")
	}
	if _, ok := attr.(datasourceschema.BoolAttribute); !ok {
		t.Fatal("expected aliases.is_internal to be a BoolAttribute")
	}
}
