# API Request & Response Examples

## Base URL
```
http://localhost:8080
```

---

## 1. Health Check

### Request
```bash
GET /health
```

### Response
```json
{
  "status": "healthy"
}
```

---

## 2. Upload Postman Collection

### Request
```bash
POST /api/v1/collections/upload
Content-Type: application/json

{
  "info": {
    "name": "My API Collection",
    "description": "Collection for testing APIs",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Users Folder",
      "item": [
        {
          "name": "Get All Users",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Accept",
                "value": "application/json"
              }
            ],
            "url": "https://jsonplaceholder.typicode.com/users"
          }
        },
        {
          "name": "Create User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}"
            },
            "url": "https://jsonplaceholder.typicode.com/users"
          }
        },
        {
          "name": "Update User (PUT)",
          "request": {
            "method": "PUT",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"id\": 1, \"name\": \"Jane Doe\", \"email\": \"jane@example.com\", \"username\": \"janedoe\"}"
            },
            "url": "https://jsonplaceholder.typicode.com/users/1"
          }
        },
        {
          "name": "Partial Update User (PATCH)",
          "request": {
            "method": "PATCH",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"email\": \"newemail@example.com\"}"
            },
            "url": "https://jsonplaceholder.typicode.com/users/1"
          }
        },
        {
          "name": "Delete User",
          "request": {
            "method": "DELETE",
            "header": [],
            "url": "https://jsonplaceholder.typicode.com/users/1"
          }
        }
      ]
    },
    {
      "name": "Get Single Post",
      "request": {
        "method": "GET",
        "header": [],
        "url": "https://jsonplaceholder.typicode.com/posts/1"
      }
    }
  ]
}
```

### Response (Success)
```json
{
  "collection_id": 1,
  "message": "Collection imported successfully"
}
```

### Response (Error - Invalid Schema)
```json
{
  "error": "validation_error",
  "message": "unsupported Postman schema version: v1.0 (only v2.0 and v2.1 are supported)"
}
```

### Response (Error - Invalid Method)
```json
{
  "error": "validation_error",
  "message": "invalid request 'Test Request': unsupported HTTP method: TRACE (allowed: GET, POST, PUT, PATCH, DELETE)"
}
```

---

## 3. List Collections

### Request
```bash
GET /api/v1/collections
```

### Response
```json
{
  "collections": [
    {
      "id": 1,
      "name": "My API Collection",
      "description": "Collection for testing APIs",
      "created_at": "2026-02-14T12:00:00Z",
      "updated_at": "2026-02-14T12:00:00Z"
    },
    {
      "id": 2,
      "name": "Another Collection",
      "description": "More tests",
      "created_at": "2026-02-14T13:00:00Z",
      "updated_at": "2026-02-14T13:00:00Z"
    }
  ]
}
```

---

## 4. Get Collection Tree

### Request
```bash
GET /api/v1/collections/1/tree
```

### Response
```json
{
  "collection": {
    "id": 1,
    "name": "My API Collection",
    "description": "Collection for testing APIs",
    "created_at": "2026-02-14T12:00:00Z",
    "updated_at": "2026-02-14T12:00:00Z"
  },
  "items": [
    {
      "id": 1,
      "name": "Users Folder",
      "item_type": "folder",
      "sort_order": 0,
      "children": [
        {
          "id": 2,
          "name": "Get All Users",
          "item_type": "request",
          "sort_order": 0,
          "method": "GET",
          "url": "https://jsonplaceholder.typicode.com/users",
          "headers": "[{\"key\":\"Accept\",\"value\":\"application/json\"}]",
          "children": []
        },
        {
          "id": 3,
          "name": "Create User",
          "item_type": "request",
          "sort_order": 1,
          "method": "POST",
          "url": "https://jsonplaceholder.typicode.com/users",
          "headers": "[{\"key\":\"Content-Type\",\"value\":\"application/json\"}]",
          "body": "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}",
          "children": []
        }
      ]
    },
    {
      "id": 4,
      "name": "Get Single Post",
      "item_type": "request",
      "sort_order": 1,
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/posts/1",
      "headers": "[]",
      "children": []
    }
  ]
}
```

### Response (Error - Not Found)
```json
{
  "error": "not_found",
  "message": "Collection not found"
}
```

---

## 5. Get Item Details

### Request
```bash
GET /api/v1/items/2
```

### Response (Folder)
```json
{
  "id": 1,
  "collection_id": 1,
  "name": "Users Folder",
  "item_type": "folder",
  "sort_order": 0,
  "created_at": "2026-02-14T12:00:00Z",
  "updated_at": "2026-02-14T12:00:00Z"
}
```

### Response (Request)
```json
{
  "id": 2,
  "collection_id": 1,
  "parent_id": 1,
  "name": "Get All Users",
  "item_type": "request",
  "sort_order": 0,
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/users",
  "headers": [
    {
      "key": "Accept",
      "value": "application/json"
    }
  ],
  "created_at": "2026-02-14T12:00:00Z",
  "updated_at": "2026-02-14T12:00:00Z"
}
```

### Response (Error - Not Found)
```json
{
  "error": "noSuccess - PUT Request)
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json; charset=utf-8"
  },
  "body": "{\"id\":1,\"name\":\"Jane Doe\",\"email\":\"jane@example.com\",\"username\":\"janedoe\"}",
  "duration_ms": 187
}
```

### Response (Success - PATCH Request)
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json; charset=utf-8"
  },
  "body": "{\"id\":1,\"name\":\"Leanne Graham\",\"email\":\"newemail@example.com\",\"username\":\"Bret\"}",
  "duration_ms": 156
}
```

### Response (Success - DELETE Request)
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json; charset=utf-8"
  },
  "body": "{}",
  "duration_ms": 143
}
```

