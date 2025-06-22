#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing Update User Endpoint${NC}"
echo "================================"

# Clean up the database first
echo -e "\n${BLUE}Cleaning up database...${NC}"
docker exec -i alyatest-pg sh -c 'PGPASSWORD=alyatest psql -U alyatest -d alyatest -c "TRUNCATE TABLE users RESTART IDENTITY CASCADE;"' > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Database cleaned successfully${NC}"
else
    echo -e "${RED}Warning: Could not clean database${NC}"
fi

# First, create a user to update
echo -e "\n${BLUE}1. Creating a test user...${NC}"
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "Test User",
      "email": "test@valid.com",
      "username": "testuser",
      "phone_number": "+1234567890"
    }
  }')

echo "Response: $CREATE_RESPONSE"

# Extract the user ID from the response (now it's in data.id)
USER_ID=$(echo $CREATE_RESPONSE | jq -r '.data.id // empty')

if [ -z "$USER_ID" ]; then
    echo -e "${RED}Failed to create user or extract ID${NC}"
    exit 1
fi

echo -e "${GREEN}Created user with ID: $USER_ID${NC}"

# Test 1: Update name only
echo -e "\n${BLUE}2. Testing partial update - name only...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"name\": \"Updated Name\"
    }
  }" | jq '.'

# Test 2: Update email only
echo -e "\n${BLUE}3. Testing partial update - email only...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"email\": \"updated@valid.com\"
    }
  }" | jq '.'

# Test 3: Update multiple fields
echo -e "\n${BLUE}4. Testing multiple field update...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"name\": \"Final Name\",
      \"email\": \"final@valid.com\",
      \"phone_number\": \"+9876543210\"
    }
  }" | jq '.'

# Test 4: Invalid email format
echo -e "\n${BLUE}5. Testing validation - invalid email...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"email\": \"invalid-email\"
    }
  }" | jq '.'

# Test 5: Name too short
echo -e "\n${BLUE}6. Testing validation - name too short...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"name\": \"A\"
    }
  }" | jq '.'

# Test 6: Non-existent user
echo -e "\n${BLUE}7. Testing update on non-existent user...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "id": 99999,
      "name": "Should Fail"
    }
  }' | jq '.'

# Test 7: Missing user ID
echo -e "\n${BLUE}8. Testing with missing user ID...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "Should Fail"
    }
  }' | jq '.'

# Test 8: Banned email domain
echo -e "\n${BLUE}9. Testing banned email domain...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"email\": \"test@banned.com\"
    }
  }" | jq '.'

# Test 9: Empty request (no fields)
echo -e "\n${BLUE}10. Testing empty update request...${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID
    }
  }" | jq '.'

# Test 10: Multiple validation errors
echo -e "\n${BLUE}11. Testing multiple validation errors...${NC}"
echo -e "${BLUE}This will return language-independent error codes with msgid for multi-lingual support${NC}"
curl -X POST http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -d "{
    \"data\": {
      \"id\": $USER_ID,
      \"name\": \"A\",
      \"email\": \"not-an-email\",
      \"phone_number\": \"123\"
    }
  }" | jq '.'

echo -e "\n${GREEN}Update endpoint tests completed!${NC}"
echo -e "${BLUE}Note: Error messages contain msgid (101-106) for client-side translation${NC}"
echo -e "${BLUE}See messages.json for message templates in English and Hindi${NC}"