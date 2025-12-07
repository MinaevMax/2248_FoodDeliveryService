#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${YELLOW}=== Testing Food Delivery Service API ===${NC}\n"

# Test 1: Register user
echo -e "${YELLOW}1. Testing POST /auth/register${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"login":"testuser","password":"password123"}')

echo "Response: $REGISTER_RESPONSE"

# Extract token (if available)
if echo "$REGISTER_RESPONSE" | grep -q "id"; then
  echo -e "${GREEN}✓ Registration successful${NC}\n"
  USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
else
  echo -e "${RED}✗ Registration failed${NC}\n"
  USER_ID=""
fi

# Test 2: Login user
echo -e "${YELLOW}2. Testing POST /auth/login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"login":"testuser","password":"password123"}')

echo "Response: $LOGIN_RESPONSE"

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
  echo -e "${GREEN}✓ Login successful, token: ${TOKEN:0:20}...${NC}\n"
else
  echo -e "${RED}✗ Login failed${NC}\n"
fi

# Test 3: Create order
if [ -n "$TOKEN" ]; then
  echo -e "${YELLOW}3. Testing POST /orders/create (with token)${NC}"
  CREATE_ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/orders/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"amount":100}')

  echo "Response: $CREATE_ORDER_RESPONSE"

  if echo "$CREATE_ORDER_RESPONSE" | grep -q "[0-9]"; then
    echo -e "${GREEN}✓ Order created successfully${NC}\n"
  else
    echo -e "${RED}✗ Order creation failed${NC}\n"
  fi
fi

# Test 4: Get orders list
if [ -n "$TOKEN" ]; then
  echo -e "${YELLOW}4. Testing GET /orders/list (with token)${NC}"
  GET_ORDERS_RESPONSE=$(curl -s -X GET "$BASE_URL/orders/list?active=true" \
    -H "Authorization: Bearer $TOKEN")

  echo "Response: $GET_ORDERS_RESPONSE"

  if echo "$GET_ORDERS_RESPONSE" | grep -q "count"; then
    echo -e "${GREEN}✓ Orders list retrieved successfully${NC}\n"
  else
    echo -e "${RED}✗ Orders list retrieval failed${NC}\n"
  fi
fi

# Test 5: Test without token
echo -e "${YELLOW}5. Testing /orders/create (without token - should fail)${NC}"
NO_TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/orders/create" \
  -H "Content-Type: application/json" \
  -d '{"amount":100}')

echo "Response: $NO_TOKEN_RESPONSE"

if echo "$NO_TOKEN_RESPONSE" | grep -q "Unauthorized"; then
  echo -e "${GREEN}✓ Correctly rejected request without token${NC}\n"
else
  echo -e "${RED}✗ Should have rejected request without token${NC}\n"
fi

echo -e "${YELLOW}=== Tests Complete ===${NC}"
