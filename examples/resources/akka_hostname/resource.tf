resource "akka_hostname" "api_domain" {
  hostname   = "api.example.com"
  project    = akka_project.example.name
  tls_secret = akka_secret.tls.name
}

output "hostname_status" {
  value = akka_hostname.api_domain.status
}
