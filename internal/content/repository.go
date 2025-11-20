package content

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"circles.diy/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Post methods

func (r *Repository) CreatePost(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (
			id, circle_id, author_profile_id, body, body_format,
			content_warning, visibility, reply_count, attachments_count,
			moderation_status, created_at, updated_at, cid, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		post.ID,
		post.CircleID,
		post.AuthorProfileID,
		post.Body,
		post.BodyFormat,
		post.ContentWarning,
		post.Visibility,
		post.ReplyCount,
		post.AttachmentsCount,
		post.ModerationStatus,
		post.CreatedAt,
		post.UpdatedAt,
		post.CID,
		post.Signature,
	)

	return err
}

func (r *Repository) GetPostByID(ctx context.Context, postID string) (*domain.Post, error) {
	query := `
		SELECT
			id, circle_id, author_profile_id, body, body_format,
			content_warning, visibility, reply_count, attachments_count,
			moderation_status, edited_at, created_at, updated_at, deleted_at,
			cid, signature
		FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var post domain.Post
	var editedAt, deletedAt *time.Time

	err := r.pool.QueryRow(ctx, query, postID).Scan(
		&post.ID,
		&post.CircleID,
		&post.AuthorProfileID,
		&post.Body,
		&post.BodyFormat,
		&post.ContentWarning,
		&post.Visibility,
		&post.ReplyCount,
		&post.AttachmentsCount,
		&post.ModerationStatus,
		&editedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&deletedAt,
		&post.CID,
		&post.Signature,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	post.EditedAt = editedAt
	post.DeletedAt = deletedAt

	return &post, nil
}

func (r *Repository) GetPostsByCircleID(ctx context.Context, circleID string, limit, offset int) ([]*domain.Post, error) {
	query := `
		SELECT
			id, circle_id, author_profile_id, body, body_format,
			content_warning, visibility, reply_count, attachments_count,
			moderation_status, edited_at, created_at, updated_at, deleted_at,
			cid, signature
		FROM posts
		WHERE circle_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, circleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var editedAt, deletedAt *time.Time

		err := rows.Scan(
			&post.ID,
			&post.CircleID,
			&post.AuthorProfileID,
			&post.Body,
			&post.BodyFormat,
			&post.ContentWarning,
			&post.Visibility,
			&post.ReplyCount,
			&post.AttachmentsCount,
			&post.ModerationStatus,
			&editedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&deletedAt,
			&post.CID,
			&post.Signature,
		)
		if err != nil {
			return nil, err
		}

		post.EditedAt = editedAt
		post.DeletedAt = deletedAt

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *Repository) GetPostsByAuthorID(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error) {
	query := `
		SELECT
			id, circle_id, author_profile_id, body, body_format,
			content_warning, visibility, reply_count, attachments_count,
			moderation_status, edited_at, created_at, updated_at, deleted_at,
			cid, signature
		FROM posts
		WHERE author_profile_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var editedAt, deletedAt *time.Time

		err := rows.Scan(
			&post.ID,
			&post.CircleID,
			&post.AuthorProfileID,
			&post.Body,
			&post.BodyFormat,
			&post.ContentWarning,
			&post.Visibility,
			&post.ReplyCount,
			&post.AttachmentsCount,
			&post.ModerationStatus,
			&editedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&deletedAt,
			&post.CID,
			&post.Signature,
		)
		if err != nil {
			return nil, err
		}

		post.EditedAt = editedAt
		post.DeletedAt = deletedAt

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *Repository) UpdatePost(ctx context.Context, post *domain.Post) error {
	query := `
		UPDATE posts
		SET body = $1, body_format = $2, content_warning = $3,
		    visibility = $4, moderation_status = $5, edited_at = $6,
		    updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		post.Body,
		post.BodyFormat,
		post.ContentWarning,
		post.Visibility,
		post.ModerationStatus,
		post.EditedAt,
		post.UpdatedAt,
		post.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *Repository) DeletePost(ctx context.Context, postID string) error {
	query := `
		UPDATE posts
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, time.Now().UTC(), postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *Repository) IncrementReplyCount(ctx context.Context, postID string) error {
	query := `
		UPDATE posts
		SET reply_count = reply_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, postID)
	return err
}

