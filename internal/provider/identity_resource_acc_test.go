package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIdentityResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	resourceName := "migadu_identity.test"
	mailboxLocalPart := fmt.Sprintf("tfacc-mbx-%d", time.Now().UnixNano())
	identityLocalPart := fmt.Sprintf("tfacc-id-%d", time.Now().UnixNano())
	importID := fmt.Sprintf("%s/%s/%s", os.Getenv("MIGADU_TEST_DOMAIN"), mailboxLocalPart, identityLocalPart)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityConfig(mailboxLocalPart, identityLocalPart),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", os.Getenv("MIGADU_TEST_DOMAIN")),
					resource.TestCheckResourceAttr(resourceName, "mailbox", mailboxLocalPart),
					resource.TestCheckResourceAttr(resourceName, "local_part", identityLocalPart),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     importID,
				// password is write-only; exclude from import verification
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccIdentityConfig(mailboxLocalPart, identityLocalPart string) string {
	return fmt.Sprintf(`
resource "migadu_mailbox" "owner" {
  domain_name     = "%s"
  local_part      = "%s"
  name            = "Terraform Acceptance Owner %s"
  password_method = "invitation"
}

resource "migadu_identity" "test" {
  domain_name = "%s"
  mailbox     = migadu_mailbox.owner.local_part
  local_part  = "%s"
  name        = "Terraform Acceptance Identity"
}
`, os.Getenv("MIGADU_TEST_DOMAIN"), mailboxLocalPart, mailboxLocalPart,
		os.Getenv("MIGADU_TEST_DOMAIN"), identityLocalPart)
}
