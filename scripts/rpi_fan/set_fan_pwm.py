#!/usr/bin/env python3
import sys
from gpiozero import PWMOutputDevice
from gpiozero import Device

# Чтобы gpiozero использовал lgpio (уже установлен у тебя)
Device.pin_factory = None  # по умолчанию использует lgpio на RPi5

if len(sys.argv) != 3:
    print("Usage: python set_fan_pwm.py <gpio_pin> <duty_cycle_0-100>")
    sys.exit(1)

pin = int(sys.argv[1])
duty = float(sys.argv[2]) / 100.0  # 0.0 - 1.0

if not (0 <= duty <= 1.0):
    print("Duty cycle must be 0-100")
    sys.exit(1)

try:
    fan = PWMOutputDevice(pin, frequency=100, initial_value=duty)  # 100 Гц — оптимально для вентиляторов
    fan.value = duty
    print(f"OK: Fan on GPIO{pin} set to {duty*100:.0f}% PWM")
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)