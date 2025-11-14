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

// Comment methods - TODO: Implement based on actual domain model structure
// The domain model uses reply_to_comment_id for nested threading,
// not a direct post_id relationship

/*
func (r *Repository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (
			id, post_id, author_profile_id, body, body_format,
			reply_count, moderation_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		comment.ID,
		comment.PostID,
		comment.AuthorProfileID,
		comment.Body,
		comment.BodyFormat,
		comment.ReplyCount,
		comment.ModerationStatus,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	return err
}

func (r *Repository) GetCommentByID(ctx context.Context, commentID string) (*domain.Comment, error) {
	query := `
		SELECT
			id, post_id, author_profile_id, body, body_format,
			reply_count, moderation_status, edited_at, created_at,
			updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var comment domain.Comment
	var editedAt, deletedAt *time.Time

	err := r.pool.QueryRow(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorProfileID,
		&comment.Body,
		&comment.BodyFormat,
		&comment.ReplyCount,
		&comment.ModerationStatus,
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

	return &comment, nil
}

func (r *Repository) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*domain.Comment, error) {
	query := `
		SELECT
			id, post_id, author_profile_id, body, body_format,
			reply_count, moderation_status, edited_at, created_at,
			updated_at, deleted_at
		FROM comments
		WHERE post_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var comment domain.Comment
		var editedAt, deletedAt *time.Time

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorProfileID,
			&comment.Body,
			&comment.BodyFormat,
			&comment.ReplyCount,
			&comment.ModerationStatus,
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

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *Repository) UpdateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET body = $1, body_format = $2, moderation_status = $3,
		    edited_at = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		comment.Body,
		comment.BodyFormat,
		comment.ModerationStatus,
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

func (r *Repository) DeleteComment(ctx context.Context, commentID string) error {
	query := `
		UPDATE comments
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, time.Now().UTC(), commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}
*/
