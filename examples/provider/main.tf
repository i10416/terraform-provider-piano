data "piano_app" "sample" {
  aid = var.PIANO_APP_ID
}

resource "piano_resource" "sample" {
  aid  = var.PIANO_APP_ID
  name = "Sample"
  is_fbia_resource = false
}
