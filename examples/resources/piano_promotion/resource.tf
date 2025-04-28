

resource "piano_promotion" "sample" {
  aid = "sample-aid"
  name = "sample"
  # null indicates unlimited uses
  uses_allowed = null
}
