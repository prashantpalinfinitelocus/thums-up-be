# Swagger Documentation Updates

## Overview

Updated Swagger documentation to include the new Contest Week Management APIs and updated Thunder Seat/Winner endpoints.

## New API Endpoints Added

### Contest Week Management

#### 1. Get All Contest Weeks
```
GET /api/v1/contest-weeks
```
- **Description**: Retrieve all contest weeks with their configurations
- **Authentication**: None
- **Response**: Array of ContestWeekResponse objects

#### 2. Create Contest Week
```
POST /api/v1/contest-weeks
```
- **Description**: Create a new contest week with date range and winner count
- **Authentication**: Bearer token required
- **Request Body**: ContestWeekRequest
- **Response**: ContestWeekResponse with success message

#### 3. Get Active Contest Week
```
GET /api/v1/contest-weeks/active
```
- **Description**: Retrieve the currently active contest week
- **Authentication**: None
- **Response**: ContestWeekResponse

#### 4. Activate Contest Week
```
POST /api/v1/contest-weeks/activate
```
- **Description**: Activate a specific contest week (deactivates others)
- **Authentication**: Bearer token required
- **Request Body**: ActivateWeekRequest
- **Response**: ContestWeekResponse with success message

#### 5. Get Contest Week by Number
```
GET /api/v1/contest-weeks/{weekNumber}
```
- **Description**: Retrieve a specific contest week by week number
- **Authentication**: None
- **Path Parameter**: weekNumber (integer)
- **Response**: ContestWeekResponse

### Thunder Seat APIs

#### 1. Submit Answer
```
POST /api/v1/thunder-seat/submit
```
- **Description**: Submit an answer for a question. Week number is automatically detected from active contest week.
- **Authentication**: Bearer token required
- **Request Body**: ThunderSeatSubmitRequest (no week_number needed)
- **Response**: ThunderSeatResponse with success message
- **Note**: Validates submission is within active week's date range

#### 2. Get User Submissions
```
GET /api/v1/thunder-seat/submissions
```
- **Description**: Retrieve all submissions for authenticated user
- **Authentication**: Bearer token required
- **Response**: Array of ThunderSeatResponse objects

#### 3. Get Current Active Week
```
GET /api/v1/thunder-seat/current-week
```
- **Description**: Retrieve the currently active contest week details
- **Authentication**: None
- **Response**: CurrentWeekResponse (includes winner_count and is_active)

### Winner APIs

#### 1. Get All Winners
```
GET /api/v1/winners?limit={limit}&offset={offset}
```
- **Description**: Retrieve all winners with pagination support
- **Authentication**: None
- **Query Parameters**:
  - limit (required, 1-100): Number of results per page
  - offset (optional, min 0): Number of results to skip
- **Response**: Array of WinnerResponse with PaginationMeta

#### 2. Get Winners by Week
```
GET /api/v1/winners/week/{weekNumber}
```
- **Description**: Retrieve all winners for a specific week
- **Authentication**: None
- **Path Parameter**: weekNumber (integer)
- **Response**: Array of WinnerResponse objects

### Admin APIs

#### 1. Select Winners
```
POST /api/v1/admin/winners/select
```
- **Description**: Randomly select winners for a specific week. Winner count is automatically determined from contest week configuration.
- **Authentication**: API Key required (X-API-Key header)
- **Request Body**: SelectWinnersRequest (no number_of_winners needed)
- **Response**: Array of WinnerResponse with success message
- **Note**: Prevents selecting more winners than configured for the week

## New Data Models (Definitions)

### ContestWeekRequest
```json
{
  "week_number": 1,
  "start_date": "2025-01-10",
  "end_date": "2025-01-16",
  "winner_count": 6
}
```
- **Required Fields**: week_number, start_date, end_date, winner_count
- **Validation**: 
  - week_number must be minimum 1
  - start_date and end_date must be in YYYY-MM-DD format
  - winner_count must be minimum 1

### ContestWeekResponse
```json
{
  "id": 1,
  "week_number": 1,
  "start_date": "2025-01-10",
  "end_date": "2025-01-16",
  "winner_count": 6,
  "is_active": true,
  "created_on": "2025-01-01T00:00:00Z"
}
```

### ActivateWeekRequest
```json
{
  "week_number": 1
}
```
- **Required Fields**: week_number
- **Validation**: week_number must be minimum 1

### ThunderSeatSubmitRequest (Updated)
```json
{
  "question_id": 1,
  "answer": "User's answer"
}
```
- **Changed**: Removed `week_number` field (auto-detected)
- **Required Fields**: question_id, answer

### ThunderSeatResponse
```json
{
  "id": 1,
  "user_id": "uuid-string",
  "question_id": 1,
  "week_number": 1,
  "answer": "User's answer",
  "created_on": "2025-01-10T12:00:00Z"
}
```

