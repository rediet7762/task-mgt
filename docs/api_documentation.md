# Task Management API — Documentation

A Gin-based REST API for managing tasks, secured with JWT authentication and
role-based authorization. Data is stored in MongoDB.

## Setup

1. Create a `.env` file (already present in this repo) with:
   - `MONGODB_URI` — your MongoDB connection string (required)
   - `MONGODB_DB` — database name (default: `task_manager`)
   - `MONGODB_COLLECTION` — task collection name (default: `tasks`)
   - `JWT_SECRET` — secret used to sign JWTs (already generated for you; keep
     it private and rotate it if it's ever exposed)
2. Install dependencies and run (the entry point now lives in `Delivery/`,
   and `.env` must be readable from the working directory you run from):
   ```bash
   go mod tidy
   go run ./Delivery
   ```
   The server starts on `http://localhost:8080`. On first run it connects to
   MongoDB and seeds a few sample tasks if the `tasks` collection is empty.

See [`ARCHITECTURE.md`](../ARCHITECTURE.md) in the project root for how the
codebase is organized into layers.

## Roles

- **admin** — can create, update, and delete tasks; can promote other users
  to admin; can also do everything a regular user can do.
- **user** — can view all tasks and view a single task by ID.

The **first user ever registered** (i.e. the users collection is empty at
registration time) is automatically made an admin. Every user registered
after that starts out as a regular user.

## Authentication flow

1. `POST /api/v1/register` to create an account.
2. `POST /api/v1/login` with those credentials to receive a JWT.
3. Send that token on every subsequent request as:
   ```
   Authorization: Bearer <token>
   ```
   Tokens expire 24 hours after they're issued; log in again for a new one.

---

## Endpoints

### `POST /api/v1/register` — public

Create a new user account.

**Request body**
```json
{
  "username": "alice",
  "password": "a-strong-password"
}
```

**Response `201 Created`**
```json
{
  "data": {
    "id": "665f1b2e8f1b2c0012345678",
    "username": "alice",
    "role": "admin",
    "created_at": "2026-07-28T09:00:00Z"
  },
  "message": "User registered successfully"
}
```

**Errors**
- `400 Bad Request` — missing/invalid username or password
- `409 Conflict` — username already taken

---

### `POST /api/v1/login` — public

Authenticate and receive a JWT.

**Request body**
```json
{
  "username": "alice",
  "password": "a-strong-password"
}
```

**Response `200 OK`**
```json
{
  "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." },
  "message": "Login successful"
}
```

**Errors**
- `400 Bad Request` — missing username/password
- `401 Unauthorized` — invalid username or password

---

### `POST /api/v1/promote/:username` — admin only

Promote an existing user to admin.

**Headers**
```
Authorization: Bearer <admin token>
```

**Response `200 OK`**
```json
{ "message": "bob has been promoted to admin" }
```

**Errors**
- `401 Unauthorized` — missing/invalid token
- `403 Forbidden` — caller is not an admin
- `404 Not Found` — no such username

---

### `GET /api/v1/tasks/` — authenticated (any role)

Returns every task. Supports an optional `status` query parameter
(`pending`, `in-progress`, `completed`) to filter results.

**Headers**
```
Authorization: Bearer <token>
```

**Response `200 OK`**
```json
{
  "data": [
    {
      "id": "665f1c3e8f1b2c0012345679",
      "title": "Write docs",
      "description": "Document the auth endpoints",
      "due_date": "2026-08-01",
      "status": "in-progress",
      "created_at": "2026-07-28T09:00:00Z",
      "updated_at": "2026-07-28T09:00:00Z"
    }
  ],
  "count": 1
}
```

---

### `GET /api/v1/tasks/:id` — authenticated (any role)

Returns a single task by ID.

**Headers**
```
Authorization: Bearer <token>
```

**Errors**
- `400 Bad Request` — malformed task ID
- `401 Unauthorized` — missing/invalid token
- `404 Not Found` — no task with that ID

---

### `POST /api/v1/tasks/` — admin only

Create a new task.

**Headers**
```
Authorization: Bearer <admin token>
```

**Request body**
```json
{
  "title": "Write docs",
  "description": "Document the auth endpoints",
  "due_date": "2026-08-01",
  "status": "pending"
}
```
`status` must be one of `pending`, `in-progress`, `completed`.

**Response `201 Created`** — the created task, including its generated `id`.

**Errors**
- `400 Bad Request` — invalid body
- `401 Unauthorized` — missing/invalid token
- `403 Forbidden` — caller is not an admin

---

### `PUT /api/v1/tasks/:id` — admin only

Partially updates an existing task — only the fields present in the body are
applied.

**Errors**
- `400 Bad Request` — invalid body or task ID
- `401 Unauthorized` — missing/invalid token
- `403 Forbidden` — caller is not an admin
- `404 Not Found` — no task with that ID

---

### `DELETE /api/v1/tasks/:id` — admin only

**Response `200 OK`**
```json
{ "message": "Task deleted successfully" }
```

**Errors**
- `400 Bad Request` — malformed task ID
- `401 Unauthorized` — missing/invalid token
- `403 Forbidden` — caller is not an admin
- `404 Not Found` — no task with that ID

---

## Security notes

- Passwords are hashed with **bcrypt** before being stored; the plain-text
  password is never persisted or logged, and the hash is never returned in
  API responses (`User.Password` is tagged `json:"-"`).
- Tokens are signed with HMAC-SHA256 (`HS256`) using `JWT_SECRET` and carry
  the user ID, username, role, and an expiration claim.
- All endpoints other than `/register` and `/login` require a valid,
  unexpired token; write and promotion endpoints additionally require the
  `admin` role.

## Manual testing checklist (Postman or curl)

1. Register the first user → confirm `role` is `admin`.
2. Register a second user → confirm `role` is `user`.
3. Log in as each → confirm a token is returned.
4. Call `GET /api/v1/tasks/` with no `Authorization` header → expect `401`.
5. Call `GET /api/v1/tasks/` as the regular user → expect `200`.
6. Call `POST /api/v1/tasks/` as the regular user → expect `403`.
7. Call `POST /api/v1/tasks/` as the admin → expect `201`.
8. Call `POST /api/v1/promote/<regular-username>` as the admin → expect `200`.
9. Log in again as the promoted user and confirm they can now create tasks.
