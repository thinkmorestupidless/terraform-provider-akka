package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceResource_deploy(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-svc-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_service.test", "name", svcName),
					resource.TestCheckResourceAttr("akka_service.test", "image", "docker.io/library/nginx:latest"),
					resource.TestCheckResourceAttr("akka_service.test", "paused", "false"),
					resource.TestCheckResourceAttr("akka_service.test", "exposed", "false"),
				),
			},
		},
	})
}

func TestAccServiceResource_pause(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-svc-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", false, false),
				Check:  resource.TestCheckResourceAttr("akka_service.test", "paused", "false"),
			},
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", true, false),
				Check:  resource.TestCheckResourceAttr("akka_service.test", "paused", "true"),
			},
		},
	})
}

func TestAccServiceResource_expose(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-svc-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", false, false),
				Check:  resource.TestCheckResourceAttr("akka_service.test", "exposed", "false"),
			},
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_service.test", "exposed", "true"),
					resource.TestCheckResourceAttrSet("akka_service.test", "hostname"),
				),
			},
		},
	})
}

func TestAccServiceResource_import(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-svc-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceConfig(projName, svcName, "docker.io/library/nginx:latest", false, false),
			},
			{
				ResourceName:      "akka_service.test",
				ImportState:       true,
				ImportStateId:     projName + "/" + svcName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDataSource(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-svc-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig(projName, svcName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.akka_service.test", "name", svcName),
					resource.TestCheckResourceAttrSet("data.akka_service.test", "image"),
				),
			},
		},
	})
}

func testAccServiceResourceConfig(projName, svcName, image string, paused, exposed bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_service" "test" {
  name    = %q
  project = akka_project.test.name
  image   = %q
  paused  = %v
  exposed = %v
}
`, projName, svcName, image, paused, exposed)
}

func testAccServiceDataSourceConfig(projName, svcName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "test" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_service" "test" {
  name    = %q
  project = akka_project.test.name
  image   = "docker.io/library/nginx:latest"
}

data "akka_service" "test" {
  name       = akka_service.test.name
  project    = akka_project.test.name
  depends_on = [akka_service.test]
}
`, projName, svcName)
}
