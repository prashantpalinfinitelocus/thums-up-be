# Thums Up Backend - High-Level Design (HLD) & Low-Level Design (LLD)

## Overview

The Thums Up Backend is an enterprise-grade Go application that powers a gamified engagement platform with Thunder Seat contests. Users can authenticate via OTP, manage profiles and addresses, submit answers to weekly questions, and participate in Thunder Seat contests to win prizes. The system integrates with Infobip for SMS/WhatsApp notifications, Google Cloud Storage (GCS) for file management, and uses PostgreSQL for data persistence. The architecture follows Clean Architecture principles with proper separation of concerns, transaction management, and graceful shutdown mechanisms.

---

## System Architecture

### Components

#### Backend Services

| Component | Responsibility |
|-----------|---------------|
| **Server (Gin)** | HTTP server initialization, middleware chain, routing, graceful shutdown with SIGTERM/SIGINT handling. Runs on configurable port (default: 8080). |
| **Handlers** | HTTP request/response handling, JSON binding/validation, auth context extraction, error envelope wrapping. |
| **Services** | Business logic orchestration, transaction management, validation rules, cross-entity operations. |
| **Repositories** | Database CRUD operations, query building, GORM entity mapping, transaction support. |
| **Vendors/Clients** | External service integrations: Infobip (OTP/SMS), GCS (file storage), Firebase (push notifications). |
| **Middlewares** | CORS, Auth (JWT), API Key validation, error handling, logging, panic recovery. |
| **Worker Pool** | Background async task processing with configurable pool size (default: 10 workers, queue size: 100). |
| **Database (PostgreSQL)** | Relational data store with GORM ORM, connection pooling, auto-migrations, soft deletes. |
| **Storage (GCS)** | Cloud storage for user-uploaded assets with signed URL generation, bucket: configurable via env. |

#### Key Integrations

- **Infobip**: OTP delivery via SMS/WhatsApp, configurable base URL + API key
- **Firebase**: Push notifications for contest updates and winner announcements
- **Google Cloud Storage**: File uploads/downloads with service account authentication
- **PostgreSQL**: Primary data store with ACID transactions

---

## Database Schema

### users

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | UUID | NO | PRI | uuid_generate_v4() | Primary Key |
| phone_number | VARCHAR(15) | NO | UNI | | Unique, indexed |
| name | VARCHAR(255) | YES | | | User display name |
| email | VARCHAR(255) | YES | UNI | | Unique, optional |
| is_active | BOOLEAN | NO | | true | Soft delete flag |
| is_verified | BOOLEAN | NO | | false | Phone verified flag |
| referral_code | VARCHAR(20) | YES | UNI | | System-generated |
| referred_by | VARCHAR(20) | YES | | | Referrer's code |
| device_token | TEXT | YES | | | FCM token |
| created_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | Auto-update |
| deleted_at | TIMESTAMP | YES | IDX | | GORM soft delete |

### otp_logs

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| phone_number | VARCHAR(15) | NO | IDX | | Indexed for lookup |
| otp | VARCHAR(6) | NO | | | 6-digit code |
| expires_at | TIMESTAMP | NO | | | OTP expiry time |
| is_verified | BOOLEAN | NO | | false | Verification flag |
| verified_at | TIMESTAMP | YES | | | Timestamp of verification |
| attempts | INTEGER | NO | | 0 | Failed attempts counter |
| created_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |

### refresh_tokens

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | UUID | NO | PRI | uuid_generate_v4() | Primary Key |
| user_id | UUID | NO | IDX | | FK to users |
| token | TEXT | NO | UNI | | JWT refresh token |
| expires_at | TIMESTAMP | NO | | | Token expiry (default: 30 days) |
| is_revoked | BOOLEAN | NO | | false | Revocation flag |
| created_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |

**Constraint**: FK users(id) ON DELETE CASCADE

### address

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| user_id | UUID | NO | IDX | | FK to users |
| address1 | VARCHAR(500) | NO | | | Primary address line |
| address2 | VARCHAR(500) | YES | | | Secondary address line |
| pincode | INTEGER | NO | | | Postal code |
| pin_code_id | INTEGER | NO | | | FK to pin_code |
| city_id | INTEGER | NO | | | FK to city |
| state_id | INTEGER | NO | | | FK to state |
| nearest_landmark | VARCHAR(255) | YES | | | Landmark reference |
| shipping_mobile | VARCHAR(15) | YES | | | Alternate contact |
| is_default | BOOLEAN | NO | | false | Primary address flag |
| is_active | BOOLEAN | NO | | true | Active status |
| is_deleted | BOOLEAN | NO | | false | Soft delete |
| created_by | UUID | NO | | | User ID |
| created_on | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |
| last_modified_by | UUID | YES | | | User ID |
| last_modified_on | TIMESTAMP | YES | | | |

### state

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| name | VARCHAR(255) | NO | | | State name |
| is_active | BOOLEAN | NO | | true | |
| is_deleted | BOOLEAN | NO | | false | |

### city

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| state_id | INTEGER | NO | IDX | | FK to state |
| name | VARCHAR(255) | NO | | | City name |
| is_active | BOOLEAN | NO | | true | |
| is_deleted | BOOLEAN | NO | | false | |

### pin_code

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| city_id | INTEGER | NO | IDX | | FK to city |
| pincode | INTEGER | NO | | | 6-digit pincode |
| is_deliverable | BOOLEAN | NO | | false | Delivery availability |
| is_active | BOOLEAN | NO | | true | |
| is_deleted | BOOLEAN | NO | | false | |

