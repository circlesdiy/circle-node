package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"circles.diy/internal/domain"
)

type Service struct {
	repo         *Repository
	reactionRepo *ReactionRepository
}

func NewService(repo *Repository, reactionRepo *ReactionRepository) *Service {
	return &Service{
		repo:         repo,
		reactionRepo: reactionRepo,
	}
}

// Post operations

// CreatePost creates a new post with validation and sanitization
func (s *Service) CreatePost(ctx context.Context, circleID, authorProfileID, body, bodyFormat, contentWarning, visibility string) (*domain.Post, error) {
	// Validate required fields
	if circleID == "" {
		return nil, fmt.Errorf("circle_id is required")
	}
	if authorProfileID == "" {
		return nil, fmt.Errorf("author_profile_id is required")
	}
	if body == "" {
		return nil, fmt.Errorf("body is required")
	}

	// Validate body format
	if bodyFormat == "" {
		bodyFormat = domain.BodyFormatMarkdown
	}
	if bodyFormat != domain.BodyFormatMarkdown && bodyFormat != domain.BodyFormatPlaintext && bodyFormat != domain.BodyFormatHTML {
		return nil, fmt.Errorf("invalid body_format: %s", bodyFormat)
	}

	// Validate visibility
	if visibility == "" {
		visibility = domain.PostVisibilityMembersOnly
	}
	if visibility != domain.PostVisibilityPublic && visibility != domain.PostVisibilityMembersOnly && visibility != domain.PostVisibilityPrivate {
		return nil, fmt.Errorf("invalid visibility: %s", visibility)
	}

	// Create post
	now := time.Now().UTC()
	post := &domain.Post{
		ID:               uuid.New().String(),
		CircleID:         circleID,
		AuthorProfileID:  authorProfileID,
		Body:             body,
		BodyFormat:       bodyFormat,
		ContentWarning:   contentWarning,
		Visibility:       visibility,
		ReplyCount:       0,
		AttachmentsCount: 0,
		ModerationStatus: domain.ModerationStatusApproved, // Auto-approve for now
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := s.repo.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

// GetPost retrieves a post by ID
func (s *Service) GetPost(ctx context.Context, postID string) (*domain.Post, error) {
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

// GetPostsByCircle retrieves posts for a circle with pagination
func (s *Service) GetPostsByCircle(ctx context.Context, circleID string, limit, offset int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	posts, err := s.repo.GetPostsByCircleID(ctx, circleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by circle: %w", err)
	}

	return posts, nil
}

// GetPostsByAuthor retrieves posts by an author with pagination
func (s *Service) GetPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	posts, err := s.repo.GetPostsByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author: %w", err)
	}

	return posts, nil
}

// UpdatePost updates an existing post
func (s *Service) UpdatePost(ctx context.Context, postID, body, bodyFormat, contentWarning, visibility string) (*domain.Post, error) {
	// Get existing post
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}

	// Validate body format if provided
	if bodyFormat != "" {
		if bodyFormat != domain.BodyFormatMarkdown && bodyFormat != domain.BodyFormatPlaintext && bodyFormat != domain.BodyFormatHTML {
			return nil, fmt.Errorf("invalid body_format: %s", bodyFormat)
		}
		post.BodyFormat = bodyFormat
	}

	// Validate visibility if provided
	if visibility != "" {
		if visibility != domain.PostVisibilityPublic && visibility != domain.PostVisibilityMembersOnly && visibility != domain.PostVisibilityPrivate {
			return nil, fmt.Errorf("invalid visibility: %s", visibility)
		}
		post.Visibility = visibility
	}

	// Update fields
	if body != "" {
		post.Body = body
	}
	post.ContentWarning = contentWarning

	now := time.Now().UTC()
	post.EditedAt = &now
	post.UpdatedAt = now

	err = s.repo.UpdatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

// DeletePost soft-deletes a post
func (s *Service) DeletePost(ctx context.Context, postID string) error {
	err := s.repo.DeletePost(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

// CanEditPost checks if a user can edit a post
func (s *Service) CanEditPost(ctx context.Context, postID, profileID string) (bool, error) {
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return false, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return false, fmt.Errorf("post not found")
	}

	// Only the author can edit their post
	return post.AuthorProfileID == profileID, nil
}

// CanDeletePost checks if a user can delete a post
func (s *Service) CanDeletePost(ctx context.Context, postID, profileID string) (bool, error) {
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return false, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return false, fmt.Errorf("post not found")
	}

	// Author can delete, or circle owners/admins (to be implemented with circle service integration)
	return post.AuthorProfileID == profileID, nil
}

// Comment operations

// CreateCommentOnPost creates a new comment on a post
func (s *Service) CreateCommentOnPost(ctx context.Context, postID, authorProfileID, body, bodyFormat string) (*domain.Comment, error) {
	// Validate required fields
	if postID == "" {
		return nil, fmt.Errorf("post_id is required")
	}
	if authorProfileID == "" {
		return nil, fmt.Errorf("author_profile_id is required")
	}
	if body == "" {
		return nil, fmt.Errorf("body is required")
	}

	// Validate body format
	if bodyFormat == "" {
		bodyFormat = domain.BodyFormatMarkdown
	}
	if bodyFormat != domain.BodyFormatMarkdown && bodyFormat != domain.BodyFormatPlaintext && bodyFormat != domain.BodyFormatHTML {
		return nil, fmt.Errorf("invalid body_format: %s", bodyFormat)
	}

	// Verify post exists
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}

	// Create comment
	now := time.Now()
	comment := &domain.Comment{
		ID:              uuid.New().String(),
		AuthorProfileID: authorProfileID,
		Body:            body,
		BodyFormat:      bodyFormat,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Use the new repository method that handles junction table
	err = s.repo.CreateCommentForPost(ctx, comment, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Increment reply count on post
	err = s.repo.IncrementReplyCount(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment reply count: %w", err)
	}

	return comment, nil
}

// GetComment retrieves a comment by ID
func (s *Service) GetComment(ctx context.Context, commentID string) (*domain.Comment, error) {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return nil, fmt.Errorf("comment not found")
	}
	return comment, nil
}

// GetCommentsByPost retrieves comments for a post with pagination
func (s *Service) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*domain.Comment, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	comments, err := s.repo.GetCommentsByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return comments, nil
}

