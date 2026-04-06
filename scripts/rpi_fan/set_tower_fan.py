#!/usr/bin/env python3
import sys

if len(sys.argv) != 2:
    print("Usage: python set_tower_fan.py <pwm_value_0-4>")
    sys.exit(1)

pwm = int(sys.argv[1])
if not (0 <= pwm <= 4):
    print("PWM must be 0-4")
    sys.exit(1)

try:
    with open('/sys/class/thermal/cooling_device0/cur_state', 'w') as f:
        f.write(str(pwm))
    print(f"OK: Tower fan PWM set to {pwm}")
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)