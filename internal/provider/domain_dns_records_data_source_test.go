package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestNewDomainDNSRecordsDataSourceMetadata(t *testing.T) {
	d := NewDomainDNSRecordsDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_domain_dns_records" {
		t.Fatalf("expected type name %q, got %q", "migadu_domain_dns_records", resp.TypeName)
	}
}

func TestNewDomainDNSRecordsDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewDomainDNSRecordsDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewDomainDNSRecordsDataSourceSchemaExpectations(t *testing.T) {
	d := NewDomainDNSRecordsDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	records := requireDataSourceListNestedAttributeComputed(t, attrs, "records")
	nestedAttrs := records.NestedObject.Attributes

	requireDataSourceStringAttribute(t, nestedAttrs, "type", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "name", false, true)
	requireDataSourceStringAttribute(t, nestedAttrs, "value", false, true)

	priorityAttr, ok := nestedAttrs["priority"]
	if !ok {
		t.Fatal("expected nested attribute records.priority")
	}
	if intAttr, ok := priorityAttr.(datasourceschema.Int64Attribute); !ok || !intAttr.Computed {
		t.Fatal("expected records.priority to be a computed Int64Attribute")
	}

	ttlAttr, ok := nestedAttrs["ttl"]
	if !ok {
		t.Fatal("expected nested attribute records.ttl")
	}
	if intAttr, ok := ttlAttr.(datasourceschema.Int64Attribute); !ok || !intAttr.Computed {
		t.Fatal("expected records.ttl to be a computed Int64Attribute")
	}
}
