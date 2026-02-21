package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestNewDomainDiagnosticsDataSourceMetadata(t *testing.T) {
	d := NewDomainDiagnosticsDataSource()

	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "migadu"}, &resp)

	if resp.TypeName != "migadu_domain_diagnostics" {
		t.Fatalf("expected type name %q, got %q", "migadu_domain_diagnostics", resp.TypeName)
	}
}

func TestNewDomainDiagnosticsDataSourceSchemaHasAttributes(t *testing.T) {
	d := NewDomainDiagnosticsDataSource()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Fatal("expected schema attributes to be non-empty")
	}
}

func TestNewDomainDiagnosticsDataSourceSchemaExpectations(t *testing.T) {
	d := NewDomainDiagnosticsDataSource()
	resp := mustDataSourceSchema(t, d)
	attrs := resp.Schema.Attributes

	requireDataSourceStringAttribute(t, attrs, "domain_name", true, false)
	requireDataSourceListAttribute(t, attrs, "mx", false, true)
	requireDataSourceListAttribute(t, attrs, "spf", false, true)
	requireDataSourceListAttribute(t, attrs, "dkim", false, true)
	requireDataSourceListAttribute(t, attrs, "dmarc", false, true)
}
