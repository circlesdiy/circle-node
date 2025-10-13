-- Chat, events, and notifications schema

-- Chats
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    last_message_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_chats_type ON chats(type);
CREATE INDEX idx_chats_last_message_at ON chats(last_message_at DESC);
CREATE INDEX idx_chats_deleted_at ON chats(deleted_at);

-- Chat participants
CREATE TABLE chat_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(chat_id, profile_id)
);

CREATE INDEX idx_chat_participants_chat_id ON chat_participants(chat_id);
CREATE INDEX idx_chat_participants_profile_id ON chat_participants(profile_id);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text',
    encryption_scheme VARCHAR(50),
    nonce VARCHAR(255),
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    is_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_sender_profile_id ON messages(sender_profile_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_reply_to_message_id ON messages(reply_to_message_id);
CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);

-- Message read receipts
CREATE TABLE message_reads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    read_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(message_id, profile_id)
);

CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);
CREATE INDEX idx_message_reads_profile_id ON message_reads(profile_id);

-- Message attachments junction table
CREATE TABLE message_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    attachment_id UUID NOT NULL REFERENCES attachments(id) ON DELETE CASCADE,
    UNIQUE(message_id, attachment_id)
);

CREATE INDEX idx_message_attachments_message_id ON message_attachments(message_id);
CREATE INDEX idx_message_attachments_attachment_id ON message_attachments(attachment_id);

-- Events
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    organizer_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    location TEXT,
    timezone VARCHAR(100),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    requires_ticket BOOLEAN NOT NULL DEFAULT FALSE,
    capacity INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_events_circle_id ON events(circle_id);
CREATE INDEX idx_events_organizer_profile_id ON events(organizer_profile_id);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_deleted_at ON events(deleted_at);

-- Event RSVPs
CREATE TABLE event_rsvps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, profile_id)
);

CREATE INDEX idx_event_rsvps_event_id ON event_rsvps(event_id);
CREATE INDEX idx_event_rsvps_profile_id ON event_rsvps(profile_id);
CREATE INDEX idx_event_rsvps_status ON event_rsvps(status);

-- Orders (for tickets and marketplace)
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    payment_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer_profile_id ON orders(buyer_profile_id);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);

-- Tickets
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    ticket_status VARCHAR(50) NOT NULL,
    qr_code TEXT,
    issued_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tickets_event_id ON tickets(event_id);
CREATE INDEX idx_tickets_profile_id ON tickets(profile_id);
CREATE INDEX idx_tickets_order_id ON tickets(order_id);
CREATE INDEX idx_tickets_ticket_status ON tickets(ticket_status);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    actor_profile_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    channel VARCHAR(50) NOT NULL DEFAULT 'in_app',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_profile_id ON notifications(profile_id);
CREATE INDEX idx_notifications_actor_profile_id ON notifications(actor_profile_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);

-- Activity log (for feeds and audit trails)
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    actor_profile_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id UUID,
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW(),
    payload JSONB
);

CREATE INDEX idx_activities_circle_id ON activities(circle_id);
CREATE INDEX idx_activities_actor_profile_id ON activities(actor_profile_id);
CREATE INDEX idx_activities_occurred_at ON activities(occurred_at DESC);
CREATE INDEX idx_activities_target ON activities(target_type, target_id);

-- Reports (content moderation)
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    target_type VARCHAR(50) NOT NULL,
    target_id UUID NOT NULL,
    reason TEXT NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_reporter_profile_id ON reports(reporter_profile_id);
CREATE INDEX idx_reports_target ON reports(target_type, target_id);
CREATE INDEX idx_reports_state ON reports(state);

-- Moderation actions
CREATE TABLE moderation_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    moderator_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    target_type VARCHAR(50) NOT NULL,
    target_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moderation_actions_moderator_profile_id ON moderation_actions(moderator_profile_id);
CREATE INDEX idx_moderation_actions_target ON moderation_actions(target_type, target_id);

-- Blocks (user-level blocking)
CREATE TABLE blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blocker_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    blocked_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(blocker_profile_id, blocked_profile_id)
);

CREATE INDEX idx_blocks_blocker_profile_id ON blocks(blocker_profile_id);
CREATE INDEX idx_blocks_blocked_profile_id ON blocks(blocked_profile_id);

-- Mutes (user-level muting)
CREATE TABLE mutes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    muter_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    muted_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(muter_profile_id, muted_profile_id)
);

CREATE INDEX idx_mutes_muter_profile_id ON mutes(muter_profile_id);
CREATE INDEX idx_mutes_muted_profile_id ON mutes(muted_profile_id);

-- Export bundles (data portability)
CREATE TABLE export_bundles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    format VARCHAR(50) NOT NULL,
    location TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_export_bundles_profile_id ON export_bundles(profile_id);

-- Triggers for updated_at
CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_event_rsvps_updated_at BEFORE UPDATE ON event_rsvps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE ON tickets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reports_updated_at BEFORE UPDATE ON reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
