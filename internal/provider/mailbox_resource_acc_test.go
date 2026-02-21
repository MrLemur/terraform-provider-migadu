package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailboxResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	resourceName := "migadu_mailbox.test"
	localPart := fmt.Sprintf("tfacc-mailbox-%d", time.Now().UnixNano())
	importID := fmt.Sprintf("%s/%s", os.Getenv("MIGADU_TEST_DOMAIN"), localPart)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMailboxConfig(localPart),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", os.Getenv("MIGADU_TEST_DOMAIN")),
					resource.TestCheckResourceAttr(resourceName, "local_part", localPart),
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
