-- Optional circle scope for chats (multiple circle-linked chats per circle allowed)

ALTER TABLE chats
ADD COLUMN circle_id UUID REFERENCES circles(id) ON DELETE CASCADE;

CREATE INDEX idx_chats_circle_id ON chats(circle_id)
WHERE circle_id IS NOT NULL AND deleted_at IS NULL;
