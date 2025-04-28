resource "piano_licensee" "example" {
  aid         = "example"
  name        = "example"
  description = "example piano licensee"
  managers = [
    {
      first_name    = "John",
      last_name     = "Doe",
      personal_name = "John Doe",
      uid           = "sample"
    }
  ]
  representatives = [
    {
      "email" : "john.doe@example.com"
    }
  ]
}
