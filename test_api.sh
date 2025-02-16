#!/bin/bash

# Задаем базовый URL вашего API
BASE_URL="http://localhost:8080"

# Авторизация - Получаем JWT токен
echo "1. Получение JWT токена..."
AUTH_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth -H "Content-Type: application/json" -d '{"username": "test_user1", "password": "password"}')
TOKEN=$(echo $AUTH_RESPONSE | jq -r .token)  # Извлекаем токен из JSON ответа
echo "Получен токен: $TOKEN"
echo ""

# 1. Получить информацию о пользователе
echo "2. Получение информации о пользователе..."
curl -s -X GET $BASE_URL/api/info -H "Authorization: Bearer $TOKEN"  -H "Content-Type: application/json" | jq
echo ""
# 2. Перевести монеты пользователю
echo "4. Перевод монет..."
TRANSFER_RESPONSE=$(curl -s -X POST $BASE_URL/api/sendCoin -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"toUser": "user1", "amount": 50}')
echo $TRANSFER_RESPONSE
echo ""


echo "4. Перевод монет...self"
TRANSFER_RESPONSE=$(curl -s -X POST $BASE_URL/api/sendCoin -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"toUser": "test_user1", "amount": 50}')
echo $TRANSFER_RESPONSE
echo ""

# 3. Покупка товара
echo "5. Покупка товара..."
BUY_RESPONSE=$(curl -s -X GET $BASE_URL/api/buy/pen -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
echo $BUY_RESPONSE
echo ""

echo "Тестирование завершено."

