package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccComplete exercises the full resource graph in a single config:
// project → symmetric secret → service → hostname → route → role binding.
// It also verifies idempotency by applying the same config twice (T073).
func TestAccComplete(t *testing.T) {
	suffix := acctest.RandString(6)
	projName := "tf-acc-complete-" + suffix
	svcName := "svc-" + suffix
	secretName := "sec-" + suffix
	routeName := "rt-" + suffix
	hostname := fmt.Sprintf("%s.example.com", suffix)
	secretValue := acctest.RandString(32)

	user := os.Getenv("AKKA_TEST_USER")
	if user == "" {
		user = "test-user@example.com"
	}

	cfg := testAccCompleteConfig(projName, svcName, secretName, routeName, hostname, secretValue, user)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create entire stack
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("akka_project.complete", "name", projName),
					resource.TestCheckResourceAttr("akka_service.complete", "name", svcName),
					resource.TestCheckResourceAttr("akka_secret.complete", "name", secretName),
					resource.TestCheckResourceAttr("akka_route.complete", "name", routeName),
					resource.TestCheckResourceAttr("akka_hostname.complete", "hostname", hostname),
					resource.TestCheckResourceAttrSet("akka_hostname.complete", "status"),
				),
			},
			// Step 2: verify idempotency — second plan must show no changes (T073)
			{
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCompleteConfig(projName, svcName, secretName, routeName, hostname, secretValue, user string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "akka_project" "complete" {
  name   = %q
  region = "gcp-us-east1"
}

resource "akka_secret" "complete" {
  name    = %q
  project = akka_project.complete.name
  type    = "symmetric"
  value   = %q
}

resource "akka_service" "complete" {
  name    = %q
  project = akka_project.complete.name
  image   = "docker.io/library/nginx:latest"
  env = {
    SECRET_NAME = akka_secret.complete.name
  }
  depends_on = [akka_secret.complete]
}

resource "akka_hostname" "complete" {
  hostname = %q
  project  = akka_project.complete.name
}

resource "akka_route" "complete" {
  name     = %q
  project  = akka_project.complete.name
  hostname = akka_hostname.complete.hostname
  paths = {
    "/" = akka_service.complete.name
  }
  depends_on = [akka_hostname.complete, akka_service.complete]
}

resource "akka_role_binding" "complete" {
  user    = %q
  role    = "developer"
  project = akka_project.complete.name
  scope   = "project"
}
`, projName, secretName, secretValue, svcName, hostname, routeName, user)
}