### Response (t_found",
  "message": "Item not found"
}
```

---

## 6. Update Item (Request)

### Request (Full Update - PUT)
```bash
PUT /api/v1/items/2
Content-Type: application/json

{
  "name": "Get All Users (Updated)",
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/users?_limit=10",
  "headers": [
    {
      "key": "Accept",
      "value": "application/json"
    },
    {
      "key": "User-Agent",
      "value": "PostmanRunner/1.0"
    }
  ]
}
```

### Request (Partial Update - PUT)
```bash
PUT /api/v1/items/3
Content-Type: application/json

{
  "url": "https://jsonplaceholder.typicode.com/users/2",
  "headers": [
    {
      "key": "Content-Type",
      "value": "application/json"
    },
    {
      "key": "Authorization",
      "value": "Bearer token123"
    }
  ],
  "body": "{\"name\": \"Updated Name\", \"email\": \"updated@example.com\"}"
}
```

### Response (Success)
```json
{
  "message": "Item updated successfully"
}
```

### Response (Error - Not Found)
```json
{
  "error": "not_found",
  "message": "Item not found"
}
```

### Response (Error - Invalid Item Type)
```json
{
  "error": "invalid_item_type",
  "message": "Only items of type 'request' can be updated with this endpoint"
}
```

### Response (Error - Invalid Method)
```9son
{
  "error": "invalid_method",
  "message": "Method must be one of: GET, POST, PUT, PATCH, DELETE"
}
```

---

## 7. Delete Item

### Request
```bash
DELETE /api/v1/items/5
```

### Response (Success)
```json
{
  "message": "Item deleted successfully"
}
```

### Response (Error - Not Found)
```json
{
  "error": "not_found",
  "message": "Item not found"
}
```

**Note**: Deleting a folder will also delete all its child items (cascading delete).

---

## 8. Execute Request

### Request
```bash
POST /api/v1/items/2/execute
```

### Response (Success - GET Request)
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": "5645",
    "Date": "Fri, 14 Feb 2026 12:00:00 GMT",
    "Server": "cloudflare"
  },
  "body": "[{\"id\":1,\"name\":\"Leanne Graham\",\"username\":\"Bret\",\"email\":\"Sincere@april.biz\"},{\"id\":2,\"name\":\"Ervin Howell\",\"username\":\"Antonette\",\"email\":\"Shanna@melissa.tv\"}]",
  "duration_ms": 245
}
```

### Response (Success - POST Request)
```json
{
  "status": 201,
  "headers": {
    "Content-Type": "application/json; charset=utf-8",
    "Location": "/users/101"
  },
  "body": "{\"id\":101,\"name\":\"John Doe\",\"email\":\"john@example.com\"}",
  "duration_ms": 312
}
```

### Response (Error - Invalid Item Type)
```json
{
  "error": "invalid_item_type",
  "message": "Only items of type 'request' can be executed"
}
```

### Response (Error - SSRF Protection)
```json
{
  "error": "ssrf_protection",
  "message": "URL blocked by SSRF protection: requests to private IP ranges are not allowed: 169.254.169.254"
}
```

### Response (Error - Execution Failed)
```json
{
  "error": "execution_error",
  "message": "Failed to execute request: request failed: Get \"https://invalid-domain-xyz.com\": dial tcp: lookup invalid-domain-xyz.com: no such host"
}
```

### Response (Error - Timeout)
```json
{
  "error": "execution_error",
  "message": "Failed to execute request: request failed: context deadline exceeded"
}
```

---

## 7. Rate Limiting

### Response (Rate Limit Exceeded)
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests. Please try again later."
}
```

---

## cURL Examples

### Upload Collection
```bash
curl -X POST http://localhost:8080/api/v1/collections/upload \
  -H "Content-Type: application/json" \
  -d @my-collection.json
```

### List Collections
```bash
curl http://localhost:8080/api/v1/collections
```

### Get Collection Tree
```bash
curl http://localhost:8080/api/v1/collections/1/tree
```

### Execute Request
```bash
curl -X POST http://localhost:8080/api/v1/items/2/execute
```

### Update Item
```bash
curl -X PUT http://localhost:8080/api/v1/items/2 \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Request", "url": "https://api.example.com/v2/users"}'
```

### Delete Item
```bash
curl -X DELETE http://localhost:8080/api/v1/items/5
```

### Execute Request with Pretty Output
```bash
curl -X POST http://localhost:8080/api/v1/items/2/execute | jq
```

---

## HTTP Status Codes

| Status Code | Description | When Used |
|-------------|-------------|-----------|
| 200 | OK | Successful GET/list operations |
| 201 | Created | Collection uploaded successfully |
| 400 | Bad Request | Invalid input, validation errors |
| 403 | Forbidden | SSRF protection blocked the request |
| 404 | Not Found | Collection or item not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Database or system errors |
| 502 | Bad Gateway | Request execution failed |

---

## Error Response Format
invalid_json` - Invalid JSON in request body
- `invalid_method` - HTTP method not allowed
- `invalid_headers` - Failed to serialize headers
- `no_updates` - No fields provided for update
- `not_found` - Resource doesn't exist
- `invalid_item_type` - Trying to execute/upda
```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

### Common Error Types
- `validation_error` - Invalid Postman collection format
- `invalid_id` - ID parameter is not a valid integer
- `not_found` - Resource doesn't exist
- `invalid_item_type` - Trying to execute a folder
- `invalid_request` - Request missing required fields
- `ssrf_protection` - URL blocked by security
- `rate_limit_exceeded` - Too many requests
- `database_error` - Database operation failed
- `execution_error` - HTTP request execution failed
- `import_error` - Failed to import collection items
