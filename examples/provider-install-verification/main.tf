terraform {
  required_providers {
    akka = {
      source  = "thinkmorestupidless/akka"
      version = "~> 0.1"
    }
  }
}

provider "akka" {
  organization = var.akka_org
}

variable "akka_org" {
  description = "Akka organization name"
  type        = string
}

# List available regions — verifies provider is configured and can reach Akka
data "akka_regions" "check" {}

output "available_regions" {
  value = data.akka_regions.check.regions[*].name
}
