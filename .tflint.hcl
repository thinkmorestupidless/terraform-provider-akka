config {
  # Fail on rule violations (don't just warn)
  force = false
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}
