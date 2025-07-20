module "kitchen_light" {
    source     = "../../modules/light"
    device_ip = "192.168.1.80"
    input_id = 0
    light_name = "Küche"
    switch_type = "button"
}
