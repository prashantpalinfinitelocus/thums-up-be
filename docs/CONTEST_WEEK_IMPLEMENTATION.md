# Contest Week Implementation Summary

## Overview

The system has been updated to support week-wise contest management with specific date ranges and configurable winner counts per week.

## Changes Made

### 1. New Entity: ContestWeek

**File**: `entities/contest_week.go`

New entity to manage contest weeks with the following fields:
- `week_number`: Unique identifier for each week (1-6)
- `start_date`: Contest week start date
- `end_date`: Contest week end date
- `winner_count`: Number of winners for this week
- `is_active`: Flag to indicate if this week is currently active

### 2. New Repository: ContestWeekRepository

**File**: `repository/contest_week_repository.go`

Provides methods to:
- Find contest week by week number
- Find the currently active week
- Get all contest weeks
- Deactivate all weeks (used when activating a new week)

### 3. New Service: ContestWeekService

**File**: `services/contest_week_service.go`

Handles business logic for:
- Creating contest weeks
- Activating/deactivating weeks
- Retrieving contest week information
- Validating date ranges and week configurations

### 4. New Handler: ContestWeekHandler

**File**: `handlers/contest_week_handler.go`

API endpoints for:
- `POST /api/v1/contest-weeks/` - Create a new contest week
- `GET /api/v1/contest-weeks/` - Get all contest weeks
- `GET /api/v1/contest-weeks/:weekNumber` - Get specific week
- `GET /api/v1/contest-weeks/active` - Get currently active week
- `POST /api/v1/contest-weeks/activate` - Activate a contest week

### 5. Updated Thunder Seat Service

**File**: `services/thunder_seat_service.go`

Changes:
- `SubmitAnswer()`: Now automatically uses the active contest week
- Validates submissions are within the active week's date range
- Removed manual `week_number` requirement from users
- `GetCurrentWeek()`: Returns active contest week details

### 6. Updated Winner Service

**File**: `services/winner_service.go`

Changes:
- `SelectWinners()`: Automatically uses winner count from contest week configuration
- Validates against configured winner count per week
- Prevents selecting more winners than configured
- Filters entries by week number

### 7. Updated DTOs

**File**: `dtos/thunder_seat_dto.go`

Changes:
- `ThunderSeatSubmitRequest`: Removed `week_number` field (auto-detected)
- `SelectWinnersRequest`: Removed `number_of_winners` field (from config)
- `CurrentWeekResponse`: Added `winner_count` and `is_active` fields
- Added new DTOs:
  - `ContestWeekRequest`
  - `ContestWeekResponse`
  - `ActivateWeekRequest`

### 8. Updated Repository

**File**: `repository/thunder_seat_repository.go`

Changes:
- Added `GetRandomEntriesByWeek()`: Select winners from specific week
- Updated `GetRandomEntries()`: Ensures distinct users

### 9. New Routes

**File**: `routes/contest_week_routes.go`

New route group for contest week management with proper authentication.

### 10. Updated Server Initialization

**Files**: `cmd/server.go`, `cmd/types.go`

Changes:
- Added `contestWeek` repository initialization
- Added `contestWeekService` initialization
- Added `contestWeekHandler` initialization
- Injected `contestWeekRepo` into dependent services

### 11. Database Migration

**File**: `utils/db_migrations.go`

Added `ContestWeek` entity to AutoMigrate list.

### 12. Setup Script

**File**: `scripts/setup-contest-weeks.sh`

New script to automatically create all 6 contest weeks with the correct dates and winner counts.

## Key Improvements

### 1. Simplified User Experience
- Users no longer need to specify week number when submitting
- Week number is automatically determined from active contest week

### 2. Centralized Configuration
- Winner counts are managed centrally per week
- Admin can configure all weeks upfront
- Easy to modify winner counts if needed

### 3. Better Control
- Only one week can be active at a time
- Submissions only allowed during active week dates
- Automatic validation of date ranges

### 4. Flexible Winner Selection
- Can select winners incrementally
- Respects configured winner count limits
- Prevents duplicate winners in same week

### 5. Week Management
- Easy activation/deactivation of weeks
- Clear visibility of active week
- Historical tracking of all weeks

## Workflow

### Setup Phase
1. Create all 6 contest weeks using API or script
2. Configure dates and winner counts for each week

### During Contest
1. Activate Week 1
2. Users submit answers (automatically tagged with Week 1)
3. After week ends, select winners for Week 1
4. Activate Week 2
5. Repeat for all 6 weeks

### Winner Selection
1. Call winner selection API with week number
2. System automatically:
   - Gets winner count from contest week config
   - Selects random users from that week's submissions
   - Excludes users who already won in that week
   - Respects the configured winner limit

## Database Schema

### contest_week Table
```sql
CREATE TABLE contest_week (
    id SERIAL PRIMARY KEY,
    week_number INTEGER UNIQUE NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    winner_count INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_by VARCHAR(255),
    updated_on TIMESTAMP
);
```

## API Examples

### Create Contest Week
```bash
curl -X POST http://localhost:8080/api/v1/contest-weeks/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "week_number": 1,
    "start_date": "2025-01-10",
    "end_date": "2025-01-16",
    "winner_count": 6
  }'
```

### Activate Week
```bash
curl -X POST http://localhost:8080/api/v1/contest-weeks/activate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"week_number": 1}'
```

### Submit Answer (User)
```bash
curl -X POST http://localhost:8080/api/v1/thunder-seat/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer USER_TOKEN" \
  -d '{
    "question_id": 1,
    "answer": "My answer"
  }'
```

### Select Winners
```bash
curl -X POST http://localhost:8080/api/v1/admin/winners/select \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{"week_number": 1}'
```

## Validation Rules

1. **Contest Week Creation**
   - Week number must be unique
   - End date must be after start date
   - Winner count must be at least 1

2. **Week Activation**
   - Only one week can be active at a time
   - Previous active week is automatically deactivated

3. **Answer Submission**
   - Must have an active contest week
   - Current time must be within active week's date range
   - User can submit only once per question

4. **Winner Selection**
   - Contest week must exist
   - Cannot select more winners than configured
   - Users already selected as winners are excluded from subsequent selections

## Testing

To test the implementation:

1. Start the server
2. Run the setup script: `./scripts/setup-contest-weeks.sh`
3. Activate week 1
4. Submit test answers
5. Select winners for week 1
6. Verify winner count matches configuration

## Migration Notes

For existing deployments:
1. Run the application to auto-create the `contest_week` table
2. Create contest weeks using the API or script
3. Activate the appropriate week
4. Existing `thunder_seat` records will work as-is

