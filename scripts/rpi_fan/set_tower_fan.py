#!/usr/bin/env python3
import sys
import os
import time
import subprocess

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

# Официальные значения скорости (как в Raspberry Pi OS)
duty_map = [0, 75, 125, 175, 250]   # 0-255
duty = duty_map[level]

try:
    # === 1. ОСВОБОЖДАЕМ КАНАЛ ===
    subprocess.run(["modprobe", "-r", "pwm_fan"],
                   check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    # Если канал уже экспортирован — сначала выгружаем
    if os.path.exists(PWM_PATH):
        try:
            with open(f"{PWM_CHIP}/unexport", "w") as f:
                f.write(str(PWM_NUM))
            time.sleep(0.1)
        except:
            pass

    # === 2. Экспортируем канал ===
    if not os.path.exists(PWM_PATH):
        with open(f"{PWM_CHIP}/export", "w") as f:
            f.write(str(PWM_NUM))
        time.sleep(0.2)

    # === 3. Настраиваем PWM (1 кГц) ===
    with open(f"{PWM_PATH}/period", "w") as f:
        f.write("1000000")                    # 1 кГц
    with open(f"{PWM_PATH}/duty_cycle", "w") as f:
        f.write(str(duty * 1000000 // 255))   # переводим в наносекунды
    with open(f"{PWM_PATH}/enable", "w") as f:
        f.write("1")

    print(f"OK: Tower fan level {level} (duty {duty}/255)")

except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)