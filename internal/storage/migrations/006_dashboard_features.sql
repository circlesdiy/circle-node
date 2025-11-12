-- Dashboard features migration

-- Add visual fields to circles (avatar, icon, banner)
ALTER TABLE circles ADD COLUMN icon VARCHAR(10);
ALTER TABLE circles ADD COLUMN icon_bg_color VARCHAR(50);
ALTER TABLE circles ADD COLUMN avatar_url TEXT;
ALTER TABLE circles ADD COLUMN banner_url TEXT;

-- Coordination tracking for events
CREATE TABLE event_coordination_needs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    author_profile_id UUID NOT NULL REFERENCES profiles(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE INDEX idx_event_coordination_needs_event_id ON event_coordination_needs(event_id);
CREATE INDEX idx_event_coordination_needs_resolved ON event_coordination_needs(resolved_at);

-- Indexes for dashboard queries
CREATE INDEX idx_activities_dashboard ON activities(circle_id, occurred_at DESC);
CREATE INDEX idx_events_upcoming ON events(start_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_messages_unread ON messages(chat_id, created_at DESC) WHERE deleted_at IS NULL;
