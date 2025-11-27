# Transaction Split API

A RESTful API for managing shared expenses and transaction splitting within groups. This API allows users to create groups, track transactions, and split costs among group members.

## Base URL

```
http://localhost:8080
```

## Quick Reference - All Routes

### Public Routes (No Authentication Required)
- `GET /` - Health check / Hello World
- `POST /auth/register` - Register new user
- `POST /auth/login` - Login
- `POST /auth/refresh` - Refresh access token

### Protected Routes (Authentication + CSRF Required)

#### Authentication
- `GET /auth/me` - Get current authenticated user
- `POST /auth/logout` - Logout (revoke refresh token)
- `GET /auth/csrf-token` - Get CSRF token

#### Users
- `GET /users/` - List users (paginated, filtered by authenticated user)
- `POST /users/` - Create user
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `PATCH /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user
- `GET /users/{user_id}/transactions` - List transactions by user
- `GET /users/{user_id}/balances` - Get user balances across all groups

#### Groups
- `GET /groups/` - List groups (filtered by authenticated user's membership)
- `POST /groups/` - Create group (creator automatically added as member)
- `GET /groups/{id}` - Get group by ID
- `PUT /groups/{id}` - Update group
- `PATCH /groups/{id}` - Update group
- `DELETE /groups/{id}` - Delete group
- `GET /groups/{group_id}/members` - List group members
- `POST /groups/{group_id}/members` - Add member to group
- `POST /groups/{group_id}/members/batch` - Add multiple members (batch)
- `PUT /groups/{group_id}/members/batch` - Replace all members (batch)
- `PATCH /groups/{group_id}/members/batch` - Replace all members (batch)
- `DELETE /groups/{group_id}/members/batch` - Delete all members (batch)
- `GET /groups/{group_id}/transactions` - List group transactions (with date range)
- `POST /groups/{group_id}/transactions` - Create transaction in group
- `GET /groups/{group_id}/balances` - Get group balance report

#### Group Members (Direct Access)
- `POST /group_members/` - Create group member
- `GET /group_members/{id}` - Get group member by ID
- `PUT /group_members/{id}` - Update group member
- `PATCH /group_members/{id}` - Update group member
- `DELETE /group_members/{id}` - Delete/unlink group member
- `GET /group_members/group/{group_id}` - List members by group ID

#### Transactions
- `GET /transactions/` - List transactions (filtered by authenticated user's groups)
- `POST /transactions/` - Create transaction
- `GET /transactions/{id}` - Get transaction by ID
- `PUT /transactions/{id}` - Update transaction
- `PATCH /transactions/{id}` - Update transaction
- `DELETE /transactions/{id}` - Delete transaction
- `GET /transactions/{transaction_id}/splits` - List splits for transaction
- `POST /transactions/{transaction_id}/splits` - Create split for transaction

#### Splits
- `GET /splits/` - List splits (filtered by authenticated user's groups)
- `GET /splits/{id}` - Get split by ID
- `GET /splits/transaction/{transaction_id}` - List splits by transaction ID
- `GET /splits/user/{user_id}` - List splits by user ID
- `POST /splits/transaction/{transaction_id}/batch` - Create/replace all splits (batch, recommended)
- `PUT /splits/transaction/{transaction_id}/batch` - Replace all splits (batch, recommended)
- `PATCH /splits/transaction/{transaction_id}/batch` - Replace all splits (batch, recommended)

**Note:** All protected routes require:
1. Valid authentication token (via cookie or Authorization header)
2. Valid CSRF token (for state-changing operations: POST, PUT, PATCH, DELETE)

## Table of Contents

- [API Structure](#api-structure)
- [Authentication](#authentication)
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
- `/users/{user_id}/balances` - Balance reports for a User (group agnostic)

### Direct Access Routes (Backwards Compatible)

For flexibility, the API also provides direct access to resources:

- `/group_members/` - Direct group member management
- `/transactions/` - Direct transaction management
- `/splits/` - Direct split management

Both approaches are fully supported and can be used interchangeably.

---

## Authentication

The API uses JWT-based authentication with refresh tokens and CSRF protection. Most endpoints require authentication, except for registration, login, and token refresh.

### Register

Create a new user account.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | User's full name |
| `email` | string | Yes | User's email address (must be unique) |
| `password` | string | Yes | Password (minimum 8 characters) |

**Response:** `201 Created`
```json
{
  "token": "token_value",
  "csrf_token": "csrf_token_value",
  "user": {
    "id": 1,
    "name": "John Doe",
    "created_at": "2024-01-15T10:30:00Z",
    "modified_at": "2024-01-15T10:30:00Z"
  }
}
```

**Cookies Set:**
- `access_token` - JWT access token (30 minutes)
- `refresh_token` - Refresh token (7 days, configurable)
- `csrf_token` - CSRF protection token (24 hours)

**Error Responses:**
- `400 Bad Request` - Invalid JSON, missing required fields, or password too short
- `409 Conflict` - Email already registered

### Login

Authenticate and receive access tokens.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | User's email address |
| `password` | string | Yes | User's password |

**Response:** `200 OK`
```json
{
  "token": "token_value",
  "csrf_token": "csrf_token_value",
  "user": {
    "id": 1,
    "name": "John Doe",
    "created_at": "2024-01-15T10:30:00Z",
    "modified_at": "2024-01-15T10:30:00Z"
  }
}
```

**Cookies Set:** Same as register endpoint

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields
- `401 Unauthorized` - Invalid email or password

### Refresh Token

Refresh access token using a valid refresh token.

**Endpoint:** `POST /auth/refresh`

**Request:** No body required. Refresh token should be provided via cookie or `Authorization` header.

**Response:** `200 OK`
```json
{
  "token": "token_value",
  "csrf_token": "csrf_token_value"
}
```

**Cookies Set:** New tokens with updated expiration times

**Error Responses:**
- `401 Unauthorized` - Refresh token missing, invalid, expired, or revoked

### Get Current User

Get information about the currently authenticated user.

**Endpoint:** `GET /auth/me`

**Authentication:** Required (access token)

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
- `401 Unauthorized` - Authentication required

### Logout

Logout and revoke refresh token.

**Endpoint:** `POST /auth/logout`

**Authentication:** Required (access token + CSRF token)

**Request:** No body required. Refresh token should be provided via cookie or `Authorization` header.

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

**Cookies Cleared:** All authentication cookies are cleared

**Error Responses:**
- `401 Unauthorized` - Authentication required

### Get CSRF Token

Get a CSRF token for state-changing operations.

**Endpoint:** `GET /auth/csrf-token`

**Authentication:** Required (access token)

**Response:** `200 OK`
```json
{
  "csrf_token": "csrf_token_value"
}
```

**Cookies Set:** CSRF token cookie (24 hours)

**Error Responses:**
- `401 Unauthorized` - Authentication required

### Authentication Headers

For API clients that prefer header-based authentication:

1. **Access Token:** Include in `Authorization` header:
   ```
   Authorization: Bearer <access_token>
   ```

2. **CSRF Token:** Include in `X-CSRF-Token` header for POST, PUT, PATCH, DELETE requests:
   ```
   X-CSRF-Token: <csrf_token>
   ```

3. **Refresh Token:** Include in `Authorization` header when refreshing:
   ```
   Authorization: Bearer <refresh_token>
   ```

**Note:** Cookie-based authentication is also supported and is the recommended approach for web applications.

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

### Get User Balances

Retrieve comprehensive balance information for a specific user across all groups.

**Endpoint:** `GET /users/{user_id}/balances`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | integer | Yes | User ID |

**Response:** `200 OK`
```json
{
  "user_id": 1,
  "summary": {
    "net_balance": "125.50",
    "total_owed": "75.25",
    "total_owed_to_user": "200.75"
  },
  "balances_by_group": [
    {
      "group_id": 1,
      "group_name": "Roommates",
      "net_balance": "125.50",
      "total_owed": "0.00",
      "total_owed_to_user": "125.50"
    }
  ],
  "balances_by_member": [
    {
      "member_id": 2,
      "member_name": "Jane Smith",
      "net_balance": "-62.75",
      "total_owed": "62.75",
      "total_owed_to_user": "0.00"
    }
  ],
  "group_count": 1,
  "member_count": 1
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `user_id` | integer | The user ID |
| `summary` | object | Overall balance summary across all groups |
| `summary.net_balance` | string (decimal) | Net balance (positive = owed to user, negative = user owes) |
| `summary.total_owed` | string (decimal) | Total amount user owes to others |
| `summary.total_owed_to_user` | string (decimal) | Total amount others owe to user |
| `balances_by_group` | array | Breakdown by group |
| `balances_by_member` | array | Breakdown by individual members |
| `group_count` | integer | Number of groups with balances |
| `member_count` | integer | Number of members with balances |

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

### Batch Group Member Operations

Batch operations allow you to manage multiple group members atomically.

#### Create Group Members (Batch)

Add multiple users to a group in a single atomic operation.

**Endpoint:** `POST /groups/{group_id}/members/batch`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Request Body:**
```json
{
  "members": [
    {
      "user_id": 1
    },
    {
      "user_id": 2
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `members` | array | Yes | Array of member objects |
| `members[].user_id` | integer | No | User ID (nullable) |

**Response:** `201 Created`
```json
{
  "group": {
    "id": 1,
    "name": "Roommates"
  },
  "group_members": [
    {
      "id": 1,
      "group_id": 1,
      "member_name": "John Doe",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "group_id": 1,
      "member_name": "Jane Smith",
      "user_id": 2,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 2
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields
- `400 Bad Request` - At least one member is required

#### Update Group Members (Batch)

Replace all members of a group with a new set. This operation atomically deletes existing members and creates new ones.

**Endpoint:** `PUT /groups/{group_id}/members/batch` or `PATCH /groups/{group_id}/members/batch`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Request Body:**
```json
{
  "members": [
    {
      "user_id": 2
    },
    {
      "user_id": 3
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "group": {
    "id": 1,
    "name": "Roommates"
  },
  "deleted_members": [
    {
      "id": 1,
      "group_id": 1,
      "member_name": "John Doe",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "new_members": [
    {
      "id": 2,
      "group_id": 1,
      "member_name": "Jane Smith",
      "user_id": 2,
      "created_at": "2024-01-15T12:00:00Z"
    },
    {
      "id": 3,
      "group_id": 1,
      "member_name": "Bob Johnson",
      "user_id": 3,
      "created_at": "2024-01-15T12:00:00Z"
    }
  ],
  "deleted_count": 1,
  "new_count": 2
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields
- `400 Bad Request` - At least one member is required

#### Delete All Group Members (Batch)

Remove all members from a group in a single atomic operation.

**Endpoint:** `DELETE /groups/{group_id}/members/batch`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group_id` | integer | Yes | Group ID |

**Response:** `200 OK`
```json
{
  "group_id": 1,
  "message": "All group members deleted successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid group ID format

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
| `start_date` | string | No | 1 year ago | Start date in YYYY-MM-DD format |
| `end_date` | string | No | today | End date in YYYY-MM-DD format |
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
- `400 Bad Request` - Invalid group ID or invalid date format
- `403 Forbidden` - User is not a member of this group

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
| `by_user` | integer | Yes | Group Member ID who created the transaction (not User ID) |

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
| `start_date` | string | No | 1 year ago | Start date in YYYY-MM-DD format |
| `end_date` | string | No | today | End date in YYYY-MM-DD format |
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
- `400 Bad Request` - Invalid group ID or invalid date format
- `403 Forbidden` - User is not a member of this group

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
| `by_user` | integer | Yes | Group Member ID who created the transaction (not User ID) |

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
| `by_user` | integer | Yes | Group Member ID who created the transaction (not User ID) |

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

### Batch Split Operations

Batch operations ensure that splits always add up to 100% of the transaction amount. These are the recommended endpoints for managing splits.

#### Create Splits for Transaction (Batch)

Create multiple splits for a transaction atomically, ensuring they add up to 100%.

**Endpoint:** `POST /splits/transaction/{transaction_id}/batch`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |

**Request Body:**
```json
{
  "splits": [
    {
      "split_percent": 0.50,
      "split_amount": 62.75,
      "split_user": 1
    },
    {
      "split_percent": 0.50,
      "split_amount": 62.75,
      "split_user": 2
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `splits` | array | Yes | Array of split objects |
| `splits[].split_percent` | decimal | Yes | Percentage of transaction amount (0.0 to 1.0) |
| `splits[].split_amount` | decimal | Yes | Amount assigned to this split |
| `splits[].split_user` | integer | No | Group Member ID responsible for this split (not User ID, nullable) |

**Validation:**
- ✅ All split percentages must sum to exactly 1.0 (100%)
- ✅ All split amounts must sum to transaction amount (within 1 cent tolerance)
- ✅ At least one split is required
- ✅ Transaction must exist

**Response:** `201 Created`
```json
{
  "splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "0.500000",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "0.500000",
      "split_amount": "62.75",
      "split_user": 2,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "message": "Successfully created 2 splits"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields
- `400 Bad Request` - Split percentages must add up to 100%
- `400 Bad Request` - Split amounts must add up to transaction amount
- `400 Bad Request` - At least one split is required

#### Update All Splits for Transaction (Batch)

Atomically replace ALL existing splits with a new set. This operation ensures splits always add up to 100%.

**Endpoint:** `PUT /splits/transaction/{transaction_id}/batch` or `PATCH /splits/transaction/{transaction_id}/batch`

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `transaction_id` | integer | Yes | Transaction ID |

**Request Body:**
```json
{
  "splits": [
    {
      "split_percent": 0.333333,
      "split_amount": 33.33,
      "split_user": 1
    },
    {
      "split_percent": 0.333333,
      "split_amount": 33.33,
      "split_user": 2
    },
    {
      "split_percent": 0.333334,
      "split_amount": 33.34,
      "split_user": 3
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "deleted_splits": [
    {
      "id": 1,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "0.500000",
      "split_amount": "62.75",
      "split_user": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "modified_at": "2024-01-15T10:30:00Z"
    }
  ],
  "new_splits": [
    {
      "id": 2,
      "transaction_id": 1,
      "tx_amount": "125.50",
      "split_percent": "0.333333",
      "split_amount": "33.33",
      "split_user": 1,
      "created_at": "2024-01-15T12:00:00Z",
      "modified_at": "2024-01-15T12:00:00Z"
    }
  ],
  "message": "Successfully replaced 1 splits with 1 new splits"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid JSON or missing required fields
- `400 Bad Request` - Split percentages must add up to 100%
- `400 Bad Request` - Split amounts must add up to transaction amount
- `400 Bad Request` - At least one split is required

**Note:** See [SPLIT_API_GUIDE.md](Documentation/SPLIT_API_GUIDE.md) for detailed information on safe split management.

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

⚠️ **Warning:** Creating individual splits can leave transactions in an invalid state. Use batch operations (`POST /splits/transaction/{transaction_id}/batch`) instead.

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
| `split_user` | integer | No | Group Member ID responsible for this split (not User ID, nullable) |

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

⚠️ **Warning:** Creating individual splits can leave transactions in an invalid state. Use batch operations (`POST /splits/transaction/{transaction_id}/batch`) instead.

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
| `split_user` | integer | No | Group Member ID responsible for this split (not User ID, nullable) |

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

⚠️ **Warning:** Updating individual splits can leave transactions in an invalid state. Use batch operations (`PUT /splits/transaction/{transaction_id}/batch`) instead.

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
| `split_user` | integer | No | Group Member ID responsible for this split (not User ID, nullable) |

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

⚠️ **Warning:** Deleting individual splits can leave transactions in an invalid state. Use batch operations (`PUT /splits/transaction/{transaction_id}/batch`) to replace all splits instead.

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

**Note:** For safe split management, see [SPLIT_API_GUIDE.md](Documentation/SPLIT_API_GUIDE.md)

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

