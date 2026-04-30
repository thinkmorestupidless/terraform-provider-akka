resource "akka_route" "api" {
  name     = "api-route"
  project  = akka_project.example.name
  hostname = "api.example.com"

  paths = {
    "/api/v1" = akka_service.v1.name
    "/api/v2" = akka_service.v2.name
  }
}
