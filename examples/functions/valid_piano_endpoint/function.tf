# Validate if the string is valid piano endpoint url prefix
output "validation_result" {
  value = provider::piano::valid_piano_endpoint("https://sandbox.piano.io/api/v3")
}