### question_master

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| question_text | TEXT | NO | | | Question content |
| language_id | INTEGER | NO | | | Language reference |
| is_active | BOOLEAN | NO | | true | Active for current week |
| is_deleted | BOOLEAN | NO | | false | |
| created_by | VARCHAR | NO | | | Admin user ID |
| created_on | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |
| last_modified_by | VARCHAR | YES | | | |
| last_modified_on | TIMESTAMP | YES | | | |

### thunder_seat

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| user_id | UUID | NO | IDX | | FK to users |
| question_id | INTEGER | NO | | | FK to question_master |
| week_number | INTEGER | NO | IDX | | Contest week |
| answer | TEXT | NO | | | User's answer |
| created_by | UUID | NO | | | User ID |
| created_on | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |

**Unique Constraint**: (user_id, question_id, week_number) - One submission per user per question per week

### thunder_seat_winner

| Field | Type | Null | Key | Default | Extra |
|-------|------|------|-----|---------|-------|
| id | SERIAL | NO | PRI | | Auto-increment |
| user_id | UUID | NO | IDX | | FK to users |
| thunder_seat_id | INTEGER | NO | | | FK to thunder_seat |
| week_number | INTEGER | NO | IDX | | Contest week |
| created_by | UUID | NO | | | Admin/system ID |
| created_on | TIMESTAMP | NO | | CURRENT_TIMESTAMP | |

---

## Component Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                            │
│  (Mobile Apps, Web App, Admin Dashboard)                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GIN HTTP SERVER                            │
│  - CORS Middleware                                               │
│  - Logger Middleware                                             │
│  - Recovery Middleware                                           │
│  - Error Handler Middleware                                      │
│  - Auth Middleware (JWT)                                         │
│  - API Key Middleware (Admin)                                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        HANDLER LAYER                            │
│  - AuthHandler                                                   │
│  - ProfileHandler                                                │
│  - AddressHandler                                                │
│  - QuestionHandler                                               │
│  - ThunderSeatHandler                                            │
│  - WinnerHandler                                                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        SERVICE LAYER                            │
│  - AuthService (OTP, Signup, JWT)                               │
│  - UserService (Profile, Address)                               │
│  - QuestionService (Question CRUD)                              │
│  - ThunderSeatService (Submissions, Winners)                     │
│  - WinnerService (Selection, Retrieval)                         │
│  - NotificationService (Push/SMS)                                │
└────────────────────────────┬────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ▼                    ▼                    ▼
┌──────────────┐   ┌──────────────────┐   ┌──────────────┐
│ REPOSITORY   │   │ VENDOR CLIENTS   │   │ WORKER POOL  │
│ LAYER        │   │                  │   │              │
│              │   │ - InfobipClient  │   │ Async Tasks  │
│ - UserRepo   │   │ - FirebaseClient │   │ (10 workers) │
│ - OTPRepo    │   │ - GCS Service    │   │              │
│ - TokenRepo  │   │                  │   │              │
│ - AddressRepo│   └──────────────────┘   └──────────────┘
│ - QuestionRepo│
│ - ThunderRepo │
│ - WinnerRepo  │
└────────┬─────┘
         ▼
┌──────────────────┐
│   PostgreSQL     │
│   Database       │
└──────────────────┘
```

---

## API Endpoints

### Authentication & Authorization

#### 1. Send OTP

**Endpoint**: `[POST] /api/v1/auth/send-otp`

**Description**: Generate and send 6-digit OTP to user's phone via Infobip SMS/WhatsApp

**Authentication**: None (Public)

**Request Body**:
```json
{
  "phone_number": "9876543210"
}
```

**Validation**:
- `phone_number`: required, exactly 10 digits, numeric only

**Response** (200 OK):
```json
{
  "success": true,
  "message": "OTP sent successfully",
  "data": null
}
```

**Response** (400 Bad Request):
```json
{
  "success": false,
  "error": "Validation failed",
  "details": {
    "phone_number": "Phone number must be exactly 10 digits"
  }
}
```

**Response** (429 Too Many Requests):
```json
{
  "success": false,
  "error": "Too many OTP requests. Please try after 60 seconds"
}
```

**Business Logic**:
1. Validate phone number format (10 digits, numeric)
2. Check rate limiting (max 3 requests per phone per hour)
3. Generate 6-digit random OTP
4. Set expiry to 5 minutes from now
5. Store in `otp_logs` table
6. Call Infobip API to send SMS/WhatsApp
7. Return success response

---

#### 2. Verify OTP

**Endpoint**: `[POST] /api/v1/auth/verify-otp`

**Description**: Verify OTP and return temporary access token (for existing users only)

**Authentication**: None (Public)

**Request Body**:
```json
{
  "phone_number": "9876543210",
  "otp": "123456"
}
```

**Validation**:
- `phone_number`: required, exactly 10 digits
- `otp`: required, exactly 6 digits

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "phone_number": "9876543210",
    "name": "John Doe"
  }
}
```

**Response** (400 Bad Request - New User)**:
```json
{
  "success": false,
  "error": "User not found. Please complete signup"
}
```

**Response** (401 Unauthorized - Invalid OTP)**:
```json
{
  "success": false,
  "error": "Invalid or expired OTP"
}
```

**Business Logic**:
1. Validate request payload
2. Find latest non-verified OTP for phone number
3. Check OTP expiry (must be within 5 minutes)
4. Verify OTP matches (increment attempts counter on failure)
5. Check max attempts (3 attempts allowed)
6. Find user by phone number
7. If user not found, return 400 with signup message
8. Mark OTP as verified with timestamp
9. Generate JWT access token (expires in 1 hour)
10. Generate refresh token (expires in 30 days)
11. Store refresh token in database
12. Return tokens with user info

