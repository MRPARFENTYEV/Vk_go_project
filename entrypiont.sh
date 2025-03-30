#!/bin/sh

# Запускаем Tarantool в фоне
echo "Starting Tarantool..."
tarantool /etc/tarantool/tarantool.cfg &

# Ждем инициализации
sleep 3

# Проверяем подключение
if nc -z localhost 3301; then
    echo "Tarantool is ready"
else
    echo "Tarantool failed to start"
    exit 1
fi

# Запускаем приложение
echo "Starting voting app..."
exec /app/vk_go