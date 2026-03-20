#!/bin/bash

# ============================================
# СКРИПТ УСТАНОВКИ ЗАВИСИМОСТЕЙ
# ============================================

set -e  # Остановка при ошибке

# Загрузка конфигурации
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# ============================================
# ФУНКЦИИ
# ============================================

log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")  echo -e "${BLUE}[$timestamp] [INFO]${NC} $message" ;;
        "SUCCESS") echo -e "${GREEN}[$timestamp] [SUCCESS]${NC} $message" ;;
        "WARNING") echo -e "${YELLOW}[$timestamp] [WARNING]${NC} $message" ;;
        "ERROR") echo -e "${RED}[$timestamp] [ERROR]${NC} $message" ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log "WARNING" "Некоторые операции требуют прав суперпользователя"
        return 1
    fi
    return 0
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

install_apt_packages() {
    log "INFO" "Установка пакетов через apt..."
    
    if ! check_root; then
        log "INFO" "Запрос прав суперпользователя для apt..."
        sudo apt-get update -qq
        sudo apt-get install -y "${APT_PACKAGES[@]}"
    else
        apt-get update -qq
        apt-get install -y "${APT_PACKAGES[@]}"
    fi
    
    log "SUCCESS" "APT пакеты установлены"
}

create_venv() {
    log "INFO" "Создание виртуального окружения..."
    
    if [ -d "$VENV_NAME" ]; then
        log "WARNING" "Виртуальное окружение уже существует, удаляем..."
        rm -rf "$VENV_NAME"
    fi
    
    python3 -m venv "$VENV_NAME"
    log "SUCCESS" "Виртуальное окружение создано: $VENV_NAME"
}

activate_venv() {
    source "$VENV_NAME/bin/activate"
    log "INFO" "Виртуальное окружение активировано"
}

install_pip_packages() {
    log "INFO" "Установка пакетов через pip..."
    
    # Обновление pip
    pip install --upgrade pip -q
    
    # Установка пакетов из конфига
    if [ ${#PIP_PACKAGES[@]} -ne 0 ]; then
        pip install "${PIP_PACKAGES[@]}" -q
        log "SUCCESS" "PIP пакеты из конфига установлены"
    fi
    
    # Установка из requirements.txt
    if [ -f "$REQUIREMENTS_FILE" ]; then
        log "INFO" "Установка зависимостей из $REQUIREMENTS_FILE..."
        pip install -r "$REQUIREMENTS_FILE" -q
        log "SUCCESS" "Зависимости из requirements.txt установлены"
    else
        log "WARNING" "Файл $REQUIREMENTS_FILE не найден"
    fi
}

run_custom_scripts() {
    log "INFO" "Запуск дополнительных скриптов..."
    
    for script in "${CUSTOM_SCRIPTS[@]}"; do
        if [ -f "$script" ]; then
            log "INFO" "Выполнение: $script"
            chmod +x "$script"
            bash "$script"
            log "SUCCESS" "Скрипт выполнен: $script"
        else
            log "WARNING" "Скрипт не найден: $script"
        fi
    done
}

show_summary() {
    echo ""
    echo "============================================"
    log "SUCCESS" "Установка завершена успешно!"
    echo "============================================"
    echo ""
    echo "Для активации виртуального окружения выполните:"
    echo -e "${GREEN}source $VENV_NAME/bin/activate${NC}"
    echo ""
    echo "Лог файл: $LOG_FILE"
    echo "============================================"
}

# ============================================
# ОСНОВНАЯ ЛОГИКА
# ============================================

main() {
    echo ""
    echo "============================================"
    echo "  Установка зависимостей проекта"
    echo "============================================"
    echo ""
    
    # Инициализация лог файла
    echo "=== Setup started at $(date) ===" > "$LOG_FILE"
    
    # Проверка Python
    if ! check_command "python3"; then
        log "ERROR" "Python3 не найден. Установите его сначала."
        exit 1
    fi
    
    log "INFO" "Python версия: $(python3 --version)"
    
    # Установка apt пакетов
    install_apt_packages
    
    # Создание и активация venv
    create_venv
    activate_venv
    
    # Установка pip пакетов
    install_pip_packages
    
    # Запуск дополнительных скриптов
    run_custom_scripts
    
    # Показ итогов
    show_summary
}

# Обработка аргументов
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --apt-only     Установить только apt пакеты"
        echo "  --pip-only     Установить только pip пакеты"
        echo "  --scripts-only Запустить только дополнительные скрипты"
        echo "  --clean        Очистить виртуальное окружение"
        exit 0
        ;;
    --apt-only)
        install_apt_packages
        ;;
    --pip-only)
        activate_venv
        install_pip_packages
        ;;
    --scripts-only)
        activate_venv
        run_custom_scripts
        ;;
    --clean)
        log "INFO" "Очистка виртуального окружения..."
        rm -rf "$VENV_NAME"
        log "SUCCESS" "Очистка завершена"
        ;;
    *)
        main
        ;;
esac