---

#### 3. Sign Up

**Endpoint**: `[POST] /api/v1/auth/signup`

**Description**: Register new user and return JWT tokens

**Authentication**: None (Public)

**Request Body**:
```json
{
  "phone_number": "9876543210",
  "name": "John Doe",
  "email": "john@example.com",
  "referral_code": "REF123",
  "device_token": "fcm_token_here"
}
```

**Validation**:
- `phone_number`: required, exactly 10 digits
- `name`: required, non-empty
- `email`: optional, valid email format if provided
- `referral_code`: optional, must exist in database if provided
- `device_token`: optional, FCM token

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "phone_number": "9876543210",
    "name": "John Doe"
  }
}
```

**Response** (400 Bad Request - Duplicate)**:
```json
{
  "success": false,
  "error": "User already exists with this phone number"
}
```

**Response** (400 Bad Request - Invalid Referral)**:
```json
{
  "success": false,
  "error": "Invalid referral code"
}
```

**Business Logic**:
1. Validate request payload
2. Check if user already exists with phone number
3. Validate referral code if provided (must exist and be active)
4. Generate unique referral code for new user (8 chars alphanumeric)
5. Create user record with provided details
6. Set `is_verified` = false, `is_active` = true
7. Generate JWT access token (expires in 1 hour)
8. Generate refresh token (expires in 30 days)
9. Store refresh token in database
10. Return tokens with user info

---

#### 4. Refresh Token

**Endpoint**: `[POST] /api/v1/auth/refresh`

**Description**: Refresh access token using refresh token

**Authentication**: None (Public)

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "phone_number": "9876543210",
    "name": "John Doe"
  }
}
```

**Response** (401 Unauthorized)**:
```json
{
  "success": false,
  "error": "Invalid or expired refresh token"
}
```

**Business Logic**:
1. Validate refresh token JWT signature
2. Find refresh token in database
3. Check if token is revoked
4. Check if token is expired
5. Get user from database
6. Generate new JWT access token
7. Generate new refresh token
8. Revoke old refresh token
9. Store new refresh token
10. Return new tokens

---

### Profile Management

#### 5. Get User Profile

**Endpoint**: `[GET] /api/v1/profile`

**Description**: Retrieve authenticated user's profile

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "phone_number": "9876543210",
    "name": "John Doe",
    "email": "john@example.com",
    "is_active": true,
    "is_verified": true,
    "referral_code": "ABC12345",
    "referred_by": "REF123",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T15:45:00Z"
  }
}
```

**Response** (401 Unauthorized):
```json
{
  "error": "User not authenticated"
}
```

**Business Logic**:
1. Extract user from auth context (set by middleware)
2. Fetch user details from database
3. Return user profile

---

#### 6. Update User Profile

**Endpoint**: `[PATCH] /api/v1/profile`

**Description**: Update user's name and/or email

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "name": "John Updated",
  "email": "johnupdated@example.com"
}
```

**Validation**:
- `name`: optional, if provided must be non-empty
- `email`: optional, if provided must be valid email format

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "phone_number": "9876543210",
  "name": "John Updated",
  "email": "johnupdated@example.com",
  "is_active": true,
  "is_verified": true,
  "referral_code": "ABC12345",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-25T12:00:00Z"
}
```

**Response** (400 Bad Request - Email In Use)**:
```json
{
  "error": "Email already in use by another account"
}
```

**Business Logic**:
1. Extract user ID from auth context
2. Validate request payload
3. Check if email is already in use by another user
4. Update user record with new values
5. Return updated profile

---

### Address Management

#### 7. Get User Addresses

**Endpoint**: `[GET] /api/v1/profile/address`

**Description**: Get all active addresses for authenticated user

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "address1": "123 Main Street",
    "address2": "Apartment 4B",
    "pincode": 400001,
    "state": "Maharashtra",
    "city": "Mumbai",
    "nearest_landmark": "Near City Mall",
    "shipping_mobile": "9876543210",
    "is_default": true,
    "is_active": true,
    "created_on": "2024-01-15T10:30:00Z",
    "last_modified_on": "2024-01-20T15:45:00Z"
  }
]
```

**Business Logic**:
1. Extract user ID from auth context
2. Query addresses for user where `is_active=true` and `is_deleted=false`
3. Join with state, city tables to get names
4. Return address list

---

#### 8. Add Address

**Endpoint**: `[POST] /api/v1/profile/address`

**Description**: Add new address for authenticated user

**Authentication**: Bearer Token (JWT)

**Request Body**:
```json
{
  "address1": "123 Main Street",
  "address2": "Apartment 4B",
  "pincode": 400001,
  "state": "Maharashtra",
  "city": "Mumbai",
  "nearest_landmark": "Near City Mall",
  "shipping_mobile": "9876543210",
  "is_default": false
}
```

**Validation**:
- `address1`: required, max 500 chars
- `address2`: optional, max 500 chars
- `pincode`: required, 6 digits
- `state`: required, must exist in state table
- `city`: required, must exist in city table and belong to state
- `nearest_landmark`: optional, max 255 chars
- `shipping_mobile`: optional, 10 digits
- `is_default`: optional, boolean

**Response** (201 Created):
```json
{
  "id": 1,
  "address1": "123 Main Street",
  "address2": "Apartment 4B",
  "pincode": 400001,
  "state": "Maharashtra",
  "city": "Mumbai",
  "nearest_landmark": "Near City Mall",
  "shipping_mobile": "9876543210",
  "is_default": false,
  "is_active": true,
  "created_on": "2024-01-25T12:00:00Z"
}
```

