package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"migadu": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("MIGADU_USERNAME") == "" {
		t.Fatal("MIGADU_USERNAME must be set for acceptance tests")
	}

	if os.Getenv("MIGADU_API_KEY") == "" {
		t.Fatal("MIGADU_API_KEY must be set for acceptance tests")
	}

	if os.Getenv("MIGADU_TEST_DOMAIN") == "" {
		t.Fatal("MIGADU_TEST_DOMAIN must be set for acceptance tests")
	}
}

func testAccMailboxConfig(localPart string) string {
	return fmt.Sprintf(`
resource "migadu_mailbox" "test" {
  domain_name     = "%s"
  local_part      = "%s"
  name            = "Terraform Acceptance %s"
  password_method = "invitation"
}
`, os.Getenv("MIGADU_TEST_DOMAIN"), localPart, localPart)
}

func testAccAliasConfig(mailboxLocalPart, aliasLocalPart string) string {
	return fmt.Sprintf(`
resource "migadu_mailbox" "dest" {
  domain_name     = "%s"
  local_part      = "%s"
  name            = "Terraform Acceptance Destination %s"
  password_method = "invitation"
}

resource "migadu_alias" "test" {
  domain_name  = "%s"
  local_part   = "%s"
  destinations = [migadu_mailbox.dest.address]
}
`, os.Getenv("MIGADU_TEST_DOMAIN"), mailboxLocalPart, mailboxLocalPart, os.Getenv("MIGADU_TEST_DOMAIN"), aliasLocalPart)
}
