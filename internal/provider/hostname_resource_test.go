package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostnameResource_basic(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-hn-proj")
	hostname := fmt.Sprintf("%s.example.com", acctest.RandomWithPrefix("tf-acc"))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostnameResourceConfig(projName, hostname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_hostname.test", "hostname", hostname),
					resource.TestCheckResourceAttrSet("akka_hostname.test", "status"),
				),
			},
		},
	})
}

func testAccHostnameResourceConfig(projName, hostname string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_hostname" "test" {
  hostname = %q
  project  = akka_project.test.name
}
`, projName, hostname)
}
