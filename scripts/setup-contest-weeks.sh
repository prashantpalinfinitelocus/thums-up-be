#!/bin/bash

API_URL="${API_URL:-http://localhost:8080/api/v1}"

echo "Setting up contest weeks..."

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 1,
    "start_date": "2025-01-10",
    "end_date": "2025-01-16",
    "winner_count": 6
  }'

echo ""

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 2,
    "start_date": "2025-01-17",
    "end_date": "2025-01-23",
    "winner_count": 6
  }'

echo ""

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 3,
    "start_date": "2025-01-24",
    "end_date": "2025-01-30",
    "winner_count": 7
  }'

echo ""

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 4,
    "start_date": "2025-01-31",
    "end_date": "2025-02-06",
    "winner_count": 7
  }'

echo ""

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 5,
    "start_date": "2025-02-07",
    "end_date": "2025-02-13",
    "winner_count": 7
  }'

echo ""

curl -X POST "$API_URL/contest-weeks/" \
  -H "Content-Type: application/json" \
  -d '{
    "week_number": 6,
    "start_date": "2025-02-14",
    "end_date": "2025-02-20",
    "winner_count": 7
  }'

echo ""
echo "Contest weeks setup complete!"