func (r *Repository) DecrementReplyCount(ctx context.Context, postID string) error {
	query := `
		UPDATE posts
		SET reply_count = GREATEST(reply_count - 1, 0)
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, postID)
	return err
}

// Comment methods - Implemented with junction table pattern (post_comments)

// CreateCommentForPost creates a new comment and links it to a post using the junction table
func (r *Repository) CreateCommentForPost(ctx context.Context, comment *domain.Comment, postID string) error {
	// Use a transaction to ensure atomicity
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert comment
	commentQuery := `
		INSERT INTO comments (
			id, author_profile_id, body, body_format, reply_to_comment_id,
			cid, signature, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.Exec(ctx, commentQuery,
		comment.ID,
		comment.AuthorProfileID,
		comment.Body,
		comment.BodyFormat,
		comment.ReplyToCommentID,
		comment.CID,
		comment.Signature,
		comment.CreatedAt,
		comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	// Create junction table entry
	junctionQuery := `
		INSERT INTO post_comments (post_id, comment_id)
		VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, junctionQuery, postID, comment.ID)
	if err != nil {
		return fmt.Errorf("failed to link comment to post: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCommentByID retrieves a single comment by ID
func (r *Repository) GetCommentByID(ctx context.Context, commentID string) (*domain.Comment, error) {
	query := `
		SELECT
			id, author_profile_id, body, body_format, reply_to_comment_id,
			cid, signature, edited_at, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var comment domain.Comment
	var editedAt, deletedAt *time.Time
	var replyToCommentID *string

	err := r.pool.QueryRow(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.AuthorProfileID,
		&comment.Body,
		&comment.BodyFormat,
		&replyToCommentID,
		&comment.CID,
		&comment.Signature,
		&editedAt,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	comment.EditedAt = editedAt
	comment.DeletedAt = deletedAt
	comment.ReplyToCommentID = replyToCommentID

	return &comment, nil
}

// GetCommentsByPostID retrieves all comments for a post via the junction table
func (r *Repository) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*domain.Comment, error) {
	query := `
		SELECT
			c.id, c.author_profile_id, c.body, c.body_format, c.reply_to_comment_id,
			c.cid, c.signature, c.edited_at, c.created_at, c.updated_at, c.deleted_at
		FROM comments c
		INNER JOIN post_comments pc ON pc.comment_id = c.id
		WHERE pc.post_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize with empty slice instead of nil to ensure JSON returns []
	comments := make([]*domain.Comment, 0)
	for rows.Next() {
		var comment domain.Comment
		var editedAt, deletedAt *time.Time
		var replyToCommentID *string

		err := rows.Scan(
			&comment.ID,
			&comment.AuthorProfileID,
			&comment.Body,
			&comment.BodyFormat,
			&replyToCommentID,
			&comment.CID,
			&comment.Signature,
			&editedAt,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		comment.EditedAt = editedAt
		comment.DeletedAt = deletedAt
		comment.ReplyToCommentID = replyToCommentID

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// UpdateComment updates an existing comment
func (r *Repository) UpdateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET body = $1, body_format = $2, edited_at = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		comment.Body,
		comment.BodyFormat,
		comment.EditedAt,
		comment.UpdatedAt,
		comment.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// DeleteComment soft-deletes a comment and removes junction table entry
func (r *Repository) DeleteComment(ctx context.Context, commentID string) error {
	// Use a transaction to ensure atomicity
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Soft delete the comment
	deleteQuery := `
		UPDATE comments
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, deleteQuery, time.Now().UTC(), commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	// Remove from junction tables (post_comments and discussion_comments)
	// Note: We don't delete from junction tables to maintain data integrity,
	// but we could if needed. The deleted_at check in queries handles filtering.

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
