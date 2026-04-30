package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.akka_regions.test", "regions.#"),
					resource.TestCheckResourceAttrSet("data.akka_regions.test", "regions.0.name"),
				),
			},
		},
	})
}

func testAccRegionsDataSourceConfig() string {
	return testAccProviderConfig() + `
data "akka_regions" "test" {}
`
}
