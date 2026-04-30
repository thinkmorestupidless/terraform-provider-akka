resource "akka_service" "example" {
  name    = "my-service"
  project = akka_project.example.name
  image   = "docker.io/myorg/app:1.0"

  replicas = 2
  exposed  = true

  env = {
    APP_ENV   = "production"
    LOG_LEVEL = "info"
  }

  timeouts {
    create = "15m"
    update = "15m"
  }
}

output "service_hostname" {
  value = akka_service.example.hostname
}
