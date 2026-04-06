#!/usr/bin/env python3
import sys
import subprocess

if len(sys.argv) != 2:
    print("Usage: python set_tower_fan.py <0-4>")
    sys.exit(1)

pwm = int(sys.argv[1])
if not (0 <= pwm <= 4):
    print("ERROR: PWM must be 0-4")
    sys.exit(1)

try:
    # 1. Принудительно переводим пин FAN_PWM в manual-режим (отключаем firmware override)
    subprocess.run(["pinctrl", "FAN_PWM", "op", "dl"], check=False, stdout=subprocess.DEVNULL)

    # 2. Записываем скорость
    with open('/sys/class/thermal/cooling_device0/cur_state', 'w') as f:
        f.write(str(pwm))

    print(f"OK: Tower fan PWM = {pwm}")
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)