resource "shelly_input_config" "light_input" {
  ip = var.device_ip
  id = var.input_id

  name = "${var.light_name} ${var.switch_type == "switch" ? "Switch" : "Button"}"
}

resource "shelly_switch_config" "light_switch" {
  ip = var.device_ip
  id = var.input_id

  name          = "${var.light_name} Light"
  in_mode       = var.switch_type == "switch" ? "follow" : "momentary"
  initial_state = var.switch_type == "switch" ? "match_input" : "restore_last"

  # This ensures that the input is configured before the switch,
  # which is good practice for dependent configurations.
  depends_on = [
    shelly_input_config.light_input
  ]
}