// UpdateComment updates an existing comment
func (s *Service) UpdateComment(ctx context.Context, commentID, body, bodyFormat string) (*domain.Comment, error) {
	// Get existing comment
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return nil, fmt.Errorf("comment not found")
	}

	// Validate body format if provided
	if bodyFormat != "" {
		if bodyFormat != domain.BodyFormatMarkdown && bodyFormat != domain.BodyFormatPlaintext && bodyFormat != domain.BodyFormatHTML {
			return nil, fmt.Errorf("invalid body_format: %s", bodyFormat)
		}
		comment.BodyFormat = bodyFormat
	}

	// Update fields
	if body != "" {
		comment.Body = body
	}

	now := time.Now()
	comment.EditedAt = &now
	comment.UpdatedAt = now

	err = s.repo.UpdateComment(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

// DeleteCommentOnPost soft-deletes a comment and decrements post reply count
func (s *Service) DeleteCommentOnPost(ctx context.Context, commentID, postID string) error {
	// Verify comment exists
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return fmt.Errorf("comment not found")
	}

	err = s.repo.DeleteComment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Decrement reply count on post
	err = s.repo.DecrementReplyCount(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to decrement reply count: %w", err)
	}

	return nil
}

// CanEditComment checks if a user can edit a comment
func (s *Service) CanEditComment(ctx context.Context, commentID, profileID string) (bool, error) {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return false, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return false, fmt.Errorf("comment not found")
	}

	// Only the author can edit their comment
	return comment.AuthorProfileID == profileID, nil
}

// CanDeleteComment checks if a user can delete a comment
func (s *Service) CanDeleteComment(ctx context.Context, commentID, profileID string) (bool, error) {
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return false, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return false, fmt.Errorf("comment not found")
	}

	// Author can delete, or circle owners/admins (to be implemented with circle service integration)
	return comment.AuthorProfileID == profileID, nil
}

// Reaction operations

// AddReaction adds or updates a reaction to a target
func (s *Service) AddReaction(ctx context.Context, targetType, targetID, profileID, key string) (*domain.Reaction, error) {
	// Validate required fields
	if targetType == "" {
		return nil, fmt.Errorf("target_type is required")
	}
	if targetID == "" {
		return nil, fmt.Errorf("target_id is required")
	}
	if profileID == "" {
		return nil, fmt.Errorf("profile_id is required")
	}
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Check if reaction already exists
	existing, err := s.reactionRepo.GetReactionByProfileAndTarget(ctx, profileID, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing reaction: %w", err)
	}

	// If exists and same key, do nothing
	if existing != nil && existing.Key == key {
		return existing, nil
	}

	// If exists but different key, delete old reaction
	if existing != nil {
		err = s.reactionRepo.DeleteReaction(ctx, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete old reaction: %w", err)
		}
	}

	// Create new reaction
	reaction := &domain.Reaction{
		ID:         uuid.New().String(),
		TargetType: targetType,
		TargetID:   targetID,
		ProfileID:  profileID,
		Key:        key,
		CreatedAt:  time.Now().UTC(),
	}

	err = s.reactionRepo.CreateReaction(ctx, reaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create reaction: %w", err)
	}

	return reaction, nil
}

// RemoveReaction removes a reaction from a target
func (s *Service) RemoveReaction(ctx context.Context, targetType, targetID, profileID string) error {
	existing, err := s.reactionRepo.GetReactionByProfileAndTarget(ctx, profileID, targetType, targetID)
	if err != nil {
		return fmt.Errorf("failed to get reaction: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("reaction not found")
	}

	err = s.reactionRepo.DeleteReaction(ctx, existing.ID)
	if err != nil {
		return fmt.Errorf("failed to delete reaction: %w", err)
	}

	return nil
}

// GetReactions retrieves all reactions for a target
func (s *Service) GetReactions(ctx context.Context, targetType, targetID string) ([]*domain.Reaction, error) {
	reactions, err := s.reactionRepo.GetReactionsByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}
	return reactions, nil
}

// GetReactionCounts retrieves reaction counts grouped by key
func (s *Service) GetReactionCounts(ctx context.Context, targetType, targetID string) (map[string]int, error) {
	counts, err := s.reactionRepo.GetReactionCounts(ctx, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction counts: %w", err)
	}
	return counts, nil
}

// GetUserReaction retrieves the user's reaction for a target
func (s *Service) GetUserReaction(ctx context.Context, targetType, targetID, profileID string) (*domain.Reaction, error) {
	reaction, err := s.reactionRepo.GetReactionByProfileAndTarget(ctx, profileID, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reaction: %w", err)
	}
	return reaction, nil
}
