#!/bin/bash

# ============================================
# КОНФИГУРАЦИЯ ЗАВИСИМОСТЕЙ
# ============================================

# Пакеты для установки через apt
APT_PACKAGES=(
    "python3"
    "python3-pip"
    "python3-venv"
    "git"
    "curl"
    "wget"
    "build-essential"
    "libssl-dev"
    "libffi-dev"
    "python3-dev"
    "kmod"
    "i2c-tools"
    "python3-gpiozero"
    "lsof"
    "pyudev"
)

# Пакеты для установки через pip
PIP_PACKAGES=(
    "requests"
    "psutil"
    "adafruit-circuitpython-neopixel-spi"
    "Adafruit-Blinka==8.59.0"
    "Pillow"
    "smbus2"
    "gpiozero"
    "gpiod"
    "rpi.lgpio"
)

# Дополнительные bash-скрипты для выполнения
CUSTOM_SCRIPTS=(
    "scripts/install_lgpio.sh"
)

# Путь к requirements.txt (если есть)
# REQUIREMENTS_FILE="requirements.txt"

# Имя виртуального окружения
VENV_NAME="venv"

# Лог файл
LOG_FILE="setup.log"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color