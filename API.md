# API Documentation

## Base URL

```bash
http://localhost:8080/api/v1
```

### Base Response

```typescript
{
  success: boolean,
  message: string,
  data?: object ,
  errors?: []{
    field: string,
    message: string
  }
}
```

## Authentication

### Register User

**Endpoint:** `POST /auth/signup`

**Request Body:**

```json
{
  "email": "test@example.com",
  "password": "password123",
  "firstName": "John",
  "lastName": "Doe",
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Data:**

```json
{
  "account": {
    "id": "123xyz123xyz",
    "email":"test@example.com"
  },
  "profiles": [{
    "id": "123xyz123xyz",
    "firstName": "John",
    "lastName": "Doe"
    "avatarUrl": "https://example.com/avatar.jpg"
  }],
  "tokens": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

### Login

**Endpoint:** `POST /auth/signin`

**Request Body:**

```json
{
  "email": "test@example.com",
  "password": "password123",
    "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Data:**

```json
{
  "account": {
    "id": "123xyz123xyz",
    "email":"test@example.com"
  },
  "profiles": [{
    "id": "123xyz123xyz",
    "firstName": "John",
    "lastName": "Doe",
    "avatarUrl": "https://example.com/avatar.jpg"
  }],
  "tokens": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

## Authorization

### Logout

**Endpoint:** `POST /auth/logout`

**Headers:**

```
Authorization: Bearer YOUR_TOKEN_HERE
```

**Response:**

```json
{
  "success": true,
  "message": "Logout successful"
}
```

---

## Users

### Get User Profile (Public)

**Endpoint:** `GET /profiles/:id`

**Response Data:**

```json
{
  "id": "123xyz123xyz",
  "avatarUrl": "https://example.com/avatar.jpg",
  "displayName": "John Doe",
}
```

---

### Update Profile

**Endpoint:** `PUT /profiles/:id`

**Headers:**

```
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "profileId": "123xyz123xyz",
  "firstName": "John",
  "lastName": "Doe",
  "displayName": "John Doe",
  "avatar": "binary_file_data"
}
```

**Response Data:**

```json
{
    "id": "123xyz123xyz",
    "firstName": "John",
    "lastName": "Doe",
    "avatarUrl": "https://example.com/avatar.jpg"
}
```

---
