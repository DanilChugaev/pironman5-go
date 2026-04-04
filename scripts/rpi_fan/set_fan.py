#!/usr/bin/env python3
import sys
import lgpio

if len(sys.argv) != 3:
    print("Usage: python set_fan.py <gpio_pin> <0|1>")
    sys.exit(1)

pin = int(sys.argv[1])
state = int(sys.argv[2])

try:
    h = lgpio.gpiochip_open(4)          # ← на RPi 5 основной чип = 4
    lgpio.gpio_claim_output(h, pin)
    lgpio.gpio_write(h, pin, state)
    status = "ON" if state else "OFF"
    print(f"OK: Fan GPIO{pin} → {status} (lgpio, chip=4)")
    # Не закрываем handle здесь — состояние сохранится
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)