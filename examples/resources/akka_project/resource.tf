resource "akka_project" "example" {
  name        = "my-project"
  description = "My Akka project"
  region      = "gcp-us-east1"
}

output "project_id" {
  value = akka_project.example.id
}
