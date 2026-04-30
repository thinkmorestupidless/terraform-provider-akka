package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecretResource_symmetric(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-sec-proj")
	name := acctest.RandomWithPrefix("tf-acc-sec")
	value := acctest.RandString(32)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretSymmetricConfig(projName, name, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_secret.test", "name", name),
					resource.TestCheckResourceAttr("akka_secret.test", "type", "symmetric"),
				),
			},
		},
	})
}

func TestAccSecretResource_generic(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-sec-proj")
	name := acctest.RandomWithPrefix("tf-acc-sec")
	value := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretGenericConfig(projName, name, value),
				Check:  resource.TestCheckResourceAttr("akka_secret.test", "type", "generic"),
			},
		},
	})
}

func TestAccSecret_sensitiveFields(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-sec-proj")
	name := acctest.RandomWithPrefix("tf-acc-sec")
	value := acctest.RandString(32)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretSymmetricConfig(projName, name, value),
				Check: resource.ComposeTestCheckFunc(
					// value is sensitive and must not appear as a plain attribute
					resource.TestCheckResourceAttrSet("akka_secret.test", "value"),
				),
			},
		},
	})
}

func testAccSecretSymmetricConfig(projName, name, value string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_secret" "test" {
  name    = %q
  project = akka_project.test.name
  type    = "symmetric"
  value   = %q
}
`, projName, name, value)
}

func testAccSecretGenericConfig(projName, name, value string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_secret" "test" {
  name    = %q
  project = akka_project.test.name
  type    = "generic"
  value   = %q
}
`, projName, name, value)
}
