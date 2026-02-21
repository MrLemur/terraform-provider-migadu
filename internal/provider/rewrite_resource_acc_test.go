package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRewriteResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	resourceName := "migadu_rewrite.test"
	ruleName := fmt.Sprintf("tfacc-rule-%d", time.Now().UnixNano())
	importID := fmt.Sprintf("%s/%s", os.Getenv("MIGADU_TEST_DOMAIN"), ruleName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRewriteConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", os.Getenv("MIGADU_TEST_DOMAIN")),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "local_part_rule", "tfacc-catch+"),
					resource.TestCheckResourceAttr(resourceName, "order_num", "50"),
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

func testAccRewriteConfig(ruleName string) string {
	return fmt.Sprintf(`
resource "migadu_rewrite" "test" {
  domain_name     = "%s"
  name            = "%s"
  local_part_rule = "tfacc-catch+"
  order_num       = 50
  destinations    = ["admin@%s"]
}
`, os.Getenv("MIGADU_TEST_DOMAIN"), ruleName, os.Getenv("MIGADU_TEST_DOMAIN"))
}
