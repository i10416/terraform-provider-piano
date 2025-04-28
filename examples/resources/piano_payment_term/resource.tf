resource "piano_payment_term" "sample" {
  aid  = "sample-aid"
  name = "Sample Payment Term"
  schedule = {
    schedule_id = "sample-schedule-id"
  }
  resource = {
    rid = "sample-rid"
  }
  payment_billing_plan_description = "Sample Payment Term"
  payment_force_auto_renew         = false
  payment_trial_new_customers_only = false
  payment_allow_promo_codes        = false
  payment_allow_renew_days         = 0
  payment_new_customers_only       = false
  change_options                   = []
}
