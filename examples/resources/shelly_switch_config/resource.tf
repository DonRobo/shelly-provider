resource "shelly_switch_config" "example" {
  ip            = "192.168.1.100"
  id            = 0
  name          = "Living Room Light"
  in_mode       = "momentary"
  initial_state = "restore_last"
}