**Response** (400 Bad Request - Invalid Location)**:
```json
{
  "error": "Invalid state: Maharashtra not found"
}
```

**Response** (400 Bad Request - Not Deliverable)**:
```json
{
  "error": "Pincode 999999 is not deliverable in this area"
}
```

**Business Logic**:
1. Extract user ID from auth context
2. Validate state exists and is active
3. Validate city exists, belongs to state, and is active
4. Validate pincode exists, belongs to city, and is active
5. Check pincode deliverability (`is_deliverable=true`)
6. If `is_default=true`, unset default flag on other user addresses
7. Create address record with user_id and location IDs
8. Return created address

---

#### 9. Update Address

**Endpoint**: `[PUT] /api/v1/profile/address/:addressId`

**Description**: Update existing address

**Authentication**: Bearer Token (JWT)

**Path Parameters**:
- `addressId`: integer, address ID to update

**Request Body**: Same as Add Address

**Response** (200 OK): Same structure as Add Address

**Response** (403 Forbidden)**:
```json
{
  "error": "Address does not belong to user"
}
```

**Response** (404 Not Found)**:
```json
{
  "error": "Address not found"
}
```

**Business Logic**:
1. Extract user ID from auth context
2. Validate address exists and is not deleted
3. Verify address belongs to authenticated user
4. Validate new location data (same as Add Address)
5. Update address record
6. Return updated address

---

#### 10. Delete Address

**Endpoint**: `[DELETE] /api/v1/profile/address/:addressId`

**Description**: Soft delete address (sets `is_deleted=true`)

**Authentication**: Bearer Token (JWT)

**Path Parameters**:
- `addressId`: integer, address ID to delete

**Response** (200 OK):
```json
{
  "message": "Address deleted successfully"
}
```

**Response** (403 Forbidden)**:
```json
{
  "error": "Address does not belong to user"
}
```

**Business Logic**:
1. Extract user ID from auth context
2. Validate address exists and is not already deleted
3. Verify address belongs to authenticated user
4. Set `is_deleted=true` on address record
5. Return success message

---

### Question Management

#### 11. Get Active Questions

**Endpoint**: `[GET] /api/v1/questions/active`

**Description**: Get all active questions for current week

**Authentication**: None (Public)

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "question_text": "What makes Thums Up unique?",
      "language_id": 1,
      "is_active": true
    },
    {
      "id": 2,
      "question_text": "थम्स अप को अनोखा क्या बनाता है?",
      "language_id": 2,
      "is_active": true
    }
  ]
}
```

**Business Logic**:
1. Query `question_master` where `is_active=true` and `is_deleted=false`
2. Order by `created_on` DESC
3. Return question list

---

#### 12. Submit Question (Admin)

**Endpoint**: `[POST] /api/v1/questions`

**Description**: Submit new question for Thunder Seat contest (Admin/Internal use)

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "question_text": "What is your favorite Thums Up moment?",
  "language_id": 1
}
```

**Validation**:
- `question_text`: required, non-empty
- `language_id`: required, must be valid language ID

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 3,
    "question_text": "What is your favorite Thums Up moment?",
    "language_id": 1,
    "is_active": true
  },
  "message": "Question submitted successfully"
}
```

**Business Logic**:
1. Extract user ID from auth context
2. Validate request payload
3. Validate language_id exists
4. Create question record with `created_by` = user_id
5. Set `is_active=true`, `is_deleted=false`
6. Return created question

---

### Thunder Seat Contest

#### 13. Get Current Week

**Endpoint**: `[GET] /api/v1/thunder-seat/current-week`

**Description**: Get current Thunder Seat contest week information

**Authentication**: None (Public)

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "week_number": 3,
    "start_date": "2024-01-15T00:00:00Z",
    "end_date": "2024-01-21T23:59:59Z"
  }
}
```

**Business Logic**:
1. Calculate week number based on campaign start date (stored in constants)
2. Calculate start_date as Monday 00:00:00 of current week
3. Calculate end_date as Sunday 23:59:59 of current week
4. Return week information

---

#### 14. Submit Thunder Seat Answer

**Endpoint**: `[POST] /api/v1/thunder-seat`

**Description**: Submit answer to Thunder Seat question

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "week_number": 3,
  "question_id": 1,
  "answer": "Thums Up's unique toofani taste sets it apart!"
}
```

**Validation**:
- `week_number`: required, must be current or past week
- `question_id`: required, question must exist and be active
- `answer`: required, non-empty, max 1000 chars

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 100,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "question_id": 1,
    "week_number": 3,
    "answer": "Thums Up's unique toofani taste sets it apart!",
    "created_on": "2024-01-18T14:30:00Z"
  },
  "message": "Answer submitted successfully"
}
```

**Response** (400 Bad Request - Duplicate)**:
```json
{
  "success": false,
  "error": "You have already submitted an answer for this question this week"
}
```

**Response** (400 Bad Request - Invalid Week)**:
```json
{
  "success": false,
  "error": "Week number is invalid or contest has not started yet"
}
```

**Business Logic**:
1. Extract user from auth context
2. Validate request payload
3. Validate question exists and is active
4. Validate week_number is current or past week
5. Check if user already submitted for this question + week (unique constraint)
6. Create thunder_seat record with user_id, question_id, week_number, answer
7. Return submission details

---

#### 15. Get User Submissions

**Endpoint**: `[GET] /api/v1/thunder-seat/submissions`

**Description**: Get authenticated user's Thunder Seat submissions

**Authentication**: Bearer Token (JWT)

**Headers**:
```
Authorization: Bearer <access_token>
```