### CurrentWeekResponse (Updated)
```json
{
  "week_number": 1,
  "start_date": "2025-01-10",
  "end_date": "2025-01-16",
  "winner_count": 6,
  "is_active": true
}
```
- **Changed**: Added `winner_count` and `is_active` fields

### SelectWinnersRequest (Updated)
```json
{
  "week_number": 1
}
```
- **Changed**: Removed `number_of_winners` field (retrieved from contest week config)
- **Required Fields**: week_number

### WinnerResponse
```json
{
  "id": 1,
  "user_id": "uuid-string",
  "thunder_seat_id": 1,
  "week_number": 1,
  "created_on": "2025-01-16T23:59:59Z"
}
```

### PaginationMeta
```json
{
  "page": 1,
  "page_size": 10,
  "total_pages": 5,
  "total_count": 45
}
```

## Security Definitions

### Bearer Authentication (Existing)
- **Type**: API Key
- **Header**: Authorization
- **Format**: Bearer {token}
- **Used For**: User authenticated endpoints

### API Key Authentication (New)
- **Type**: API Key
- **Header**: X-API-Key
- **Used For**: Admin endpoints (winner selection)

## Changes Summary

### Breaking Changes in Request Models

1. **ThunderSeatSubmitRequest**
   - ❌ Removed: `week_number` field
   - ✅ Now: Automatically detected from active contest week

2. **SelectWinnersRequest**
   - ❌ Removed: `number_of_winners` field
   - ✅ Now: Retrieved from contest week configuration

### Enhanced Response Models

1. **CurrentWeekResponse**
   - ✅ Added: `winner_count` field
   - ✅ Added: `is_active` field

## Tags Added

- **Contest Week**: All contest week management endpoints
- **Thunder Seat**: Answer submission and user submission queries
- **Winners**: Winner retrieval endpoints
- **Admin**: Administrative operations (winner selection)

## Error Codes

### Contest Week Endpoints
- **400**: Invalid request, week already exists, or date validation failed
- **401**: Unauthorized (missing/invalid bearer token)
- **404**: Contest week not found or no active week
- **500**: Internal server error

### Thunder Seat Endpoints
- **400**: No active contest week, already submitted, or outside date range
- **401**: Unauthorized (missing/invalid bearer token)
- **404**: Question not found
- **500**: Internal server error

### Winner Endpoints
- **400**: Invalid parameters or winners already selected
- **401**: Unauthorized (invalid API key for admin endpoints)
- **404**: No eligible entries found
- **500**: Internal server error

## How to View Updated Swagger

The Swagger documentation is embedded in the Go application. To view:

1. Start the server:
```bash
./bin/server server
```

2. Access the Swagger UI (if configured):
```
http://localhost:8080/swagger/index.html
```

3. Or use the raw JSON:
```
http://localhost:8080/swagger/doc.json
```

## Testing with Swagger

### Example 1: Create and Activate Week
1. Create Week 1 using POST `/contest-weeks/`
2. Activate Week 1 using POST `/contest-weeks/activate`
3. Verify activation using GET `/contest-weeks/active`

### Example 2: User Submission Flow
1. Get active week using GET `/thunder-seat/current-week`
2. Submit answer using POST `/thunder-seat/submit` (no week_number needed)
3. View submissions using GET `/thunder-seat/submissions`

### Example 3: Winner Selection Flow
1. Get week details using GET `/contest-weeks/{weekNumber}`
2. Select winners using POST `/admin/winners/select` (winner count auto-determined)
3. View winners using GET `/winners/week/{weekNumber}`

## Migration Notes for API Consumers

### If You're Using Thunder Seat Submit API
**Before:**
```json
POST /api/v1/thunder-seat/submit
{
  "week_number": 1,
  "question_id": 1,
  "answer": "My answer"
}
```

**After:**
```json
POST /api/v1/thunder-seat/submit
{
  "question_id": 1,
  "answer": "My answer"
}
```

### If You're Using Winner Selection API
**Before:**
```json
POST /api/v1/admin/winners/select
{
  "week_number": 1,
  "number_of_winners": 6
}
```

**After:**
```json
POST /api/v1/admin/winners/select
{
  "week_number": 1
}
```

## Contest Schedule Reference

According to the contest configuration:

| Week | Dates | Winner Count | Week Number |
|------|-------|--------------|-------------|
| Wk1  | Jan 10-16, 2025 | 6 | 1 |
| Wk2  | Jan 17-23, 2025 | 6 | 2 |
| Wk3  | Jan 24-30, 2025 | 7 | 3 |
| Wk4  | Jan 31 - Feb 6, 2025 | 7 | 4 |
| Wk5  | Feb 7-13, 2025 | 7 | 5 |
| Wk6  | Feb 14-20, 2025 | 7 | 6 |
| **Total** | | **40** | |

