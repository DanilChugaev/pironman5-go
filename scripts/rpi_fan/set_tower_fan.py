#!/usr/bin/env python3
import sys
import time
import os

PWM_CHIP = "/sys/class/pwm/pwmchip0"
PWM_NUM = 3
PWM_PATH = f"{PWM_CHIP}/pwm{PWM_NUM}"

if len(sys.argv) != 2:
    print("Usage: python set_tower_fan.py <0-4>")
    sys.exit(1)

level = int(sys.argv[1])
if not (0 <= level <= 4):
    print("ERROR: level must be 0-4")
    sys.exit(1)

# Маппинг уровня 0-4 → duty_cycle (0-255)
duty_map = [0, 75, 125, 175, 250]   # официальные значения SunFounder/RPi
duty = duty_map[level]

try:
    # 1. Если канал ещё не экспортирован — экспортируем
    if not os.path.exists(PWM_PATH):
        with open(f"{PWM_CHIP}/export", "w") as f:
            f.write(str(PWM_NUM))
        time.sleep(0.2)  # даём ядру время

    # 2. Настраиваем PWM (один раз)
    with open(f"{PWM_PATH}/period", "w") as f:
        f.write("1000000")          # 1 кГц — оптимально для вентилятора
    with open(f"{PWM_PATH}/duty_cycle", "w") as f:
        f.write(str(duty * 1000000 // 255))   # переводим 0-255 в наносекунды
    with open(f"{PWM_PATH}/enable", "w") as f:
        f.write("1")

    print(f"OK: Tower fan level {level} (duty ~{duty}/255)")

except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)