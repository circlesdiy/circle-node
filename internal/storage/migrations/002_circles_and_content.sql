-- Circles and content schema

-- Circles (groups/communities)
CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    visibility VARCHAR(50) NOT NULL DEFAULT 'public',
    auto_mod_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_circles_owner_profile_id ON circles(owner_profile_id);
CREATE INDEX idx_circles_visibility ON circles(visibility);
CREATE INDEX idx_circles_deleted_at ON circles(deleted_at);

-- Circle memberships
CREATE TABLE circle_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    state VARCHAR(50) NOT NULL DEFAULT 'active',
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(circle_id, profile_id)
);

CREATE INDEX idx_circle_memberships_circle_id ON circle_memberships(circle_id);
CREATE INDEX idx_circle_memberships_profile_id ON circle_memberships(profile_id);
CREATE INDEX idx_circle_memberships_state ON circle_memberships(state);

-- Roles within circles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(circle_id, name)
);

CREATE INDEX idx_roles_circle_id ON roles(circle_id);

-- Permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- Role permissions mapping
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Profile roles (assign roles to profiles in circles)
CREATE TABLE profile_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(circle_id, profile_id, role_id)
);

CREATE INDEX idx_profile_roles_circle_id ON profile_roles(circle_id);
CREATE INDEX idx_profile_roles_profile_id ON profile_roles(profile_id);
CREATE INDEX idx_profile_roles_role_id ON profile_roles(role_id);

-- Posts
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    author_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    body_format VARCHAR(50) NOT NULL DEFAULT 'markdown',
    content_warning TEXT,
    visibility VARCHAR(50) NOT NULL DEFAULT 'public',
    reply_count INTEGER NOT NULL DEFAULT 0,
    attachments_count INTEGER NOT NULL DEFAULT 0,
    moderation_status VARCHAR(50) NOT NULL DEFAULT 'approved',
    edited_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    cid VARCHAR(500),
    signature TEXT
);

CREATE INDEX idx_posts_circle_id ON posts(circle_id);
CREATE INDEX idx_posts_author_profile_id ON posts(author_profile_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_visibility ON posts(visibility);
CREATE INDEX idx_posts_deleted_at ON posts(deleted_at);

-- Discussions (threaded conversations)
CREATE TABLE discussions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    author_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    reply_count INTEGER NOT NULL DEFAULT 0,
    edited_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    cid VARCHAR(500),
    signature TEXT
);

CREATE INDEX idx_discussions_circle_id ON discussions(circle_id);
CREATE INDEX idx_discussions_author_profile_id ON discussions(author_profile_id);
CREATE INDEX idx_discussions_created_at ON discussions(created_at DESC);
CREATE INDEX idx_discussions_is_pinned ON discussions(is_pinned);
CREATE INDEX idx_discussions_deleted_at ON discussions(deleted_at);

-- Comments (can be attached to posts or discussions)
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    body_format VARCHAR(50) NOT NULL DEFAULT 'markdown',
    reply_to_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    edited_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    cid VARCHAR(500),
    signature TEXT
);

CREATE INDEX idx_comments_author_profile_id ON comments(author_profile_id);
CREATE INDEX idx_comments_reply_to_comment_id ON comments(reply_to_comment_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at);

-- Post comments junction table
CREATE TABLE post_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    UNIQUE(post_id, comment_id)
);

CREATE INDEX idx_post_comments_post_id ON post_comments(post_id);
CREATE INDEX idx_post_comments_comment_id ON post_comments(comment_id);

-- Discussion comments junction table
CREATE TABLE discussion_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    discussion_id UUID NOT NULL REFERENCES discussions(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    UNIQUE(discussion_id, comment_id)
);

CREATE INDEX idx_discussion_comments_discussion_id ON discussion_comments(discussion_id);
CREATE INDEX idx_discussion_comments_comment_id ON discussion_comments(comment_id);

-- Reactions (emojis on posts, comments, messages)
CREATE TABLE reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_type VARCHAR(50) NOT NULL,
    target_id UUID NOT NULL,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    key VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(target_type, target_id, profile_id, key)
);

CREATE INDEX idx_reactions_target ON reactions(target_type, target_id);
CREATE INDEX idx_reactions_profile_id ON reactions(profile_id);

-- Attachments
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    mime_type VARCHAR(100) NOT NULL,
    size INTEGER NOT NULL,
    storage_url TEXT NOT NULL,
    cid VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_attachments_owner_profile_id ON attachments(owner_profile_id);
CREATE INDEX idx_attachments_deleted_at ON attachments(deleted_at);

-- Post attachments junction table
CREATE TABLE post_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    attachment_id UUID NOT NULL REFERENCES attachments(id) ON DELETE CASCADE,
    UNIQUE(post_id, attachment_id)
);

CREATE INDEX idx_post_attachments_post_id ON post_attachments(post_id);
CREATE INDEX idx_post_attachments_attachment_id ON post_attachments(attachment_id);

-- Triggers for updated_at
CREATE TRIGGER update_circles_updated_at BEFORE UPDATE ON circles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_circle_memberships_updated_at BEFORE UPDATE ON circle_memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_discussions_updated_at BEFORE UPDATE ON discussions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