**Query Parameters** (Optional):
- `week_number`: filter by specific week

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 100,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "question_id": 1,
      "week_number": 3,
      "answer": "Thums Up's unique toofani taste sets it apart!",
      "created_on": "2024-01-18T14:30:00Z"
    }
  ]
}
```

**Business Logic**:
1. Extract user_id from auth context
2. Query thunder_seat table for user submissions
3. If week_number provided, filter by week
4. Order by created_on DESC
5. Return submission list

---

### Winner Management

#### 16. Get All Winners (Paginated)

**Endpoint**: `[GET] /api/v1/winners`

**Description**: Get paginated list of all Thunder Seat winners

**Authentication**: None (Public)

**Query Parameters**:
- `limit`: required, integer (1-100), number of results per page
- `offset`: optional, integer (default: 0), pagination offset

**Example**: `/api/v1/winners?limit=20&offset=0`

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "thunder_seat_id": 100,
      "week_number": 2,
      "created_on": "2024-01-14T18:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_pages": 5,
    "total_count": 95
  }
}
```

**Business Logic**:
1. Validate limit (1-100) and offset (>= 0)
2. Query thunder_seat_winner with pagination
3. Count total winners for pagination metadata
4. Calculate total_pages = ceil(total_count / limit)
5. Calculate current_page = (offset / limit) + 1
6. Return winners with pagination metadata

---

#### 17. Get Winners by Week

**Endpoint**: `[GET] /api/v1/winners/week/:weekNumber`

**Description**: Get all winners for specific week

**Authentication**: None (Public)

**Path Parameters**:
- `weekNumber`: integer, week number to fetch winners for

**Example**: `/api/v1/winners/week/3`

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "thunder_seat_id": 150,
      "week_number": 3,
      "created_on": "2024-01-21T18:00:00Z"
    },
    {
      "id": 6,
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "thunder_seat_id": 151,
      "week_number": 3,
      "created_on": "2024-01-21T18:00:00Z"
    }
  ]
}
```

**Business Logic**:
1. Validate weekNumber is positive integer
2. Query thunder_seat_winner where week_number = weekNumber
3. Order by created_on ASC
4. Return winner list

---

### Admin Endpoints

#### 18. Select Winners (Admin)

**Endpoint**: `[POST] /api/v1/admin/winners/select`

**Description**: Admin endpoint to select random winners for a specific week

**Authentication**: X-API-Key Header

**Headers**:
```
X-API-Key: <admin_api_key>
```

**Request Body**:
```json
{
  "week_number": 3,
  "number_of_winners": 10
}
```

**Validation**:
- `week_number`: required, must be valid past week (not current or future)
- `number_of_winners`: required, minimum 1, cannot exceed total submissions

**Response** (201 Created):
```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "thunder_seat_id": 150,
      "week_number": 3,
      "created_on": "2024-01-21T18:00:00Z"
    },
    {
      "id": 6,
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "thunder_seat_id": 151,
      "week_number": 3,
      "created_on": "2024-01-21T18:00:00Z"
    }
  ],
  "message": "Winners selected successfully"
}
```

**Response** (400 Bad Request - Already Selected)**:
```json
{
  "success": false,
  "error": "Winners already selected for week 3"
}
```

**Response** (400 Bad Request - No Submissions)**:
```json
{
  "success": false,
  "error": "No submissions found for week 3"
}
```

**Response** (401 Unauthorized)**:
```json
{
  "success": false,
  "error": "Invalid or missing API key"
}
```

**Business Logic**:
1. Validate X-API-Key header
2. Validate request payload
3. Check if week_number is past week (not current/future)
4. Check if winners already selected for this week
5. Get all submissions for week_number
6. If submissions < number_of_winners, adjust to available count
7. Randomly select N submissions
8. Create thunder_seat_winner records for selected submissions
9. Send push notifications to winners via Firebase
10. Return winner list

---

## Flow Diagrams

### User Authentication Flow

```
┌──────────┐                                    ┌──────────┐
│  Client  │                                    │  Server  │
└────┬─────┘                                    └────┬─────┘
     │                                                │
     │  POST /auth/send-otp                           │
     │  { phone_number: "9876543210" }                │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼────────┐
     │                                    │ Validate phone      │
     │                                    │ Check rate limit    │
     │                                    │ Generate 6-digit OTP│
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Store in otp_logs   │
     │                                    │ Call Infobip API    │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { success: true, message: "OTP sent" }        │
     │<───────────────────────────────────────────────┤
     │                                                │
     │  POST /auth/verify-otp                         │
     │  { phone_number: "9876543210", otp: "123456" } │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼────────┐
     │                                    │ Validate OTP        │
     │                                    │ Check expiry        │
     │                                    │ Check attempts      │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Find user by phone  │
     │                                    │ If not found:       │
     │                                    │   return 400 error  │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Mark OTP verified   │
     │                                    │ Generate JWT tokens │
     │                                    │ Store refresh token │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { access_token, refresh_token, user_info }    │
     │<───────────────────────────────────────────────┤
     │                                                │
     │  Subsequent Requests with Bearer Token         │
     │  Authorization: Bearer <access_token>          │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Auth Middleware:    │
     │                                    │ - Validate JWT      │
     │                                    │ - Extract user_id   │
     │                                    │ - Load user entity  │
     │                                    │ - Set in context    │
     │                                    └───────────┬─────────┘
     │                                                │
     │  Protected Resource Response                   │
     │<───────────────────────────────────────────────┤
     │                                                │
