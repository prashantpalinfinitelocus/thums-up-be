-- Add name column to notify_me table
ALTER TABLE notify_me ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';

-- Update the default constraint to allow proper insertion
ALTER TABLE notify_me ALTER COLUMN name DROP DEFAULT;

