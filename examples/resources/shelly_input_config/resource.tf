resource "shelly_input_config" "example" {
  ip   = "192.168.1.100"
  id   = 0
  name = "Living Room Button"
  type = "button"
}
