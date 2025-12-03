package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"circles.diy/internal/auth"
	"circles.diy/internal/content"
	"circles.diy/internal/domain"
	"circles.diy/internal/models"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
)

type PostHandler struct {
	contentService *content.Service
	profileService *profile.Service
}

func NewPostHandler(contentService *content.Service, profileService *profile.Service) *PostHandler {
	return &PostHandler{
		contentService: contentService,
		profileService: profileService,
	}
}

// getProfileID extracts the profile ID from the request context
func getProfileID(r *http.Request) string {
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		return ""
	}
	return *session.ActiveProfileID
}

// HandlePosts handles listing and creating posts
func (h *PostHandler) HandlePosts(w http.ResponseWriter, r *http.Request) {
	h.handlePosts(w, r)
}

// HandlePost handles single post operations (get, update, delete)
func (h *PostHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	h.handlePost(w, r)
}

/* Comment handlers are disabled until comment service methods are ready
// HandleComment handles single comment operations
func (h *PostHandler) HandleComment(w http.ResponseWriter, r *http.Request) {
	h.handleComment(w, r)
}
*/

// handlePosts handles listing and creating posts
func (h *PostHandler) handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListPosts(w, r)
	case http.MethodPost:
		h.handleCreatePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePost handles single post operations (get, update, delete)