```

### Thunder Seat Submission Flow

```
┌──────────┐                                    ┌──────────┐
│  Client  │                                    │  Server  │
└────┬─────┘                                    └────┬─────┘
     │                                                │
     │  GET /thunder-seat/current-week                │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Calculate week info │
     │                                    │ - week_number       │
     │                                    │ - start/end dates   │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { week_number: 3, start_date, end_date }      │
     │<───────────────────────────────────────────────┤
     │                                                │
     │  GET /questions/active                         │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Query active Qs     │
     │                                    │ where is_active=true│
     │                                    └───────────┬─────────┘
     │                                                │
     │  [ { id: 1, question_text: "...", ... } ]      │
     │<───────────────────────────────────────────────┤
     │                                                │
     │  POST /thunder-seat                            │
     │  Authorization: Bearer <token>                 │
     │  { week_number: 3, question_id: 1,             │
     │    answer: "My answer..." }                    │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Auth Middleware:    │
     │                                    │ - Extract user      │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Validate payload    │
     │                                    │ Check question      │
     │                                    │ Validate week       │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Check duplicate:    │
     │                                    │ unique(user_id,     │
     │                                    │   question_id,      │
     │                                    │   week_number)      │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Create submission   │
     │                                    │ in thunder_seat     │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { success: true, data: { id, ... } }          │
     │<───────────────────────────────────────────────┤
     │                                                │
     │  GET /thunder-seat/submissions                 │
     │  Authorization: Bearer <token>                 │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Query user's        │
     │                                    │ submissions         │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { success: true, data: [ ... ] }              │
     │<───────────────────────────────────────────────┤
     │                                                │
```

### Winner Selection Flow (Admin)

```
┌──────────┐                                    ┌──────────┐
│  Admin   │                                    │  Server  │
└────┬─────┘                                    └────┬─────┘
     │                                                │
     │  POST /admin/winners/select                    │
     │  X-API-Key: <admin_key>                        │
     │  { week_number: 3, number_of_winners: 10 }     │
     ├───────────────────────────────────────────────>│
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ API Key Middleware: │
     │                                    │ - Validate X-API-Key│
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Validate request    │
     │                                    │ Check week is past  │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Check if winners    │
     │                                    │ already selected    │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Get all submissions │
     │                                    │ for week_number     │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Random selection:   │
     │                                    │ - Shuffle list      │
     │                                    │ - Pick N winners    │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Transaction:        │
     │                                    │ - Create winner     │
     │                                    │   records           │
     │                                    │ - Commit            │
     │                                    └───────────┬─────────┘
     │                                                │
     │                                    ┌───────────▼─────────┐
     │                                    │ Async notification: │
     │                                    │ - Queue tasks for   │
     │                                    │   each winner       │
     │                                    │ - Send Firebase     │
     │                                    │   push + SMS        │
     │                                    └───────────┬─────────┘
     │                                                │
     │  { success: true, data: [ winners ] }          │
     │<───────────────────────────────────────────────┤
     │                                                │
```

---

## Error Handling Strategy

### Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "success": false,
  "error": "Human-readable error message",
  "details": {
    "field_name": "Validation error detail"
  }
}
```

### HTTP Status Codes

| Code | Usage | Example |
|------|-------|---------|
| 200 | Success | GET requests returning data |
| 201 | Created | POST requests creating resources |
| 400 | Bad Request | Validation errors, duplicate entries |
| 401 | Unauthorized | Invalid/expired JWT, missing auth header |
| 403 | Forbidden | Valid auth but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 429 | Too Many Requests | Rate limiting (OTP endpoints) |
| 500 | Internal Server Error | Unhandled exceptions, DB errors |

### Error Middleware

The global error handler middleware catches panics and formats errors:

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Error("Panic recovered: ", err)
                c.JSON(http.StatusInternalServerError, ErrorResponse{
                    Success: false,
                    Error: "Internal server error",
                })
            }
        }()
        c.Next()
    }
}
```

---

## Authentication & Authorization

### JWT Token Structure

**Access Token** (expires in 1 hour):
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "phone_number": "9876543210",
  "exp": 1705843200,
  "iat": 1705839600
}
```

**Refresh Token** (expires in 30 days):
```json
{
  "token_id": "660e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "exp": 1708435200,
  "iat": 1705839600
}
```

### Auth Middleware

Protects routes requiring authentication:

```go
func AuthMiddleware(db *gorm.DB, userRepo UserRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract Bearer token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, ErrorResponse{Error: "Missing authorization header"})
            c.Abort()
            return
        }
        
        // 2. Validate JWT signature and expiry
        claims, err := ValidateJWT(token)
        if err != nil {
            c.JSON(401, ErrorResponse{Error: "Invalid or expired token"})
            c.Abort()
            return
        }
        
        // 3. Load user from database
        user, err := userRepo.FindByID(claims.UserID)
        if err != nil || !user.IsActive {
            c.JSON(401, ErrorResponse{Error: "User not found or inactive"})
            c.Abort()
            return
        }
        
        // 4. Set user in context for handlers
        c.Set("user", user)
        c.Set("user_id", user.ID)
        c.Next()
    }
}
```

### API Key Middleware

Protects admin endpoints:

```go
func APIKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        expectedKey := config.GetConfig().XAPIKey
        
        if apiKey == "" || apiKey != expectedKey {
            c.JSON(401, ErrorResponse{Error: "Invalid or missing API key"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## Transaction Management

All service operations use transaction manager for ACID guarantees:

```go
type TransactionManager interface {
    WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error
}

