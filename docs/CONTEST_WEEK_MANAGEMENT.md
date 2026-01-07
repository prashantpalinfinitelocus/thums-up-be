# Contest Week Management

## Overview

The contest runs for 6 weeks with specific date ranges and winner counts per week.

## Contest Schedule

| Week | Dates | Winner Count |
|------|-------|--------------|
| Wk1  | 10-16th Jan | 6 |
| Wk2  | 17-23rd Jan | 6 |
| Wk3  | 24-30th Jan | 7 |
| Wk4  | 31st-6th Feb | 7 |
| Wk5  | 7-13th Feb | 7 |
| Wk6  | 14-20th Feb | 7 |
| **Total** | | **40** |

## API Endpoints

### Contest Week Management

#### Create Contest Week
```
POST /api/v1/contest-weeks/
Authorization: Required
```

Request:
```json
{
  "week_number": 1,
  "start_date": "2025-01-10",
  "end_date": "2025-01-16",
  "winner_count": 6
}
```

#### Get All Contest Weeks
```
GET /api/v1/contest-weeks/
```

#### Get Contest Week by Number
```
GET /api/v1/contest-weeks/:weekNumber
```

#### Get Active Contest Week
```
GET /api/v1/contest-weeks/active
```

#### Activate Contest Week
```
POST /api/v1/contest-weeks/activate
Authorization: Required
```

Request:
```json
{
  "week_number": 1
}
```

### User Submissions

#### Submit Answer
```
POST /api/v1/thunder-seat/submit
Authorization: Required
```

Request:
```json
{
  "question_id": 1,
  "answer": "User's answer"
}
```

Note: `week_number` is automatically determined from the active contest week.

#### Get Current Active Week
```
GET /api/v1/thunder-seat/current-week
```

### Winner Management

#### Select Winners for a Week
```
POST /api/v1/admin/winners/select
API-Key: Required
```

Request:
```json
{
  "week_number": 1
}
```

Note: `number_of_winners` is automatically determined from the contest week configuration.

#### Get Winners by Week
```
GET /api/v1/winners/week/:weekNumber
```

#### Get All Winners
```
GET /api/v1/winners/?limit=10&offset=0
```

## Setup Instructions

### 1. Initialize Contest Weeks

Run the setup script:
```bash
./scripts/setup-contest-weeks.sh
```

Or create them manually using the API.

### 2. Activate a Week

To start accepting submissions for a week:
```bash
curl -X POST http://localhost:8080/api/v1/contest-weeks/activate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"week_number": 1}'
```

### 3. Users Submit Answers

Users can submit answers only during the active contest week period.

### 4. Select Winners

After the week ends:
```bash
curl -X POST http://localhost:8080/api/v1/admin/winners/select \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{"week_number": 1}'
```

### 5. Move to Next Week

Activate the next week:
```bash
curl -X POST http://localhost:8080/api/v1/contest-weeks/activate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"week_number": 2}'
```

## Key Features

1. **Week-based Management**: Contest is divided into predefined weeks with specific date ranges
2. **Automatic Week Detection**: User submissions automatically use the active week
3. **Date Validation**: Submissions only allowed during active week's date range
4. **Winner Count Control**: Each week has its own winner count configuration
5. **Unique User Winners**: Same user cannot win multiple times in the same week
6. **Sequential Winner Selection**: Can select winners incrementally if needed

## Business Rules

1. Only one contest week can be active at a time
2. Users can only submit answers during the active week's date range
3. Winner selection respects the configured winner count per week
4. Once winners are selected for a week, the count cannot exceed the configured limit
5. Users who have already won in a week are excluded from subsequent selections for that week

