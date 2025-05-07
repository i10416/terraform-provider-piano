
resource "piano_offer" "sample" {
  name = "Sample"
  aid  = "sample-aid"
}

resource "piano_offer_term_binding" "term_a_in_sample_offer" {
  aid      = "sample-aid"
  offer_id = piano_offer.sample.offer_id
  term_id  = "term-a-id"
}

resource "piano_offer_term_binding" "term_b_in_sample_offer" {
  aid      = "sample-aid"
  offer_id = piano_offer.sample.offer_id
  term_id  = "term-b-id"
}

resource "piano_offer_term_order" "term_order_in_sample_offer" {
  aid      = "sample-aid"
  offer_id = piano_offer.sample.offer_id
  term_ids = [
    piano_offer_term_binding.term_b_in_sample_offer.term_id,
    piano_offer_term_binding.term_a_in_sample_offer.term_id,
  ]
}
