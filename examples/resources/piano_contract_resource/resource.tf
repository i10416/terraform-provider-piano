resource "piano_contract" "example" {
  aid                      = "example-aid"
  licensee_id              = "example-licensee-id"
  rid                      = "example-rid"
  contract_type            = "EMAIL_DOMAIN_CONTRACT"
  name                     = "Example Email Domain Contract"
  seats_number             = 100
  is_hard_seats_limit_type = true
  contract_periods         = []
}
