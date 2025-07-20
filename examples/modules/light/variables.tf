variable "device_ip" {
  description = "The IP address or hostname of the Shelly device."
  type        = string
}

variable "light_name" {
  description = "The name for the light."
  type        = string
}

variable "input_id" {
  description = "The ID of the physical input/output components to configure (e.g., 0, 1, 2, 3)."
  type        = number
}

variable "switch_type" {
  description = "The type of the physical light switch. Must be 'switch' or 'button'."
  type        = string

  validation {
    condition     = contains(["switch", "button"], var.switch_type)
    error_message = "The switch_type must be either 'switch' or 'button'."
  }
}
