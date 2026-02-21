package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSchemaValidateImplementation(t *testing.T) {
	p := &MigaduProvider{version: "test"}
	resp := mustProviderSchema(t, p)

	diags := resp.Schema.ValidateImplementation(context.Background())
	if diags.HasError() {
		t.Fatalf("provider schema implementation validation failed: %v", diags)
	}
}

func TestResourceSchemasValidateImplementation(t *testing.T) {
	testCases := map[string]func() resource.Resource{
		"domain":   NewDomainResource,
		"mailbox":  NewMailboxResource,
		"alias":    NewAliasResource,
		"identity": NewIdentityResource,
		"rewrite":  NewRewriteResource,
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := mustResourceSchema(t, tc())
			diags := resp.Schema.ValidateImplementation(context.Background())
			if diags.HasError() {
				t.Fatalf("resource schema implementation validation failed: %v", diags)
			}
		})
	}
}

func TestDataSourceSchemasValidateImplementation(t *testing.T) {
	testCases := map[string]func() datasource.DataSource{
		"domain":             NewDomainDataSource,
		"domains":            NewDomainsDataSource,
		"mailbox":            NewMailboxDataSource,
		"mailboxes":          NewMailboxesDataSource,
		"alias":              NewAliasDataSource,
		"aliases":            NewAliasesDataSource,
		"identity":           NewIdentityDataSource,
		"identities":         NewIdentitiesDataSource,
		"rewrite":            NewRewriteDataSource,
		"rewrites":           NewRewritesDataSource,
		"forwardings":        NewForwardingsDataSource,
		"domain_diagnostics": NewDomainDiagnosticsDataSource,
		"domain_dns_records": NewDomainDNSRecordsDataSource,
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := mustDataSourceSchema(t, tc())
			diags := resp.Schema.ValidateImplementation(context.Background())
			if diags.HasError() {
				t.Fatalf("data source schema implementation validation failed: %v", diags)
			}
		})
	}
}
