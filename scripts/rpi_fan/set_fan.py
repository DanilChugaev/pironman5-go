#!/usr/bin/env python3
import sys
import lgpio

if len(sys.argv) != 5:
    print("Usage: python set_fan.py <fan_pin> <fan_state> <led_pin> <led_state>")
    print("       states: 0=OFF, 1=ON")
    sys.exit(1)

fan_pin = int(sys.argv[1])
fan_state = int(sys.argv[2])
led_pin = int(sys.argv[3])
led_state = int(sys.argv[4])

try:
    h = lgpio.gpiochip_open(4)  # RPi 5 main gpiochip

    # Fan
    lgpio.gpio_claim_output(h, fan_pin)
    lgpio.gpio_write(h, fan_pin, fan_state)

    # LED
    lgpio.gpio_claim_output(h, led_pin)
    lgpio.gpio_write(h, led_pin, led_state)

    fan_status = "ON" if fan_state else "OFF"
    led_status = "ON" if led_state else "OFF"
    print(f"OK: Fan GPIO{fan_pin}={fan_status} | LED GPIO{led_pin}={led_status}")

except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)