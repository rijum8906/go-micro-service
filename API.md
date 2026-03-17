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

### Base Authorization Type

```go
const (
  PassAuthzType = "password_authorization"
  MFAAuthzType  = "mfa_authorization"
)
```

### Base Scope Actions

```go
const (
 ActionChangePassword = "change_password"
 ActionChangeEmail    = "change_email"
 ActionChangeName     = "change_name"
 ActionChangePhone    = "change_phone"
 ActionChangeRole     = "change_role"
 ActionChangeStatus   = "change_status"
)
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

### Request Email Verification

**Endpoint:** `POST /auth/request-email-verification`

**Request Body:**

```json
{
  "email": "test@example.com",
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Verification email sent"
}
```

### Request Password Reset

**Endpoint:** `POST /auth/request-password-reset`

**Request Body:**

```json
{
  "email": "test@example.com",
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Successful message"
}
```

### verify email

**endpoint:** `post /auth/verify-email`

**request body:**

```json
{
  "token": "your-verification-token",
  "metadata": {
    "deviceid": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Verification email sent"
}
```

### Reset Password

**endpoint:** `post /auth/verify-email`

**request body:**

```json
{
  "token": "your-verification-token",
  "metadata": {
    "deviceid": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Successful message"
}
```

---

## Authorization

### Logout

**Endpoint:** `POST /auth/logout`

**Headers:**

```typescript
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

## Profiles

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

### Update Profile

**Endpoint:** `PUT /profiles/:id`

**Headers:**

```typescript
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

### Delete Profile

**Endpoint:** `DELETE /profiles/:id`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Profile deleted successfully"
}
```

---

## Accounts

### Change Email

**Endpoint:** `PUT /accounts/change-email`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "token": "your-verification-token",
  "newEmail": "test@example.com",
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Email changed successfully",
  "data": {
    "email": "test@example.com"
  }
}
```

### Change Password

**Endpoint:** `PUT /accounts/change-password`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "token": "your-verification-token",
  "newPassword": "test@example.com",
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Password changed successfully",
}
```

### Delete Account

**Endpoint:** `DELETE /accounts/:id`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "success": true,
  "message": "Account deleted successfully",
}
```

### My Account

**Endpoint:** `GET /accounts/me`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Response Body:**

```json
{
  "id": "123xyz123xyz",
  "email": "test@example.com",
  "isEmailVerified": true,
  "twoFactorEnabled": false,
  "createdAt": "2023-09-20T12:34:56Z",
  "updatedAt": "2023-09-20T12:34:56Z"
}
```

### Generate Scoped Token

**Endpoint:** `POST /accounts/generate-scoped-token`

**Headers:**

```typescript
Authorization: Bearer YOUR_TOKEN_HERE
```

**Request Body:**

```json
{
  "scope": "change_password", // base scope actions
  "authorization": {
    "type": "password_authorization", // base authorization type
    "value": "password"
  },
  "metadata": {
    "deviceId": "something_id_ksaal23ua"
  }
}
```

**Response Body:**

```json
{
  "id": "123xyz123xyz",
  "email": "test@example.com",
  "isEmailVerified": true,
  "twoFactorEnabled": false,
  "createdAt": "2023-09-20T12:34:56Z",
  "updatedAt": "2023-09-20T12:34:56Z"
}
```
