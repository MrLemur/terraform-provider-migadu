package provider

import (
	"context"
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMigaduProviderConfigure(t *testing.T) {
	p := &MigaduProvider{version: "test"}
	schemaResp := mustProviderSchema(t, p)

	testCases := map[string]struct {
		username             tftypes.Value
		apiKey               tftypes.Value
		envUsername          string
		envAPIKey            string
		expectedErrSummaries []string
		expectClient         bool
	}{
		"missing username": {
			username: tftypes.NewValue(tftypes.String, nil),
			apiKey:   tftypes.NewValue(tftypes.String, "api-key"),
			expectedErrSummaries: []string{
				"Missing Migadu Username",
			},
		},
		"missing api key": {
			username: tftypes.NewValue(tftypes.String, "admin@example.com"),
			apiKey:   tftypes.NewValue(tftypes.String, nil),
			expectedErrSummaries: []string{
				"Missing Migadu API Key",
			},
		},
		"env fallback": {
			username:     tftypes.NewValue(tftypes.String, nil),
			apiKey:       tftypes.NewValue(tftypes.String, nil),
			envUsername:  "admin@example.com",
			envAPIKey:    "env-api-key",
			expectClient: true,
		},
		"unknown config values": {
			username: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			apiKey:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedErrSummaries: []string{
				"Unknown Migadu Username",
				"Unknown Migadu API Key",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Setenv("MIGADU_USERNAME", tc.envUsername)
			t.Setenv("MIGADU_API_KEY", tc.envAPIKey)

			req := frameworkprovider.ConfigureRequest{
				Config: newConfigFromSchema(schemaResp.Schema, map[string]tftypes.Value{
					"username": tc.username,
					"api_key":  tc.apiKey,
				}),
			}

			var resp frameworkprovider.ConfigureResponse
			p.Configure(context.Background(), req, &resp)

			if tc.expectClient {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected configure errors: %v", resp.Diagnostics)
				}

				if resp.ResourceData == nil || resp.DataSourceData == nil {
					t.Fatal("expected provider clients to be set on successful configure")
				}
				return
			}

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected configure errors, got none")
			}

			for _, summary := range tc.expectedErrSummaries {
				assertHasDiagnosticSummary(t, resp.Diagnostics, summary)
			}
		})
	}
}
