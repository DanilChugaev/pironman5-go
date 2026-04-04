#!/usr/bin/env python3
import sys
from gpiozero import DigitalOutputDevice

if len(sys.argv) != 3:
    print("Usage: python set_fan.py <gpio_pin> <on_or_off>  # on=1, off=0")
    sys.exit(1)

pin = int(sys.argv[1])
state = int(sys.argv[2])  # 1 = ON, 0 = OFF

try:
    # Пробуем ОБРАТНУЮ полярность (самое вероятное решение)
    fan = DigitalOutputDevice(pin, active_high=False, initial_value=False)
    fan.value = bool(state)
    status = "ON" if state else "OFF"
    print(f"OK: Fan GPIO{pin} → {status} (active_high=False)")
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)