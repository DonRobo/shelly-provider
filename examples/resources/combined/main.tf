module "living_room_light" {
  source      = "../../modules/light"
  device_ip   = "192.168.1.100"
  input_id    = 0
  light_name  = "Living Room"
  switch_type = "button"
}
