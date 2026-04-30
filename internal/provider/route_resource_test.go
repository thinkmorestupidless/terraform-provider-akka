package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRouteResource_basic(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-rt-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-rt-svc")
	routeName := acctest.RandomWithPrefix("tf-acc-rt")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfig(projName, svcName, routeName, map[string]string{"/": svcName}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_route.test", "name", routeName),
					resource.TestCheckResourceAttr("akka_route.test", "paths.%", "1"),
				),
			},
		},
	})
}

func TestAccRouteResource_update(t *testing.T) {
	projName := acctest.RandomWithPrefix("tf-acc-rt-proj")
	svcName := acctest.RandomWithPrefix("tf-acc-rt-svc")
	routeName := acctest.RandomWithPrefix("tf-acc-rt")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfig(projName, svcName, routeName, map[string]string{"/": svcName}),
				Check:  resource.TestCheckResourceAttr("akka_route.test", "paths.%", "1"),
			},
			{
				Config: testAccRouteResourceConfig(projName, svcName, routeName, map[string]string{"/api": svcName, "/": svcName}),
				Check:  resource.TestCheckResourceAttr("akka_route.test", "paths.%", "2"),
			},
		},
	})
}

func testAccRouteResourceConfig(projName, svcName, routeName string, paths map[string]string) string {
	pathsHCL := ""
	for k, v := range paths {
		pathsHCL += fmt.Sprintf("    %q = %q\n", k, v)
	}
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

resource "akka_route" "test" {
  name     = %q
  project  = akka_project.test.name
  hostname = "test.example.com"
  paths = {
%s  }
  depends_on = [akka_service.test]
}
`, projName, svcName, routeName, pathsHCL)
}