// Example usage in service
func (s *ThunderSeatService) SubmitAnswer(ctx context.Context, req DTO, userID string) error {
    return s.txnManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // 1. Check duplicate submission
        exists, err := s.repo.CheckDuplicate(tx, userID, req.QuestionID, req.WeekNumber)
        if exists {
            return errors.New("duplicate submission")
        }
        
        // 2. Create submission
        submission := &ThunderSeat{
            UserID: userID,
            QuestionID: req.QuestionID,
            WeekNumber: req.WeekNumber,
            Answer: req.Answer,
        }
        
        if err := s.repo.Create(tx, submission); err != nil {
            return err // Transaction will auto-rollback
        }
        
        // 3. Update user stats (future enhancement)
        // ...
        
        return nil // Transaction will auto-commit
    })
}
```

---

## Background Task Processing

### Worker Pool Architecture

The system uses a fixed-size worker pool for async tasks:

```go
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    wg        sync.WaitGroup
    quit      chan bool
}

// Configuration
const (
    WORKER_POOL_SIZE = 10
    TASK_QUEUE_SIZE  = 100
)
```

### Task Types

1. **OTP Delivery**: Send SMS/WhatsApp via Infobip
2. **Push Notifications**: Send Firebase notifications to winners
3. **Email Notifications**: Send email confirmations (future)

### Task Submission Example

```go
// In AuthService.SendOTP
task := queue.Task{
    Type: "send_otp",
    Payload: map[string]interface{}{
        "phone_number": phoneNumber,
        "otp": otp,
        "template": "otp_template",
    },
    Retry: 3,
}

workerPool.Submit(task)
```

---

## External Service Integration

### Infobip Client

**Configuration**:
```go
type InfobipConfig struct {
    BaseURL  string // e.g., https://api.infobip.com
    APIKey   string
    WANumber string // WhatsApp sender number
}
```

**OTP Sending**:
```go
func (c *InfobipClient) SendOTP(phoneNumber, otp string) error {
    url := fmt.Sprintf("%s/sms/2/text/advanced", c.baseURL)
    
    payload := InfobipSMSRequest{
        Messages: []Message{
            {
                From: "ThumsUp",
                Destinations: []Destination{{To: phoneNumber}},
                Text: fmt.Sprintf("Your Thums Up OTP is: %s. Valid for 5 minutes.", otp),
            },
        },
    }
    
    req := NewRequest("POST", url)
    req.Header("Authorization", "App " + c.apiKey)
    req.JSON(payload)
    
    resp, err := req.Send()
    return err
}
```

### GCS Service

**Configuration**:
```go
type GcsConfig struct {
    BucketName string // e.g., thums-up-assets
    ProjectID  string
    GcpUrl     string
}
```

**File Upload**:
```go
func (g *GCSService) Upload(ctx context.Context, fileName string, content []byte) (string, error) {
    bucket := g.client.Bucket(g.bucketName)
    obj := bucket.Object(fileName)
    
    writer := obj.NewWriter(ctx)
    if _, err := writer.Write(content); err != nil {
        return "", err
    }
    if err := writer.Close(); err != nil {
        return "", err
    }
    
    // Generate signed URL (7 days expiry)
    url, err := g.GenerateSignedURL(fileName, 7*24*time.Hour)
    return url, err
}
```

### Firebase Client

**Push Notification**:
```go
func (f *FirebaseClient) SendNotification(userID, title, body string) error {
    // 1. Get device token from users table
    user, err := f.userRepo.FindByID(userID)
    if err != nil || user.DeviceToken == nil {
        return errors.New("device token not found")
    }
    
    // 2. Create FCM message
    message := &messaging.Message{
        Token: *user.DeviceToken,
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Data: map[string]string{
            "type": "winner_announcement",
            "user_id": userID,
        },
    }
    
    // 3. Send via Firebase
    _, err = f.client.Send(ctx, message)
    return err
}
```

---

## Graceful Shutdown

The server implements graceful shutdown to ensure:
1. All in-flight HTTP requests complete
2. Worker pool drains task queue
3. Database connections close cleanly

```go
func (s *Server) waitForShutdown(httpServer *http.Server) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Info("Shutdown signal received...")
    
    // 1. Stop accepting new requests (timeout: 30s)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := httpServer.Shutdown(ctx); err != nil {
        log.Error("Server forced shutdown:", err)
    }
    
    // 2. Shutdown worker pool
    if s.workerPool != nil {
        s.workerPool.Shutdown()
    }
    
    // 3. Close database connections
    if s.db != nil {
        sqlDB, _ := s.db.DB()
        sqlDB.Close()
    }
    
    log.Info("Server shutdown complete")
}
```

---

## Performance & Scalability

### Database Optimization

1. **Connection Pooling**:
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

2. **Indexes**:
- `users(phone_number)` - Unique index for login
- `users(email)` - Unique index for email lookup
- `users(referral_code)` - Unique index for referral validation
- `otp_logs(phone_number)` - Index for OTP lookup
- `refresh_tokens(token)` - Unique index for refresh flow
- `refresh_tokens(user_id)` - Index for user token cleanup
- `address(user_id)` - Index for user address queries
- `thunder_seat(user_id)` - Index for user submission lookup
- `thunder_seat(week_number)` - Index for winner selection
- `thunder_seat(user_id, question_id, week_number)` - Unique composite index

3. **Soft Deletes**: Uses `deleted_at` column with GORM for safe data retention

### Caching Strategy (Future Enhancement)

- **Redis**: Cache active questions, current week info, winner lists
- **TTL**: 5 minutes for questions, 1 hour for winners
- **Invalidation**: On question update, winner selection

### Rate Limiting

- **OTP Endpoints**: Max 3 requests per phone per hour
- **Implementation**: In-memory map with cleanup goroutine (production should use Redis)

---

## Security Considerations

### Input Validation

All inputs validated using `binding` tags:

```go
type SendOTPRequest struct {
    PhoneNumber string `json:"phone_number" binding:"required,min=10,max=10,numeric"`
}
```

### SQL Injection Prevention

GORM uses parameterized queries automatically:

```go
db.Where("phone_number = ?", phoneNumber).First(&user)
// Generates: SELECT * FROM users WHERE phone_number = $1
```

### CORS Configuration

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### Sensitive Data Handling

1. **Environment Variables**: All secrets loaded from `.env`
2. **No Logging**: Passwords, OTPs, tokens never logged
3. **HTTPS Only**: Production enforces TLS
4. **JWT Secret**: Strong random string (min 32 chars)

---

## Monitoring & Observability

### Logging

Structured logging with `logrus`:

```go
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "submit_answer",
    "week": weekNumber,
}).Info("Thunder seat submission")
```

### Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-25T12:00:00Z",
  "services": {
    "database": "connected",
    "gcs": "connected"
  }
}
```

