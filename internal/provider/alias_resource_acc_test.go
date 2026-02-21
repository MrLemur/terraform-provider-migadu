package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAliasResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	resourceName := "migadu_alias.test"
	destMailboxLocalPart := fmt.Sprintf("tfacc-dest-%d", time.Now().UnixNano())
	aliasLocalPart := fmt.Sprintf("tfacc-alias-%d", time.Now().UnixNano())
	importID := fmt.Sprintf("%s/%s", os.Getenv("MIGADU_TEST_DOMAIN"), aliasLocalPart)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig(destMailboxLocalPart, aliasLocalPart),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", os.Getenv("MIGADU_TEST_DOMAIN")),
					resource.TestCheckResourceAttr(resourceName, "local_part", aliasLocalPart),
					resource.TestCheckResourceAttr(resourceName, "destinations.0", fmt.Sprintf("%s@%s", destMailboxLocalPart, os.Getenv("MIGADU_TEST_DOMAIN"))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     importID,
			},
		},
	})
}
