# Transaction Split API

A RESTful API for managing shared expenses and transaction splitting within groups. This API allows users to create groups, track transactions, and split costs among group members.

## Base URL

```
http://localhost:8080
```

## Table of Contents

- [API Structure](#api-structure)
- [Users](#users)
- [Groups](#groups)
- [Group Members](#group-members)
- [Transactions](#transactions)
- [Splits](#splits)
- [Group Balances](#group-balances)
- [Error Handling](#error-handling)

---

## API Structure

### Nested Resource Routes (Recommended)

The API supports intuitive nested routes that follow the resource hierarchy:

- `/groups/{group_id}/members` - Members within a group
- `/groups/{group_id}/transactions` - Transactions within a group
- `/groups/{group_id}/balances` - Balance reports for a group
- `/transactions/{transaction_id}/splits` - Splits within a transaction
- `/users/{user_id}/transactions` - Transactions created by a user

### Direct Access Routes (Backwards Compatible)

For flexibility, the API also provides direct access to resources:

- `/group_members/` - Direct group member management
- `/transactions/` - Direct transaction management
- `/splits/` - Direct split management

Both approaches are fully supported and can be used interchangeably.

---

## Users

Manage user accounts in the system.

### List Users

Retrieve a paginated list of all users.

**Endpoint:** `GET /users/`

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of users to return |
| `offset` | integer | No | 0 | Number of users to skip |

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": 1,
      "name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

### Get User by ID

Retrieve a specific user by their ID.

**Endpoint:** `GET /users/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | User ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid user ID format
- `404 Not Found` - User not found

### Create User

Create a new user.

**Endpoint:** `POST /users/`

**Request Body:**
```json
{
  "name": "John Doe"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | User's name |

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### Update User

Update an existing user's information.

**Endpoint:** `PUT /users/{id}` or `PATCH /users/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | User ID |

**Request Body:**
```json
{
  "name": "Jane Doe"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | User's updated name |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Jane Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T12:45:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid user ID or request body
- `404 Not Found` - User not found

### Delete User

Delete a user from the system.

**Endpoint:** `DELETE /users/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | User ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid user ID format
- `404 Not Found` - User not found or unable to delete

---

## Groups

Manage groups for organizing shared expenses.

### List Groups

Retrieve a paginated list of all groups.

**Endpoint:** `GET /groups/`

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of groups to return |
| `offset` | integer | No | 0 | Number of groups to skip |

**Response:** `200 OK`
```json
{
  "groups": [
    {
      "id": 1,
      "name": "Roommates"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

### Get Group by ID

Retrieve a specific group by its ID.

**Endpoint:** `GET /groups/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Roommates"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID format
- `404 Not Found` - Group not found

### Create Group

Create a new group.

**Endpoint:** `POST /groups/`

**Request Body:**
```json
{
  "name": "Roommates"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Group name |

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Roommates"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### Update Group

Update an existing group's information.

**Endpoint:** `PUT /groups/{id}` or `PATCH /groups/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group ID |

**Request Body:**
```json
{
  "name": "House Expenses"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Group's updated name |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "House Expenses"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID or request body
- `404 Not Found` - Group not found

### Delete Group

Delete a group from the system.

**Endpoint:** `DELETE /groups/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Roommates"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID format
- `404 Not Found` - Group not found or unable to delete

---

## Group Members

Manage memberships of users within groups.

### List Group Members (Nested Route)

Retrieve a paginated list of all members in a specific group.

**Endpoint:** `GET /groups/{group_id}/members`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of members to return |
| `offset` | integer | No | 0 | Number of members to skip |

**Response:** `200 OK`
```json
{
  "group_members": [
    {
      "id": 1,
      "group_id": 1,
      "group_name": "Roommates",
      "member_name": "John Doe",
      "user_id": 1,
      "user_name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID format

### Create Group Member (Nested Route)

Add a user to a group using the nested route.

**Endpoint:** `POST /groups/{group_id}/members`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Request Body:**
```json
{
  "user_id": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user_id` | integer | No | User ID (nullable) |

**Response:** `201 Created`
```json
{
  "id": 1,
  "group_id": 1,
  "member_name": "John Doe",
  "user_id": 1,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### List Group Members by Group ID (Direct Access)

Retrieve a paginated list of all members in a specific group using direct access route.

**Endpoint:** `GET /group_members/group/{group_id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of members to return |
| `offset` | integer | No | 0 | Number of members to skip |

**Response:** `200 OK`
```json
{
  "group_members": [
    {
      "id": 1,
      "group_id": 1,
      "group_name": "Roommates",
      "member_name": "John Doe",
      "user_id": 1,
      "user_name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID format

### Get Group Member by ID

Retrieve a specific group member by their ID.

**Endpoint:** `GET /group_members/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group Member ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "group_name": "Roommates",
  "member_name": "John Doe",
  "user_id": 1,
  "user_name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group member ID format
- `404 Not Found` - Group member not found

### Create Group Member (Direct Access)

Add a user to a group using direct access route.

**Endpoint:** `POST /group_members/`

**Request Body:**
```json
{
  "group_id": 1,
  "user_id": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |
| `user_id` | integer | No | User ID (nullable) |

**Response:** `201 Created`
```json
{
  "id": 1,
  "group_id": 1,
  "member_name": "John Doe",
  "user_id": 1,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### Update Group Member

Update a group member's information.

**Endpoint:** `PUT /group_members/{id}` or `PATCH /group_members/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group Member ID |

**Request Body:**
```json
{
  "group_id": 1,
  "user_id": 2
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |
| `user_id` | integer | No | User ID (nullable) |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "member_name": "Jane Smith",
  "user_id": 2,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group member ID or request body
- `404 Not Found` - Group member not found

### Delete Group Member

Remove a user from a group.

**Endpoint:** `DELETE /group_members/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Group Member ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "member_name": "John Doe",
  "user_id": 1,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group member ID format
- `404 Not Found` - Group member not found or unable to delete

---

## Transactions

Manage financial transactions within groups.

### List All Transactions

Retrieve a paginated list of all transactions.

**Endpoint:** `GET /transactions/`

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of transactions to return |
| `offset` | integer | No | 0 | Number of transactions to skip |

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "id": 1,
      "group_id": 1,
      "name": "Grocery Shopping",
      "transaction_date": "2024-01-15T00:00:00Z",
      "amount": "125.50",
      "category": "Groceries",
      "note": "Weekly shopping at Whole Foods",
      "by_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

### List Transactions by User (Nested Route)

Retrieve all transactions created by a specific user using the nested route.

**Endpoint:** `GET /users/{user_id}/transactions`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | integer | Yes | User ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of transactions to return |
| `offset` | integer | No | 0 | Number of transactions to skip |

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "id": 1,
      "group_id": 1,
      "name": "Grocery Shopping",
      "transaction_date": "2024-01-15T00:00:00Z",
      "amount": "125.50",
      "category": "Groceries",
      "note": "Weekly shopping at Whole Foods",
      "by_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid user ID format

### List Transactions by Group (Nested Route)

Retrieve all transactions for a specific group within a date range using the nested route.

**Endpoint:** `GET /groups/{group_id}/transactions`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_date` | string | Yes | - | Start date in YYYY-MM-DD format |
| `end_date` | string | Yes | - | End date in YYYY-MM-DD format |
| `limit` | integer | No | 100 | Maximum number of transactions to return |
| `offset` | integer | No | 0 | Number of transactions to skip |

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "id": 1,
      "group_id": 1,
      "name": "Grocery Shopping",
      "transaction_date": "2024-01-15T00:00:00Z",
      "amount": "125.50",
      "category": "Groceries",
      "note": "Weekly shopping at Whole Foods",
      "by_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID or missing/invalid date parameters

### Create Transaction (Nested Route)

Create a new transaction within a group using the nested route.

**Endpoint:** `POST /groups/{group_id}/transactions`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Request Body:**
```json
{
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Transaction name |
| `transaction_date` | string (ISO 8601) | Yes | Date of transaction |
| `amount` | string (decimal) | Yes | Transaction amount |
| `category` | string | No | Transaction category (nullable) |
| `note` | string | No | Additional notes (nullable) |
| `by_user` | integer | Yes | ID of user who created the transaction |

**Note:** The `group_id` from the URL path is used; any `group_id` in the request body is ignored.

**Response:** `201 Created`
```json
{
  "id": 1,
  "group_id": 1,
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### List Transactions by Group (Direct Access)

Retrieve all transactions for a specific group within a date range using direct access route.

**Endpoint:** `GET /transactions/group/{group_id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_date` | string | Yes | - | Start date in YYYY-MM-DD format |
| `end_date` | string | Yes | - | End date in YYYY-MM-DD format |
| `limit` | integer | No | 100 | Maximum number of transactions to return |
| `offset` | integer | No | 0 | Number of transactions to skip |

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "id": 1,
      "group_id": 1,
      "name": "Grocery Shopping",
      "transaction_date": "2024-01-15T00:00:00Z",
      "amount": "125.50",
      "category": "Groceries",
      "note": "Weekly shopping at Whole Foods",
      "by_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID or missing/invalid date parameters

### Get Transaction by ID

Retrieve a specific transaction by its ID.

**Endpoint:** `GET /transactions/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Transaction ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid transaction ID format
- `404 Not Found` - Transaction not found

### Create Transaction (Direct Access)

Create a new transaction using direct access route.

**Endpoint:** `POST /transactions/`

**Request Body:**
```json
{
  "group_id": 1,
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |
| `name` | string | Yes | Transaction name |
| `transaction_date` | string (ISO 8601) | Yes | Date of transaction |
| `amount` | string (decimal) | Yes | Transaction amount |
| `category` | string | No | Transaction category (nullable) |
| `note` | string | No | Additional notes (nullable) |
| `by_user` | integer | Yes | ID of user who created the transaction |

**Response:** `201 Created`
```json
{
  "id": 1,
  "group_id": 1,
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### Update Transaction

Update an existing transaction.

**Endpoint:** `PUT /transactions/{id}` or `PATCH /transactions/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Transaction ID |

**Request Body:**
```json
{
  "group_id": 1,
  "name": "Grocery Shopping - Updated",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "135.75",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods - Updated total",
  "by_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |
| `name` | string | Yes | Transaction name |
| `transaction_date` | string (ISO 8601) | Yes | Date of transaction |
| `amount` | string (decimal) | Yes | Transaction amount |
| `category` | string | No | Transaction category (nullable) |
| `note` | string | No | Additional notes (nullable) |
| `by_user` | integer | Yes | ID of user who created the transaction |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "name": "Grocery Shopping - Updated",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "135.75",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods - Updated total",
  "by_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T12:45:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid transaction ID or request body
- `404 Not Found` - Transaction not found

### Delete Transaction

Delete a transaction from the system.

**Endpoint:** `DELETE /transactions/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Transaction ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "group_id": 1,
  "name": "Grocery Shopping",
  "transaction_date": "2024-01-15T00:00:00Z",
  "amount": "125.50",
  "category": "Groceries",
  "note": "Weekly shopping at Whole Foods",
  "by_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid transaction ID format
- `404 Not Found` - Transaction not found or unable to delete

---

## Splits

Manage how transaction costs are split among group members.

### List All Splits

Retrieve a paginated list of all splits.

**Endpoint:** `GET /splits/`

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of splits to return |
| `offset` | integer | No | 0 | Number of splits to skip |

**Response:** `200 OK`
```json
{
  "splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "50.00",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

### List Splits by Transaction (Nested Route)

Retrieve all splits for a specific transaction using the nested route.

**Endpoint:** `GET /transactions/{transaction_id}/splits`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |

**Response:** `200 OK`
```json
{
  "splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "50.00",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 1,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid transaction ID format

### Create Split (Nested Route)

Create a new split for a transaction using the nested route.

**Endpoint:** `POST /transactions/{transaction_id}/splits`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |

**Request Body:**
```json
{
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `split_percent` | string (decimal) | Yes | Percentage of transaction amount |
| `split_amount` | string (decimal) | Yes | Amount assigned to this split |
| `split_user` | integer | No | User ID responsible for this split (nullable) |

**Note:** The `transaction_id` from the URL path is used; any `transaction_id` in the request body is ignored.

**Response:** `201 Created`
```json
{
  "id": 1,
  "transaction_id": 1,
  "tx_amount": "125.50",
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### List Splits by Transaction (Direct Access)

Retrieve all splits for a specific transaction using direct access route.

**Endpoint:** `GET /splits/transaction/{transaction_id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |

**Response:** `200 OK`
```json
{
  "splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "50.00",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 1,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid transaction ID format

### List Splits by User

Retrieve all splits assigned to a specific user.

**Endpoint:** `GET /splits/user/{user_id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | integer | Yes | User ID |

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 100 | Maximum number of splits to return |
| `offset` | integer | No | 0 | Number of splits to skip |

**Response:** `200 OK`
```json
{
  "splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "50.00",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "limit": 100,
  "offset": 0
}
```

**Error Responses:**
- `400 Bad Request` - Invalid user ID format

### Get Split by ID

Retrieve a specific split by its ID.

**Endpoint:** `GET /splits/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Split ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "transaction_id": 1,
  "tx_amount": "125.50",
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid split ID format
- `404 Not Found` - Split not found

### Create Split (Direct Access)

Create a new split for a transaction using direct access route.

**Endpoint:** `POST /splits/`

**Request Body:**
```json
{
  "transaction_id": 1,
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |
| `split_percent` | string (decimal) | Yes | Percentage of transaction amount |
| `split_amount` | string (decimal) | Yes | Amount assigned to this split |
| `split_user` | integer | No | User ID responsible for this split (nullable) |

**Response:** `201 Created`
```json
{
  "id": 1,
  "transaction_id": 1,
  "tx_amount": "125.50",
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields

### Update Split

Update an existing split.

**Endpoint:** `PUT /splits/{id}` or `PATCH /splits/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Split ID |

**Request Body:**
```json
{
  "split_percent": "60.00",
  "split_amount": "75.30",
  "split_user": 1
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `split_percent` | string (decimal) | Yes | Percentage of transaction amount |
| `split_amount` | string (decimal) | Yes | Amount assigned to this split |
| `split_user` | integer | No | User ID responsible for this split (nullable) |

**Response:** `200 OK`
```json
{
  "id": 1,
  "transaction_id": 1,
  "tx_amount": "125.50",
  "split_percent": "60.00",
  "split_amount": "75.30",
  "split_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T12:45:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid split ID or request body
- `404 Not Found` - Split not found

### Delete Split

Delete a split from the system.

**Endpoint:** `DELETE /splits/{id}`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | Yes | Split ID |

**Response:** `200 OK`
```json
{
  "id": 1,
  "transaction_id": 1,
  "tx_amount": "125.50",
  "split_percent": "50.00",
  "split_amount": "62.75",
  "split_user": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid split ID format
- `404 Not Found` - Split not found or unable to delete

---

## Group Balances

Retrieve balance and settlement information for groups.

### Get Group Balances

Retrieve comprehensive balance information for a group, including detailed debts, net balances, and simplified settlement paths.

**Endpoint:** `GET /groups/{group_id}/balances`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Response:** `200 OK`
```json
{
  "group_id": 1,
  "balances": [
    {
      "creditor": "John Doe",
      "debtor": "Jane Smith",
      "total_owed": 12550
    }
  ],
  "net_balances": [
    {
      "member_name": "John Doe",
      "net_balance": 12550
    },
    {
      "member_name": "Jane Smith",
      "net_balance": -12550
    }
  ],
  "simplified_owes": [
    {
      "creditor": "John Doe",
      "debtor": "Jane Smith",
      "total_owed": 12550
    }
  ],
  "count": 1,
  "net_count": 2,
  "simplified_count": 1
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `group_id` | integer | The group ID |
| `balances` | array | Detailed pairwise debts between all members |
| `net_balances` | array | Net position for each member (positive = owed, negative = owes) |
| `simplified_owes` | array | Optimized settlement paths to minimize transactions |
| `count` | integer | Number of detailed balance entries |
| `net_count` | integer | Number of members with balances |
| `simplified_count` | integer | Number of simplified settlement transactions |

**Balance Types Explained:**

1. **Balances** - Shows all pairwise debts between members. Useful for understanding the complete debt structure.

2. **Net Balances** - Shows each member's overall position in the group:
   - Positive values indicate the member is owed money
   - Negative values indicate the member owes money
   - Zero indicates the member is settled up

3. **Simplified Owes** - Shows the minimum number of transactions needed to settle all debts. This is the recommended view for settling up as it minimizes the number of payments needed.

**Note:** All amount values are in cents (e.g., 12550 = $125.50).

**Error Responses:**
- `400 Bad Request` - Invalid group ID format
- `500 Internal Server Error` - Error calculating balances

**Example Use Cases:**

- **Viewing who owes whom:** Use the `balances` array
- **Checking your overall position:** Use the `net_balances` array
- **Settling up efficiently:** Use the `simplified_owes` array

---

## Error Handling

The API uses standard HTTP status codes to indicate success or failure of requests.

### HTTP Status Codes

| Status Code | Description |
|------------|-------------|
| `200 OK` | Request succeeded |
| `201 Created` | Resource successfully created |
| `400 Bad Request` | Invalid request format, missing required fields, or invalid parameters |
| `404 Not Found` | Requested resource not found |
| `500 Internal Server Error` | Server encountered an unexpected error |

### Error Response Format

Error responses are returned as plain text messages:

```
Invalid user ID format
```

or

```
User not found
```

---

## Data Types

### Decimal Fields

Fields containing decimal values (amounts, percentages) are represented as strings to maintain precision:
- `amount`: Transaction amounts (e.g., `"125.50"`)
- `split_amount`: Split amounts (e.g., `"62.75"`)
- `split_percent`: Percentage values (e.g., `"50.00"`)

### Timestamps

All timestamp fields use ISO 8601 format:
- Format: `2024-01-15T10:30:00Z`
- Timezone: UTC (indicated by `Z`)

### Nullable Fields

Some fields may be `null`:
- `category` (in transactions)
- `note` (in transactions)
- `user_id` (in group members and splits)
- `split_user` (in splits)
- `member_name` (in group members)

---

## Getting Started

1. Ensure the API server is running on port 8080
2. Use a tool like `curl`, Postman, or your preferred HTTP client
3. All requests should use `Content-Type: application/json` header when sending JSON data
4. All responses will have `Content-Type: application/json` header

### Example Request

```bash
curl -X POST http://localhost:8080/users/ \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe"}'
```

### Example Response

```json
{
  "id": 1,
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "modified_at": "2024-01-15T10:30:00Z"
}
```

