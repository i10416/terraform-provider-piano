resource "piano_external_term" "sample" {
  aid = "sample-aid"
  resource = {
    rid = "sample-rid"
  }

  description             = "Sample External Term"
  evt_itunes_product_id   = "evt-itunes-product-id"
  external_api_id         = "external-api-id"
  evt_grace_period        = 10
  evt_itunes_bundle_id    = ""
  evt_verification_period = 86400
  name                    = "Sample External Term"
}
