-- Replace per-message read tracking with a last_read_at watermark on
-- chat_participants. This eliminates the expensive NOT EXISTS subquery
-- against message_reads and enables a single-query inbox fetch.

-- 1. Add the watermark column.
ALTER TABLE chat_participants
  ADD COLUMN last_read_at TIMESTAMP;

-- 2. Backfill from existing message_reads data: set last_read_at to the
--    created_at of the most recently read message in each chat.
UPDATE chat_participants cp
SET last_read_at = sub.max_read
FROM (
    SELECT mr.profile_id, m.chat_id, MAX(m.created_at) AS max_read
    FROM message_reads mr
    JOIN messages m ON m.id = mr.message_id
    GROUP BY mr.profile_id, m.chat_id
) sub
WHERE cp.profile_id = sub.profile_id
  AND cp.chat_id = sub.chat_id;

-- 3. Composite index that serves keyset pagination AND watermark-based
--    unread counts efficiently.
CREATE INDEX idx_messages_chat_created
  ON messages (chat_id, created_at DESC, id DESC)
  WHERE deleted_at IS NULL;
