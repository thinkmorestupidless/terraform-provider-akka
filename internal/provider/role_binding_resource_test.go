package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleBindingResource_project(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-rb-proj")
	user := os.Getenv("AKKA_TEST_USER")
	if user == "" {
		t.Skip("AKKA_TEST_USER not set; skipping role binding acceptance test")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleBindingProjectConfig(projName, user, "developer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_role_binding.test", "user", user),
					resource.TestCheckResourceAttr("akka_role_binding.test", "role", "developer"),
					resource.TestCheckResourceAttr("akka_role_binding.test", "scope", "project"),
				),
			},
		},
	})
}

func TestAccRoleBindingResource_org(t *testing.T) {
	user := os.Getenv("AKKA_TEST_USER")
	if user == "" {
		t.Skip("AKKA_TEST_USER not set; skipping org role binding acceptance test")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleBindingOrgConfig(user, "developer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_role_binding.test", "user", user),
					resource.TestCheckResourceAttr("akka_role_binding.test", "scope", "organization"),
				),
			},
		},
	})
}

func testAccRoleBindingProjectConfig(projName, user, role string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_role_binding" "test" {
  user    = %q
  role    = %q
  project = akka_project.test.name
  scope   = "project"
}
`, projName, user, role)
}

func testAccRoleBindingOrgConfig(user, role string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_role_binding" "test" {
  user  = %q
  role  = %q
  scope = "organization"
}
`, user, role)
}
