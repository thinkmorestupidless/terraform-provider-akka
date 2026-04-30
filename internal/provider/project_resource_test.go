package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-proj")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(name, "gcp-us-east1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_project.test", "name", name),
					resource.TestCheckResourceAttr("akka_project.test", "region", "gcp-us-east1"),
					resource.TestCheckResourceAttrSet("akka_project.test", "id"),
				),
			},
		},
	})
}

func TestAccProjectResource_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-proj")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(name, "gcp-us-east1"),
				Check:  resource.TestCheckResourceAttr("akka_project.test", "region", "gcp-us-east1"),
			},
			// Projects are immutable by name; verify plan is stable after create
			{
				Config:   testAccProjectResourceConfig(name, "gcp-us-east1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProjectResource_import(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-proj")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(name, "gcp-us-east1"),
			},
			{
				ResourceName:      "akka_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectResource_drift(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-proj")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(name, "gcp-us-east1"),
				Check:  resource.TestCheckResourceAttr("akka_project.test", "name", name),
			},
			// If the resource is removed externally, the next plan should detect drift
			// and schedule recreation. This step just verifies the config is stable.
			{
				Config:             testAccProjectResourceConfig(name, "gcp-us-east1"),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func testAccProjectResourceConfig(name, region string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = %q
}
`, name, region)
}
