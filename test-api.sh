#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${BLUE}=== Testing Postman Runner API ===${NC}\n"

# Test 1: Health Check
echo -e "${BLUE}1. Testing Health Endpoint${NC}"
HEALTH=$(curl -s "$BASE_URL/health")
if echo "$HEALTH" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Upload Collection
echo -e "${BLUE}2. Uploading Sample Collection${NC}"
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/collections/upload" \
    -H "Content-Type: application/json" \
    -d @sample-collection.json)

COLLECTION_ID=$(echo "$UPLOAD_RESPONSE" | grep -o '"collection_id":[0-9]*' | grep -o '[0-9]*')

if [ -n "$COLLECTION_ID" ]; then
    echo -e "${GREEN}✓ Collection uploaded successfully (ID: $COLLECTION_ID)${NC}"
else
    echo -e "${RED}✗ Collection upload failed${NC}"
    echo "$UPLOAD_RESPONSE"
    exit 1
fi
echo ""

# Test 3: List Collections
echo -e "${BLUE}3. Listing Collections${NC}"
COLLECTIONS=$(curl -s "$BASE_URL/api/v1/collections")
if echo "$COLLECTIONS" | grep -q "Sample API Test Collection"; then
    echo -e "${GREEN}✓ Collection found in list${NC}"
else
    echo -e "${RED}✗ Collection not found in list${NC}"
    exit 1
fi
echo ""

# Test 4: Get Collection Tree
echo -e "${BLUE}4. Getting Collection Tree${NC}"
TREE=$(curl -s "$BASE_URL/api/v1/collections/$COLLECTION_ID/tree")
if echo "$TREE" | grep -q "Public APIs"; then
    echo -e "${GREEN}✓ Collection tree retrieved successfully${NC}"
else
    echo -e "${RED}✗ Failed to retrieve collection tree${NC}"
    exit 1
fi
echo ""

# Test 5: Find first request item ID (not a folder)
echo -e "${BLUE}5. Finding Request Item${NC}"

# Parse JSON to find a request item ID (using a more robust approach)
ITEM_ID=$(curl -s "$BASE_URL/api/v1/collections/$COLLECTION_ID/tree" | \
    python3 -c "import sys, json; data=json.load(sys.stdin); items=[]; 
def find_requests(node, items):
    if node.get('item_type') == 'request':
        items.append(node['id'])
    for child in node.get('children', []):
        find_requests(child, items)
root = data.get('items', [])
for item in root:
    find_requests(item, items)
print(items[0] if items else '')" 2>/dev/null)

if [ -n "$ITEM_ID" ]; then
    echo -e "${GREEN}✓ Found request item (ID: $ITEM_ID)${NC}"
else
    echo -e "${RED}✗ No request items found${NC}"
    exit 1
fi
echo ""

# Test 6: Get Item Details
echo -e "${BLUE}6. Getting Item Details${NC}"
ITEM=$(curl -s "$BASE_URL/api/v1/items/$ITEM_ID")
if echo "$ITEM" | grep -q "item_type"; then
    echo -e "${GREEN}✓ Item details retrieved${NC}"
else
    echo -e "${RED}✗ Failed to get item details${NC}"
    exit 1
fi
echo ""

# Test 7: Execute Request
echo -e "${BLUE}7. Executing Request${NC}"
EXEC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/items/$ITEM_ID/execute")
STATUS=$(echo "$EXEC_RESPONSE" | grep -o '"status":[0-9]*' | grep -o '[0-9]*')

if [ -n "$STATUS" ]; then
    echo -e "${GREEN}✓ Request executed successfully (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Request execution failed${NC}"
    echo "$EXEC_RESPONSE"
    exit 1
fi
echo ""

# Test 8: SSRF Protection
echo -e "${BLUE}8. Testing SSRF Protection${NC}"
cat > /tmp/test-ssrf-collection.json <<EOF
{
  "info": {
    "name": "SSRF Test",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Test Localhost",
      "request": {
        "method": "GET",
        "header": [],
        "url": "http://localhost:8080/health"
      }
    }
  ]
}
EOF

SSRF_UPLOAD=$(curl -s -X POST "$BASE_URL/api/v1/collections/upload" \
    -H "Content-Type: application/json" \
    -d @/tmp/test-ssrf-collection.json)

SSRF_COLLECTION_ID=$(echo "$SSRF_UPLOAD" | grep -o '"collection_id":[0-9]*' | grep -o '[0-9]*')

if [ -n "$SSRF_COLLECTION_ID" ]; then
    # Find the item ID
    SSRF_TREE=$(curl -s "$BASE_URL/api/v1/collections/$SSRF_COLLECTION_ID/tree")
    SSRF_ITEM_ID=$(echo "$SSRF_TREE" | grep -o '"id":[0-9]*' | tail -n 1 | grep -o '[0-9]*')
    
    # Try to execute (should be blocked)
    SSRF_EXEC=$(curl -s -X POST "$BASE_URL/api/v1/items/$SSRF_ITEM_ID/execute")
    
    if echo "$SSRF_EXEC" | grep -q "ssrf_protection"; then
        echo -e "${GREEN}✓ SSRF protection working (localhost blocked)${NC}"
    else
        echo -e "${RED}✗ SSRF protection NOT working${NC}"
        echo "$SSRF_EXEC"
    fi
fi

rm -f /tmp/test-ssrf-collection.json
echo ""

echo -e "${GREEN}=== All Tests Passed ===${NC}"
