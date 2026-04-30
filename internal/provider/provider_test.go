package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/thinkmorestupidless/terraform-provider-akka/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"akka": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("AKKA_TOKEN"); v == "" {
		t.Fatal("AKKA_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("AKKA_TEST_ORG"); v == "" {
		t.Fatal("AKKA_TEST_ORG must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	return `
provider "akka" {
  organization = "` + os.Getenv("AKKA_TEST_ORG") + `"
}
`
}

// Ensure resource package is not unused — suppress lint warning.
var _ = resource.TestCase{}
