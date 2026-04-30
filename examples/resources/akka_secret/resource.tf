# Symmetric secret (JWT signing key)
resource "akka_secret" "jwt" {
  name    = "jwt-signing-key"
  project = akka_project.example.name
  type    = "symmetric"
  value   = var.jwt_signing_key
}

# TLS secret
resource "akka_secret" "tls" {
  name        = "api-tls"
  project     = akka_project.example.name
  type        = "tls"
  certificate = file("certs/api.pem")
  private_key = var.tls_private_key
}

# Generic secret (database password)
resource "akka_secret" "db" {
  name    = "db-password"
  project = akka_project.example.name
  type    = "generic"
  value   = var.db_password
}

variable "jwt_signing_key" { sensitive = true }
variable "tls_private_key" { sensitive = true }
variable "db_password" { sensitive = true }
