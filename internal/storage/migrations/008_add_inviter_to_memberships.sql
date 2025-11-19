-- Add inviter tracking to circle memberships
-- This allows us to track who actually sent each invitation

ALTER TABLE circle_memberships
ADD COLUMN inviter_profile_id UUID REFERENCES profiles(id) ON DELETE SET NULL;

-- Add index for performance when querying by inviter
CREATE INDEX idx_circle_memberships_inviter ON circle_memberships(inviter_profile_id);

-- Update existing invited memberships to set inviter as circle owner
-- This is a best-effort migration for existing data
UPDATE circle_memberships cm
SET inviter_profile_id = c.owner_profile_id
FROM circles c
WHERE cm.circle_id = c.id
  AND cm.state = 'invited'
  AND cm.inviter_profile_id IS NULL;
