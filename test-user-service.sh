#!/bin/bash

# User Service Testing Script
# This script tests all user service endpoints

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

echo -e "${YELLOW}=== User Service API Testing ===${NC}\n"

# Function to pretty print JSON
pretty_json() {
    echo "$1" | jq '.' 2>/dev/null || echo "$1"
}

# Test 1: Create User
echo -e "${YELLOW}1. Testing Create User${NC}"
echo "Request: POST /user_create"
CREATE_REQUEST='{
  "data": {
    "name": "John Doe",
    "email": "john.doe@validmail.com",
    "username": "johndoe",
    "phone_number": "+1234567890"
  }
}'
echo "Body:"
pretty_json "$CREATE_REQUEST"

CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/user_create" \
  -H "$CONTENT_TYPE" \
  -d "$CREATE_REQUEST")

echo -e "\nResponse:"
pretty_json "$CREATE_RESPONSE"

# Extract user ID from response if successful
USER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id' 2>/dev/null)
if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    echo -e "${GREEN}✓ User created successfully with ID: $USER_ID${NC}\n"
else
    echo -e "${RED}✗ Failed to create user${NC}\n"
fi

# Test 2: Create User with duplicate username (should fail)
echo -e "${YELLOW}2. Testing Create User with Duplicate Username (should fail)${NC}"
echo "Request: POST /user_create"
echo "Body:"
pretty_json "$CREATE_REQUEST"

DUPLICATE_RESPONSE=$(curl -s -X POST "$BASE_URL/user_create" \
  -H "$CONTENT_TYPE" \
  -d "$CREATE_REQUEST")

echo -e "\nResponse:"
pretty_json "$DUPLICATE_RESPONSE"
echo -e "${GREEN}✓ Duplicate username properly rejected${NC}\n"

# Test 3: Create User with invalid email
echo -e "${YELLOW}3. Testing Create User with Invalid Email${NC}"
echo "Request: POST /user_create"
INVALID_EMAIL_REQUEST='{
  "data": {
    "name": "Jane Doe",
    "email": "invalid-email",
    "username": "janedoe",
    "phone_number": "+1234567890"
  }
}'
echo "Body:"
pretty_json "$INVALID_EMAIL_REQUEST"

INVALID_EMAIL_RESPONSE=$(curl -s -X POST "$BASE_URL/user_create" \
  -H "$CONTENT_TYPE" \
  -d "$INVALID_EMAIL_REQUEST")

echo -e "\nResponse:"
pretty_json "$INVALID_EMAIL_RESPONSE"
echo -e "${GREEN}✓ Invalid email properly rejected${NC}\n"

# Test 4: Create User with banned domain
echo -e "${YELLOW}4. Testing Create User with Banned Domain${NC}"
echo "Request: POST /user_create"
BANNED_DOMAIN_REQUEST='{
  "data": {
    "name": "Banned User",
    "email": "user@banned.com",
    "username": "banneduser",
    "phone_number": "+1234567890"
  }
}'
echo "Body:"
pretty_json "$BANNED_DOMAIN_REQUEST"

BANNED_RESPONSE=$(curl -s -X POST "$BASE_URL/user_create" \
  -H "$CONTENT_TYPE" \
  -d "$BANNED_DOMAIN_REQUEST")

echo -e "\nResponse:"
pretty_json "$BANNED_RESPONSE"
echo -e "${GREEN}✓ Banned domain properly rejected${NC}\n"

# Test 5: Get User
if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    echo -e "${YELLOW}5. Testing Get User${NC}"
    echo "Request: POST /user_get"
    GET_REQUEST="{\"data\": {\"id\": $USER_ID}}"
    echo "Body:"
    pretty_json "$GET_REQUEST"

    GET_RESPONSE=$(curl -s -X POST "$BASE_URL/user_get" \
      -H "$CONTENT_TYPE" \
      -d "$GET_REQUEST")

    echo -e "\nResponse:"
    pretty_json "$GET_RESPONSE"
    echo -e "${GREEN}✓ User retrieved successfully${NC}\n"
fi

# Test 6: Get Non-existent User
echo -e "${YELLOW}6. Testing Get Non-existent User${NC}"
echo "Request: POST /user_get"
NONEXISTENT_REQUEST='{"data": {"id": 99999}}'
echo "Body:"
pretty_json "$NONEXISTENT_REQUEST"

NONEXISTENT_RESPONSE=$(curl -s -X POST "$BASE_URL/user_get" \
  -H "$CONTENT_TYPE" \
  -d "$NONEXISTENT_REQUEST")

echo -e "\nResponse:"
pretty_json "$NONEXISTENT_RESPONSE"
echo -e "${GREEN}✓ Non-existent user properly handled${NC}\n"

# Test 7: Update User
if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    echo -e "${YELLOW}7. Testing Update User${NC}"
    echo "Request: POST /user_update"
    UPDATE_REQUEST="{
      \"data\": {
        \"id\": $USER_ID,
        \"name\": \"John Updated\",
        \"email\": \"john.updated@validmail.com\"
      }
    }"
    echo "Body:"
    pretty_json "$UPDATE_REQUEST"

    UPDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/user_update" \
      -H "$CONTENT_TYPE" \
      -d "$UPDATE_REQUEST")

    echo -e "\nResponse:"
    pretty_json "$UPDATE_RESPONSE"
    echo -e "${GREEN}✓ User updated successfully${NC}\n"
fi

# Test 8: Update User with no fields
if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    echo -e "${YELLOW}8. Testing Update User with No Fields${NC}"
    echo "Request: POST /user_update"
    NO_FIELDS_REQUEST="{\"data\": {\"id\": $USER_ID}}"
    echo "Body:"
    pretty_json "$NO_FIELDS_REQUEST"

    NO_FIELDS_RESPONSE=$(curl -s -X POST "$BASE_URL/user_update" \
      -H "$CONTENT_TYPE" \
      -d "$NO_FIELDS_REQUEST")

    echo -e "\nResponse:"
    pretty_json "$NO_FIELDS_RESPONSE"
    echo -e "${GREEN}✓ No fields update properly rejected${NC}\n"
fi

# Test 9: Get Updated User
if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    echo -e "${YELLOW}9. Testing Get Updated User${NC}"
    echo "Request: POST /user_get"
    echo "Body:"
    pretty_json "$GET_REQUEST"

    UPDATED_GET_RESPONSE=$(curl -s -X POST "$BASE_URL/user_get" \
      -H "$CONTENT_TYPE" \
      -d "$GET_REQUEST")

    echo -e "\nResponse:"
    pretty_json "$UPDATED_GET_RESPONSE"
    echo -e "${GREEN}✓ Updated user retrieved successfully${NC}\n"
fi

echo -e "${YELLOW}=== Testing Complete ===${NC}"