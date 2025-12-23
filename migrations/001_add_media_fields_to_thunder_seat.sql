-- Migration: Add media fields to thunder_seat table
-- Created: 2025-12-22
-- Description: Adds support for audio/video file uploads in Thunder Seat submissions

-- Add media_url column to store the public URL of the uploaded file
ALTER TABLE thunder_seat 
ADD COLUMN IF NOT EXISTS media_url TEXT;

-- Add media_key column to store the GCS object key/path
ALTER TABLE thunder_seat 
ADD COLUMN IF NOT EXISTS media_key TEXT;

-- Add media_type column to store whether it's 'audio' or 'video'
ALTER TABLE thunder_seat 
ADD COLUMN IF NOT EXISTS media_type VARCHAR(50);

-- Create an index on media_type for faster queries
CREATE INDEX IF NOT EXISTS idx_thunder_seat_media_type ON thunder_seat(media_type);

-- Add a comment to the table explaining the new fields
COMMENT ON COLUMN thunder_seat.media_url IS 'Public URL of the uploaded media file (audio/video)';
COMMENT ON COLUMN thunder_seat.media_key IS 'GCS object key/path for the media file';
COMMENT ON COLUMN thunder_seat.media_type IS 'Type of media file: audio or video';


