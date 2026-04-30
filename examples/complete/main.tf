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

variable "akka_org" {}
variable "jwt_signing_key" { sensitive = true }
variable "tls_private_key" { sensitive = true }

resource "akka_project" "prod" {
  name   = "production"
  region = "gcp-us-east1"
}

resource "akka_secret" "jwt" {
  name    = "jwt-signing-key"
  project = akka_project.prod.name
  type    = "symmetric"
  value   = var.jwt_signing_key
}

resource "akka_secret" "tls" {
  name        = "api-tls"
  project     = akka_project.prod.name
  type        = "tls"
  certificate = file("certs/api.pem")
  private_key = var.tls_private_key
}

resource "akka_service" "api" {
  name    = "api-service"
  project = akka_project.prod.name
  image   = "docker.io/myorg/api:2.1.0"
  exposed = true
  env = {
    JWT_SECRET_NAME = akka_secret.jwt.name
  }
  depends_on = [akka_secret.jwt]
}

resource "akka_hostname" "api_domain" {
  hostname   = "api.example.com"
  project    = akka_project.prod.name
  tls_secret = akka_secret.tls.name
  depends_on = [akka_secret.tls]
}

resource "akka_route" "api_route" {
  name       = "api"
  project    = akka_project.prod.name
  hostname   = akka_hostname.api_domain.hostname
  paths      = { "/" = akka_service.api.name }
  depends_on = [akka_hostname.api_domain, akka_service.api]
}

resource "akka_role_binding" "dev_alice" {
  user    = "alice@example.com"
  role    = "developer"
  project = akka_project.prod.name
}