func (h *PostHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	switch r.Method {
	case http.MethodGet:
		h.handleGetPost(w, r, postID)
	case http.MethodPut, http.MethodPatch:
		h.handleUpdatePost(w, r, postID)
	case http.MethodDelete:
		h.handleDeletePost(w, r, postID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListPosts lists posts with optional filters
func (h *PostHandler) handleListPosts(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	circleID := r.URL.Query().Get("circle_id")
	authorID := r.URL.Query().Get("author_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var posts []*domain.Post
	var err error

	if circleID != "" {
		posts, err = h.contentService.GetPostsByCircle(r.Context(), circleID, limit, offset)
	} else if authorID != "" {
		posts, err = h.contentService.GetPostsByAuthor(r.Context(), authorID, limit, offset)
	} else {
		http.Error(w, "circle_id or author_id is required", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// handleCreatePost creates a new post
func (h *PostHandler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		CircleID       string `json:"circle_id"`
		Body           string `json:"body"`
		BodyFormat     string `json:"body_format"`
		ContentWarning string `json:"content_warning"`
		Visibility     string `json:"visibility"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Body = content.SanitizeInput(req.Body, req.BodyFormat)

	// Validate not empty after sanitization
	if content.IsEmptyContent(req.Body) {
		http.Error(w, "Post body cannot be empty", http.StatusBadRequest)
		return
	}

	// Create post
	post, err := h.contentService.CreatePost(
		r.Context(),
		req.CircleID,
		getProfileID(r),
		req.Body,
		req.BodyFormat,
		req.ContentWarning,
		req.Visibility,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// HandleCreatePostHTMX creates a new post and returns HTML partial for HTMX
func (h *PostHandler) HandleCreatePostHTMX(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := getProfileID(r)
	if profileID == "" {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	circleID := r.FormValue("circle_id")
	body := r.FormValue("body")
	bodyFormat := r.FormValue("body_format")
	contentWarning := r.FormValue("content_warning")
	visibility := r.FormValue("visibility")

	// Sanitize input
	body = content.SanitizeInput(body, bodyFormat)

	// Validate not empty after sanitization
	if content.IsEmptyContent(body) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="error-message" style="color: var(--error); padding: 1rem; border: 1px solid var(--error); border-radius: 8px; margin-bottom: 1rem;">Post body cannot be empty</div>`))
		return
	}

	// Create post
	post, err := h.contentService.CreatePost(
		r.Context(),
		circleID,
		profileID,
		body,
		bodyFormat,
		contentWarning,
		visibility,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message" style="color: var(--error); padding: 1rem; border: 1px solid var(--error); border-radius: 8px; margin-bottom: 1rem;">Failed to create post. Please try again.</div>`))
		return
	}

	// Convert domain.Post to models.CirclePost for template
	circlePost, err := h.mapPostToCirclePost(r.Context(), post, profileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message" style="color: var(--error); padding: 1rem; border: 1px solid var(--error); border-radius: 8px; margin-bottom: 1rem;">Failed to render post. Please refresh the page.</div>`))
		return
	}

	// Render post-card template
	w.Header().Set("Content-Type", "text/html")
	err = templates.GetTemplates().PostCard.ExecuteTemplate(w, "post-card", circlePost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message" style="color: var(--error); padding: 1rem; border: 1px solid var(--error); border-radius: 8px; margin-bottom: 1rem;">Failed to render post template</div>`))
		return
	}
}

// mapPostToCirclePost converts a domain.Post to models.CirclePost for template rendering
func (h *PostHandler) mapPostToCirclePost(ctx context.Context, post *domain.Post, viewerProfileID string) (*models.CirclePost, error) {
	// Get reaction counts
	reactionCounts, _ := h.contentService.GetReactionCounts(ctx, "post", post.ID)
	likeCount := reactionCounts["like"]

	// Check if viewer has liked
	userHasLiked := false
	if viewerProfileID != "" {
		userReaction, _ := h.contentService.GetUserReaction(ctx, "post", post.ID, viewerProfileID)
		userHasLiked = userReaction != nil && userReaction.Key == "like"
	}

	// Check if viewer can edit
	canEdit := post.AuthorProfileID == viewerProfileID

	// Format time
	formattedTime := formatTimeAgo(post.CreatedAt)

	// Check if edited
	isEdited := post.EditedAt != nil
	var editedAtStr *string
	if isEdited {
		editedStr := formatTimeAgo(*post.EditedAt)
		editedAtStr = &editedStr
	}

	// Get author profile information
	authorName := "Circle Member"
	authorAvatar := ""
	authorProfile, err := h.profileService.GetByID(ctx, post.AuthorProfileID)
	if err == nil && authorProfile != nil {
		if authorProfile.DisplayName != "" {
			authorName = authorProfile.DisplayName
		} else if authorProfile.Name != "" {
			authorName = authorProfile.Name
		} else if authorProfile.Handle != "" {
			authorName = authorProfile.Handle
		}
		authorAvatar = authorProfile.AvatarURL
	}

	return &models.CirclePost{
		ID:              post.ID,
		AuthorProfileID: post.AuthorProfileID,
		AuthorName:      authorName,
		AuthorAvatar:    authorAvatar,
		Body:            post.Body,
		BodyFormat:      post.BodyFormat,
		ContentWarning:  post.ContentWarning,
		Visibility:      post.Visibility,
		CreatedAt:       post.CreatedAt.Format(time.RFC3339),
		EditedAt:        editedAtStr,
		FormattedTime:   formattedTime,
		IsEdited:        isEdited,
		ReplyCount:      post.ReplyCount,
		LikeCount:       likeCount,
		UserHasLiked:    userHasLiked,
		CanEdit:         canEdit,
		ShowComments:    false,
	}, nil
}

// mapCommentToCircleComment converts a domain.Comment to models.CircleComment for template rendering
func (h *PostHandler) mapCommentToCircleComment(ctx context.Context, comment *domain.Comment, postID string, viewerProfileID string) (*models.CircleComment, error) {
	// Format time
	formattedTime := formatTimeAgo(comment.CreatedAt)

	// Check if edited
	isEdited := comment.EditedAt != nil
	var editedAtStr *string
	if isEdited {
		editedStr := formatTimeAgo(*comment.EditedAt)
		editedAtStr = &editedStr
	}

	// Get author profile information
	authorName := "Circle Member"
	authorAvatar := ""
	authorHandle := ""
	authorProfile, err := h.profileService.GetByID(ctx, comment.AuthorProfileID)
	if err == nil && authorProfile != nil {
		if authorProfile.DisplayName != "" {
			authorName = authorProfile.DisplayName
		} else if authorProfile.Name != "" {
			authorName = authorProfile.Name
		} else if authorProfile.Handle != "" {
			authorName = authorProfile.Handle
		}
		authorAvatar = authorProfile.AvatarURL
		authorHandle = authorProfile.Handle
	}

	// Check permissions
	canEdit := comment.AuthorProfileID == viewerProfileID
	canDelete := comment.AuthorProfileID == viewerProfileID

	return &models.CircleComment{
		ID:              comment.ID,
		PostID:          postID,
		AuthorProfileID: comment.AuthorProfileID,
		AuthorName:      authorName,
		AuthorAvatar:    authorAvatar,
		AuthorHandle:    authorHandle,
		Body:            comment.Body,
		BodyFormat:      comment.BodyFormat,
		CreatedAt:       comment.CreatedAt.Format(time.RFC3339),
		EditedAt:        editedAtStr,
		FormattedTime:   formattedTime,
		IsEdited:        isEdited,
		CanEdit:         canEdit,
		CanDelete:       canDelete,
		ReplyCount:      0, // For future nested comments
	}, nil
}

// handleGetPost retrieves a single post
func (h *PostHandler) handleGetPost(w http.ResponseWriter, r *http.Request, postID string) {
	post, err := h.contentService.GetPost(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// handleUpdatePost updates a post
func (h *PostHandler) handleUpdatePost(w http.ResponseWriter, r *http.Request, postID string) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user can edit this post
	canEdit, err := h.contentService.CanEditPost(r.Context(), postID, getProfileID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canEdit {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Body           string `json:"body"`
		BodyFormat     string `json:"body_format"`
		ContentWarning string `json:"content_warning"`
		Visibility     string `json:"visibility"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize input
	if req.Body != "" {
		req.Body = content.SanitizeInput(req.Body, req.BodyFormat)

		// Validate not empty after sanitization
		if content.IsEmptyContent(req.Body) {
			http.Error(w, "Post body cannot be empty", http.StatusBadRequest)
			return
		}
	}

	// Update post
	post, err := h.contentService.UpdatePost(
		r.Context(),
		postID,
		req.Body,
		req.BodyFormat,
		req.ContentWarning,
		req.Visibility,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// handleDeletePost deletes a post
func (h *PostHandler) handleDeletePost(w http.ResponseWriter, r *http.Request, postID string) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user can delete this post
	canDelete, err := h.contentService.CanDeletePost(r.Context(), postID, getProfileID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canDelete {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete post
	err = h.contentService.DeletePost(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Comment handlers

// HandleComments handles listing and creating comments for a post
func (h *PostHandler) HandleComments(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from path: /api/posts/{id}/comments
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	switch r.Method {
	case http.MethodGet:
		h.handleListComments(w, r, postID)
	case http.MethodPost:
		h.handleCreateComment(w, r, postID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListComments lists comments for a post and returns HTML for HTMX
func (h *PostHandler) handleListComments(w http.ResponseWriter, r *http.Request, postID string) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	comments, err := h.contentService.GetCommentsByPost(r.Context(), postID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message">Failed to load comments</div>`))
		return
	}

	// Get viewer profile ID for permission checks
	viewerProfileID := getProfileID(r)

	// Map comments to CircleComment with enriched data
	circleComments := make([]models.CircleComment, 0, len(comments))
	for _, comment := range comments {
		circleComment, err := h.mapCommentToCircleComment(r.Context(), comment, postID, viewerProfileID)
		if err != nil {
			// Log error but continue with other comments
			continue
		}
		circleComments = append(circleComments, *circleComment)
	}

	// Render comment-list template
	data := struct {
		PostID   string
		Comments []models.CircleComment
	}{PostID: postID, Comments: circleComments}

	w.Header().Set("Content-Type", "text/html")
	err = templates.GetTemplates().CommentList.ExecuteTemplate(w, "comment-list", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message">Failed to render comments</div>`))
		return
	}
}

// handleCreateComment creates a new comment and returns HTML for HTMX
func (h *PostHandler) handleCreateComment(w http.ResponseWriter, r *http.Request, postID string) {
	// Get user from session
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data (HTMX sends as form-urlencoded)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	body := r.FormValue("body")
	bodyFormat := r.FormValue("body_format")
	if bodyFormat == "" {
		bodyFormat = "plaintext"
	}

	// Sanitize input
	body = content.SanitizeInput(body, bodyFormat)

	// Validate not empty after sanitization
	if content.IsEmptyContent(body) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="error-message">Comment cannot be empty</div>`))
		return
	}

	// Create comment
	comment, err := h.contentService.CreateCommentOnPost(
		r.Context(),
		postID,
		*session.ActiveProfileID,
		body,
		bodyFormat,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message">Failed to post comment</div>`))
		return
	}

	// Map to CircleComment with enriched data
	circleComment, err := h.mapCommentToCircleComment(r.Context(), comment, postID, *session.ActiveProfileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message">Failed to render comment</div>`))
		return
	}

	// Render comment-card template
	w.Header().Set("Content-Type", "text/html")
	err = templates.GetTemplates().CommentCard.ExecuteTemplate(w, "comment-card", circleComment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error-message">Failed to render comment</div>`))
		return
	}

	// Get updated post for count
	post, _ := h.contentService.GetPost(r.Context(), postID)
	if post != nil {
		countText := fmt.Sprintf("%d %s", post.ReplyCount,
			func() string {
				if post.ReplyCount == 1 {
					return "comment"
				}
				return "comments"
			}())
		countHTML := fmt.Sprintf(`<span id="comment-count-%s" hx-swap-oob="true">%s</span>`,
			postID, countText)
		w.Write([]byte(countHTML))
	}
}

// HandleComment handles single comment operations
func (h *PostHandler) HandleComment(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from path: /api/comments/{id}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	commentID := pathParts[2]

	switch r.Method {
	case http.MethodGet:
		h.handleGetComment(w, r, commentID)
	case http.MethodPut, http.MethodPatch:
		h.handleUpdateComment(w, r, commentID)
	case http.MethodDelete:
		h.handleDeleteComment(w, r, commentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetComment retrieves a single comment
func (h *PostHandler) handleGetComment(w http.ResponseWriter, r *http.Request, commentID string) {
	comment, err := h.contentService.GetComment(r.Context(), commentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// handleUpdateComment updates a comment
func (h *PostHandler) handleUpdateComment(w http.ResponseWriter, r *http.Request, commentID string) {
	// Get user from session
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user can edit this comment
	canEdit, err := h.contentService.CanEditComment(r.Context(), commentID, *session.ActiveProfileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canEdit {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Body       string `json:"body"`
		BodyFormat string `json:"body_format"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize input
	if req.Body != "" {
		req.Body = content.SanitizeInput(req.Body, req.BodyFormat)

		// Validate not empty after sanitization
		if content.IsEmptyContent(req.Body) {
			http.Error(w, "Comment body cannot be empty", http.StatusBadRequest)
			return
		}
	}

	// Update comment
	comment, err := h.contentService.UpdateComment(
		r.Context(),
		commentID,
		req.Body,
		req.BodyFormat,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// handleDeleteComment deletes a comment and returns HTML for HTMX
func (h *PostHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request, commentID string) {
	// Get user from session
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get postID from query parameter (required for decrementing reply count)
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "post_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Check if user can delete this comment
	canDelete, err := h.contentService.CanDeleteComment(r.Context(), commentID, *session.ActiveProfileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canDelete {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete comment
	err = h.contentService.DeleteCommentOnPost(r.Context(), commentID, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return empty body (removes card) + OOB count update
	w.Header().Set("Content-Type", "text/html")

	// Get updated post for count
	post, _ := h.contentService.GetPost(r.Context(), postID)
	if post != nil {
		countText := fmt.Sprintf("%d %s", post.ReplyCount,
			func() string {
				if post.ReplyCount == 1 {
					return "comment"
				}
				return "comments"
			}())
		countHTML := fmt.Sprintf(`<span id="comment-count-%s" hx-swap-oob="true">%s</span>`,
			postID, countText)
		w.Write([]byte(countHTML))
	}
}

// handlePostReactions handles reactions on posts
func (h *PostHandler) handlePostReactions(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	h.handleReactions(w, r, "post", postID)
}

// handleCommentReactions handles reactions on comments
func (h *PostHandler) handleCommentReactions(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	commentID := pathParts[2]

	h.handleReactions(w, r, "comment", commentID)
}

// handleReactions handles reactions on any target type
func (h *PostHandler) handleReactions(w http.ResponseWriter, r *http.Request, targetType, targetID string) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetReactions(w, r, targetType, targetID)
	case http.MethodPost:
		h.handleAddReaction(w, r, targetType, targetID)
	case http.MethodDelete:
		h.handleRemoveReaction(w, r, targetType, targetID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetReactions retrieves reactions for a target
func (h *PostHandler) handleGetReactions(w http.ResponseWriter, r *http.Request, targetType, targetID string) {
	// Get counts grouped by key
	counts, err := h.contentService.GetReactionCounts(r.Context(), targetType, targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user's reaction if authenticated
	var userReaction *domain.Reaction
	user := auth.GetUser(r.Context())
	if user != nil {
		userReaction, _ = h.contentService.GetUserReaction(r.Context(), targetType, targetID, getProfileID(r))
	}

	response := models.ReactionResponse{
		Counts:       counts,
		UserReaction: userReaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAddReaction adds a reaction to a target
func (h *PostHandler) handleAddReaction(w http.ResponseWriter, r *http.Request, targetType, targetID string) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Key string `json:"key"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add reaction
	reaction, err := h.contentService.AddReaction(
		r.Context(),
		targetType,
		targetID,
		getProfileID(r),
		req.Key,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reaction)
}

// handleRemoveReaction removes a reaction from a target
func (h *PostHandler) handleRemoveReaction(w http.ResponseWriter, r *http.Request, targetType, targetID string) {
	// Get user from session
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Remove reaction
	err := h.contentService.RemoveReaction(r.Context(), targetType, targetID, getProfileID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
