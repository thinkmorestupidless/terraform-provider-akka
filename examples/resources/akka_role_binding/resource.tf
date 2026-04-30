# Project-scoped developer binding
resource "akka_role_binding" "dev_alice" {
  user    = "alice@example.com"
  role    = "developer"
  project = akka_project.example.name
  scope   = "project"
}

# Organization-scoped admin binding
resource "akka_role_binding" "org_admin" {
  user  = "bob@example.com"
  role  = "admin"
  scope = "organization"
}
