package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-ds-proj")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.akka_project.test", "name", name),
					resource.TestCheckResourceAttrSet("data.akka_project.test", "id"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

data "akka_project" "test" {
  name       = akka_project.test.name
  depends_on = [akka_project.test]
}
`, name)
}