### Metrics (Future Enhancement)

Prometheus metrics endpoints:
- `http_requests_total{method, endpoint, status}`
- `http_request_duration_seconds{method, endpoint}`
- `db_query_duration_seconds{query_type}`
- `worker_pool_queue_size`
- `active_users_count`

---

## Deployment Architecture

### Docker Configuration

```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o thums-up-backend

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/thums-up-backend .
COPY .env .
CMD ["./thums-up-backend", "server"]
```

### Environment Variables

```bash
# App Config
APP_ENV=production
APP_PORT=8080
ALLOWED_ORIGINS=https://thumsup.com

# Database
DB_HOST=postgres-instance
DB_PORT=5432
DB_USER=thumsup_user
DB_PASSWORD=secure_password
DB_NAME=thumsup_prod
DB_SSL_MODE=require

# JWT
JWT_SECRET_KEY=<32-char-random-string>
JWT_ACCESS_TOKEN_EXPIRY=3600
JWT_REFRESH_TOKEN_EXPIRY=2592000

# Infobip
INFOBIP_BASE_URL=https://api.infobip.com
INFOBIP_API_KEY=<api-key>
INFOBIP_WA_NUMBER=<whatsapp-number>

# GCS
GCP_BUCKET_NAME=thumsup-assets
GCP_PROJECT_ID=thumsup-project
GCP_URL=https://storage.googleapis.com

# Firebase
FIREBASE_SERVICE_KEY_PATH=./firebase-key.json

# Admin
X_API_KEY=<admin-api-key>
```

### Cloud Infrastructure (GCP)

1. **Compute**: Cloud Run (auto-scaling containers)
2. **Database**: Cloud SQL (PostgreSQL 14)
3. **Storage**: GCS buckets
4. **Secrets**: Secret Manager
5. **Logging**: Cloud Logging
6. **Monitoring**: Cloud Monitoring + Prometheus

---

## Testing Strategy

### Unit Tests

```bash
make test-unit
```

Coverage target: >80%

Example:
```go
func TestAuthService_SendOTP(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewOTPRepository()
    mockInfobip := mocks.NewInfobipClient()
    service := NewAuthService(mockRepo, mockInfobip)
    
    // Act
    err := service.SendOTP(ctx, "9876543210")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockInfobip.SendOTPCallCount())
}
```

### Integration Tests

```bash
make test-integration
```

Uses Docker Compose for test database:
```yaml
version: '3.8'
services:
  test-db:
    image: postgres:14
    environment:
      POSTGRES_DB: thumsup_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
```

### API Tests

Postman collection with environment variables:
- **Dev**: `http://localhost:8080`
- **Staging**: `https://api-staging.thumsup.com`
- **Prod**: `https://api.thumsup.com`

---

## Future Enhancements

1. **Redis Caching**: Cache hot data (questions, winners)
2. **Rate Limiting**: Redis-based distributed rate limiter
3. **Analytics**: User engagement metrics, submission patterns
4. **Leaderboard**: Real-time rankings of participants
5. **Referral Rewards**: Points/coins for successful referrals
6. **Multi-language Support**: Complete i18n for questions/notifications
7. **Image Uploads**: User profile pictures, answer attachments
8. **Real-time Updates**: WebSocket support for live winner announcements
9. **Advanced Winner Selection**: ML-based answer quality scoring
10. **Admin Dashboard**: Web UI for content management

---

## Appendix

### Constants

```go
const (
    WORKER_POOL_SIZE         = 10
    TASK_QUEUE_SIZE          = 100
    GRACEFUL_SHUTDOWN_TIMEOUT = 30 * time.Second
    OTP_EXPIRY               = 5 * time.Minute
    OTP_MAX_ATTEMPTS         = 3
    JWT_ACCESS_EXPIRY        = 1 * time.Hour
    JWT_REFRESH_EXPIRY       = 30 * 24 * time.Hour
    RATE_LIMIT_OTP           = 3 // requests per hour
)
```

### Error Messages

```go
const (
    ErrUserNotFound          = "User not found"
    ErrUserNotAuthenticated  = "User not authenticated"
    ErrInvalidCredentials    = "Invalid credentials"
    ErrOTPExpired            = "OTP has expired"
    ErrOTPInvalid            = "Invalid OTP"
    ErrDuplicateSubmission   = "You have already submitted for this week"
    ErrInvalidWeek           = "Invalid week number"
    ErrQuestionNotFound      = "Question not found"
    ErrAddressNotFound       = "Address not found"
    ErrPincodeNotDeliverable = "Pincode is not deliverable"
)
```

---

**Document Version**: 1.0  
**Last Updated**: January 2025  
**Maintained By**: Thums Up Backend Team



