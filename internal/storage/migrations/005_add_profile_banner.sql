-- Migration: Add banner_url to profiles table
-- This adds support for profile banner images

ALTER TABLE profiles
ADD COLUMN banner_url TEXT;

-- Add index for faster queries if needed
CREATE INDEX IF NOT EXISTS idx_profiles_banner_url ON profiles(banner_url) WHERE banner_url IS NOT NULL;
