resource "shelly_input_config" "input_config_0" {
  ip       = "192.168.1.169"
  id       = 0
  type     = "switch"
}
  
#TODO Support proper error for this:
# resource "shelly_input_config" "input_config_2" {
#   ip       = "192.168.1.169"
#   id       = 2
# }
