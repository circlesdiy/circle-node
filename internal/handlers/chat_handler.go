package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"circles.diy/internal/auth"
	"circles.diy/internal/chat"
	"circles.diy/internal/circle"
	"circles.diy/internal/domain"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/preferences"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
	"go.uber.org/zap"
)

const (
	sseHeartbeatInterval = 20 * time.Second
	sseRetryMillis       = 5000
	inboxPageLimit       = 50
	messageHistoryLimit  = 50
	maxMessageBodyLen    = 4000
)

// ChatHandler wires the chat service, hub, and profile/circle lookups needed
// to render the chat page and answer HTMX/SSE requests.
type ChatHandler struct {
	chatService    *chat.Service
	hub            *chat.Hub
	profileService *profile.Service
	circleService  *circle.Service
	prefsService   *preferences.Service
	assetVersion   string
	logger         *zap.Logger
}

func NewChatHandler(
	chatService *chat.Service,
	hub *chat.Hub,
	profileService *profile.Service,
	circleService *circle.Service,
	prefsService *preferences.Service,
	assetVersion string,
	logger *zap.Logger,
) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		hub:            hub,
		profileService: profileService,
		circleService:  circleService,
		prefsService:   prefsService,
		assetVersion:   assetVersion,
		logger:         logger,
	}
}

// RegisterRoutes mounts chat routes on the given mux. The auth handler is
// used to enforce authentication on every chat endpoint.
func (h *ChatHandler) RegisterRoutes(mux *http.ServeMux, authHandler *auth.Handler) {
	mux.HandleFunc("/chat", authHandler.RequireAuth(h.handlePage))
	mux.HandleFunc("/chat/", authHandler.RequireAuth(h.handlePage))

	mux.HandleFunc("GET /api/chats/sidebar", authHandler.RequireAuth(h.handleSidebar))
	mux.HandleFunc("GET /api/chats/events", authHandler.RequireAuth(h.handleSSE))
	mux.HandleFunc("GET /api/chats/{chatID}/messages", authHandler.RequireAuth(h.handleGetMessages))
	mux.HandleFunc("POST /api/chats/{chatID}/messages", authHandler.RequireAuth(h.handleSendMessage))
	mux.HandleFunc("GET /api/chats/{chatID}/messages/since", authHandler.RequireAuth(h.handleMessagesSince))

	mux.HandleFunc("GET /api/chats/new", authHandler.RequireAuth(h.handleNewChatModal))
	mux.HandleFunc("POST /api/chats/dm", authHandler.RequireAuth(h.handleCreateDM))
	mux.HandleFunc("POST /api/chats/group", authHandler.RequireAuth(h.handleCreateGroup))
	mux.HandleFunc("POST /api/circles/{circleID}/chats", authHandler.RequireAuth(h.handleCreateCircleChat))
}

// handlePage renders the main chat page, optionally pre-selecting a thread.
// Supports /chat/{id} and /chat?circle={id} entry points.
func (h *ChatHandler) handlePage(w http.ResponseWriter, r *http.Request) {
	user, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	activeChatID := extractActiveChatID(r.URL.Path)
	if activeChatID == "" {
		if circleID := r.URL.Query().Get("circle"); circleID != "" {
			chat, err := h.chatService.GetOrCreateDefaultCircleChat(ctx, circleID, profileID)
			if err != nil {
				h.logger.Warn("circle chat resolve failed",
					zap.String("circle_id", circleID),
					zap.String("profile_id", profileID),
					zap.Error(err),
				)
			} else if chat != nil {
				http.Redirect(w, r, "/chat/"+chat.ID, http.StatusSeeOther)
				return
			}
		}
	}

	inbox, err := h.chatService.ListInbox(ctx, profileID)
	if err != nil {
		h.logger.Error("failed to load inbox",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load chats", http.StatusInternalServerError)
		return
	}

	conversations, err := h.hydrateConversations(ctx, inbox, profileID)
	if err != nil {
		h.logger.Error("failed to hydrate conversations",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load chats", http.StatusInternalServerError)
		return
	}

	var activeChat *models.Conversation
	var messages []models.Message

	if activeChatID != "" {
		for i := range conversations {
			if conversations[i].ID == activeChatID {
				activeChat = &conversations[i]
				break
			}
		}
		if activeChat == nil {
			// Viewer may have access through circle membership but not yet be in
			// the sidebar inbox (no messages yet). Resolve directly.
			chat, err := h.chatService.GetChat(ctx, activeChatID, profileID)
			if err == nil && chat != nil {
				conv, err := h.buildConversation(ctx, chat, nil, 0, profileID)
				if err == nil {
					conversations = append([]models.Conversation{conv}, conversations...)
					activeChat = &conversations[0]
				}
			}
		}

		if activeChat != nil {
			domainMessages, err := h.chatService.LoadRecentMessages(ctx, activeChat.ID, profileID, messageHistoryLimit)
			if err != nil {
				h.logger.Error("failed to load messages",
					zap.String("chat_id", activeChat.ID),
					zap.Error(err),
				)
			} else {
				messages, err = h.hydrateMessages(ctx, domainMessages, profileID)
				if err != nil {
					h.logger.Error("failed to hydrate messages",
						zap.String("chat_id", activeChat.ID),
						zap.Error(err),
					)
				}
				messages = annotateMessageGrouping(time.Local, messages)
			}
		}
	}

	theme := h.resolveTheme(r, user, profileID)

	data := models.ChatPageData{
		BaseData: models.BaseData{
			Title:        "Chat",
			ActiveNav:    "chat",
			Theme:        theme,
			User:         user,
			CSRFToken:    middleware.GetCSRFToken(r),
			AssetVersion: h.assetVersion,
		},
		Conversations:   conversations,
		ActiveChat:      activeChat,
		Messages:        messages,
		ViewerProfileID: profileID,
	}

	if err := templates.GetTemplates().Chat.ExecuteTemplate(w, "chat", data); err != nil {
		h.logger.Error("failed to render chat template", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSidebar returns the conversation-list fragment for HTMX refreshes.
func (h *ChatHandler) handleSidebar(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	inbox, err := h.chatService.ListInbox(r.Context(), profileID)
	if err != nil {
		h.logger.Error("failed to load inbox",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load chats", http.StatusInternalServerError)
		return
	}

	conversations, err := h.hydrateConversations(r.Context(), inbox, profileID)
	if err != nil {
		http.Error(w, "Failed to load chats", http.StatusInternalServerError)
		return
	}

	activeChatID := strings.TrimSpace(r.URL.Query().Get("active"))

	data := struct {
		Conversations  []models.Conversation
		ActiveChatID   string
		ViewerProfileID string
	}{
		Conversations:   conversations,
		ActiveChatID:    activeChatID,
		ViewerProfileID: profileID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.GetTemplates().ChatConversationList.ExecuteTemplate(w, "chat-sidebar-panel", data); err != nil {
		h.logger.Error("failed to render sidebar", zap.Error(err))
	}
}

// handleGetMessages returns a chronological list of messages for a chat.
// Supports keyset pagination via ?before=<rfc3339>&beforeId=<id>.
func (h *ChatHandler) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}
	chatID := r.PathValue("chatID")
	if chatID == "" {
		http.Error(w, "Missing chat id", http.StatusBadRequest)
		return
	}

	var (
		messages []domain.Message
		err      error
	)

	beforeRaw := r.URL.Query().Get("before")
	beforeID := r.URL.Query().Get("beforeId")
	if beforeRaw != "" && beforeID != "" {
		beforeTime, parseErr := time.Parse(time.RFC3339Nano, beforeRaw)
		if parseErr != nil {
			http.Error(w, "Invalid before timestamp", http.StatusBadRequest)
			return
		}
		messages, err = h.chatService.LoadHistoryBefore(r.Context(), chatID, profileID, beforeTime, beforeID, messageHistoryLimit)
	} else {
		messages, err = h.chatService.LoadRecentMessages(r.Context(), chatID, profileID, messageHistoryLimit)
	}
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	views, err := h.hydrateMessages(r.Context(), messages, profileID)
	if err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}
	views = annotateMessageGrouping(time.Local, views)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if beforeRaw != "" && beforeID != "" {
		hasMore := len(messages) >= messageHistoryLimit
		var oldest *models.Message
		if len(views) > 0 {
			oldest = &views[0]
		}
		payload := struct {
			ChatID        string
			HasMore       bool
			OldestMessage *models.Message
			Messages      []models.Message
		}{
			ChatID:        chatID,
			HasMore:       hasMore,
			OldestMessage: oldest,
			Messages:      views,
		}
		if err := templates.GetTemplates().ChatMessageList.ExecuteTemplate(w, "chat-messages-older", payload); err != nil {
			h.logger.Error("failed to render older messages fragment", zap.Error(err))
		}
		return
	}

	for i := range views {
		if err := templates.GetTemplates().ChatMessageRow.ExecuteTemplate(w, "chat-message-row", views[i]); err != nil {
			h.logger.Error("failed to render message row", zap.Error(err))
			return
		}
	}
}

// handleSendMessage persists a new message and returns the sender's row HTML.
// Other participants receive the same fragment over SSE via the chat hub.
func (h *ChatHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}
	chatID := r.PathValue("chatID")
	if chatID == "" {
		http.Error(w, "Missing chat id", http.StatusBadRequest)
		return
	}

	body := ""
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	body = r.FormValue("body")
	if body == "" {
		body = r.FormValue("content")
	}
	if len(body) > maxMessageBodyLen {
		body = body[:maxMessageBodyLen]
	}

	message, recipients, err := h.chatService.SendTextMessage(r.Context(), chatID, profileID, body)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	senderView, err := h.hydrateMessages(r.Context(), []domain.Message{*message}, profileID)
	if err != nil || len(senderView) == 0 {
		http.Error(w, "Failed to render message", http.StatusInternalServerError)
		return
	}
	markAsStandaloneRun(&senderView[0])

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.GetTemplates().ChatMessageRow.ExecuteTemplate(w, "chat-message-row", senderView[0]); err != nil {
		h.logger.Error("failed to render sender message row", zap.Error(err))
		return
	}

	if h.hub != nil && len(recipients) > 0 {
		go h.fanOutMessage(message, chatID, recipients)
	}
}

// handleMessagesSince replays missed messages after a reconnect. The client
// passes the last event ID it rendered so the server can resume from there.
func (h *ChatHandler) handleMessagesSince(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}
	chatID := r.PathValue("chatID")
	afterID := r.URL.Query().Get("eventId")

	messages, err := h.chatService.LoadMessagesSince(r.Context(), chatID, profileID, afterID, messageHistoryLimit)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	views, err := h.hydrateMessages(r.Context(), messages, profileID)
	if err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}
	views = annotateMessageGrouping(time.Local, views)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for i := range views {
		if err := templates.GetTemplates().ChatMessageRow.ExecuteTemplate(w, "chat-message-row", views[i]); err != nil {
			return
		}
	}
}

// handleSSE holds a single long-lived text/event-stream connection per tab.
// Events are keyed by the authenticated profile; each message carries the
// chatID inside its event name so the client can dispatch appropriately.
func (h *ChatHandler) handleSSE(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	flusher, canFlush := w.(http.Flusher)
	if !canFlush {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Disable the global WriteTimeout for this long-lived connection.
	// Without this, http.Server.WriteTimeout (default 15s) kills the
	// SSE stream before the first heartbeat (20s) can fire.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		h.logger.Warn("failed to clear write deadline for SSE", zap.Error(err))
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	fmt.Fprintf(w, "retry: %d\n\n", sseRetryMillis)
	flusher.Flush()

	events, unsubscribe := h.hub.Subscribe(profileID)
	defer unsubscribe()

	heartbeat := time.NewTicker(sseHeartbeatInterval)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				return
			}
			writeSSEEvent(w, event)
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w io.Writer, event chat.Event) {
	if event.Name != "" {
		fmt.Fprintf(w, "event: %s\n", event.Name)
	}
	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}
	for _, line := range strings.Split(event.Data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}

// handleNewChatModal renders the new-chat wizard. The same endpoint serves:
//   - no `step` or `step=circles`: initial modal shell with the circle picker.
//   - `step=members&circle={id}`: member-picker fragment swapped into
//     `#chat-new-step` once a circle is chosen.
func (h *ChatHandler) handleNewChatModal(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	step := strings.TrimSpace(r.URL.Query().Get("step"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	switch step {
	case "members":
		circleID := strings.TrimSpace(r.URL.Query().Get("circle"))
		if circleID == "" {
			h.writeChatError(w, r, http.StatusBadRequest, "Missing circle id.")
			return
		}
		circleView, members, err := h.loadCircleMembersForPicker(ctx, circleID, profileID)
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		data := struct {
			Circle  models.CircleOption
			Members []models.User
		}{
			Circle:  circleView,
			Members: members,
		}
		if err := templates.GetTemplates().ChatNewMembers.ExecuteTemplate(w, "chat-new-members", data); err != nil {
			h.logger.Error("failed to render new chat members step", zap.Error(err))
		}
	case "circles", "":
		circles, err := h.circleService.GetUserCircles(ctx, profileID)
		if err != nil {
			h.logger.Error("failed to list circles for new chat", zap.Error(err))
			http.Error(w, "Failed to load circles", http.StatusInternalServerError)
			return
		}
		options := make([]models.CircleOption, 0, len(circles))
		for _, c := range circles {
			if c.IsDeleted() {
				continue
			}
			options = append(options, models.CircleOption{ID: c.ID, Name: c.Name})
		}
		if step == "circles" {
			if err := templates.GetTemplates().ChatNewCircles.ExecuteTemplate(w, "chat-new-circles", options); err != nil {
				h.logger.Error("failed to render new chat circles step", zap.Error(err))
			}
			return
		}
		if err := templates.GetTemplates().ChatNewModal.ExecuteTemplate(w, "chat-new-modal", options); err != nil {
			h.logger.Error("failed to render new chat modal", zap.Error(err))
		}
	default:
		h.writeChatError(w, r, http.StatusBadRequest, "Unknown wizard step.")
	}
}

// loadCircleMembersForPicker returns the circle view model plus active
// members hydrated with profile info, excluding the viewer. It verifies
// that the viewer is still an active member before returning anything.
func (h *ChatHandler) loadCircleMembersForPicker(ctx context.Context, circleID, viewerProfileID string) (models.CircleOption, []models.User, error) {
	isMember, err := h.circleService.IsMember(ctx, circleID, viewerProfileID)
	if err != nil {
		return models.CircleOption{}, nil, fmt.Errorf("circle membership check: %w", err)
	}
	if !isMember {
		return models.CircleOption{}, nil, chat.ErrCircleAccessDenied
	}

	circleDomain, err := h.circleService.GetCircleByID(ctx, circleID)
	if err != nil {
		return models.CircleOption{}, nil, fmt.Errorf("load circle: %w", err)
	}
	if circleDomain == nil {
		return models.CircleOption{}, nil, chat.ErrCircleAccessDenied
	}

	memberships, err := h.circleService.GetCircleMembers(ctx, circleID, 500, 0)
	if err != nil {
		return models.CircleOption{}, nil, fmt.Errorf("load members: %w", err)
	}

	profileIDs := make([]string, 0, len(memberships)+1)
	seen := map[string]struct{}{}
	for _, m := range memberships {
		mm := m
		if !mm.IsActive() {
			continue
		}
		if mm.ProfileID == viewerProfileID {
			continue
		}
		if _, ok := seen[mm.ProfileID]; ok {
			continue
		}
		seen[mm.ProfileID] = struct{}{}
		profileIDs = append(profileIDs, mm.ProfileID)
	}
	if circleDomain.OwnerProfileID != "" && circleDomain.OwnerProfileID != viewerProfileID {
		if _, ok := seen[circleDomain.OwnerProfileID]; !ok {
			seen[circleDomain.OwnerProfileID] = struct{}{}
			profileIDs = append(profileIDs, circleDomain.OwnerProfileID)
		}
	}

	profiles, err := h.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return models.CircleOption{}, nil, fmt.Errorf("hydrate profiles: %w", err)
	}
	byID := make(map[string]domain.Profile, len(profiles))
	for _, p := range profiles {
		byID[p.ID] = p
	}

	members := make([]models.User, 0, len(profileIDs))
	for _, pid := range profileIDs {
		p, ok := byID[pid]
		if !ok {
			continue
		}
		members = append(members, models.User{
			ID:     p.ID,
			Handle: "@" + p.Handle,
			Name:   profileDisplayName(p),
			Avatar: p.AvatarURL,
		})
	}

	return models.CircleOption{ID: circleDomain.ID, Name: circleDomain.Name}, members, nil
}

// handleCreateDM finds or creates a 1:1 chat and returns an HX-Redirect.
// The caller must supply `circle_id` identifying an active circle both
// participants belong to; this enforces the shared-circle messaging policy.
// Accepts either `other_profile_id` (canonical) or `other_handle` (used by
// the legacy modal entry point).
func (h *ChatHandler) handleCreateDM(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	circleID := strings.TrimSpace(formOrJSONValue(r, "circle_id"))
	if circleID == "" {
		h.writeChatError(w, r, http.StatusBadRequest, "Pick a circle first.")
		return
	}

	other := formOrJSONValue(r, "other_profile_id")
	if other == "" {
		handle := strings.TrimSpace(strings.TrimPrefix(formOrJSONValue(r, "other_handle"), "@"))
		if handle == "" {
			h.writeChatError(w, r, http.StatusBadRequest, "Choose a member to message.")
			return
		}
		profile, err := h.profileService.GetByHandle(r.Context(), handle)
		if err != nil {
			h.logger.Warn("failed to resolve handle for dm",
				zap.String("handle", handle),
				zap.Error(err),
			)
			h.writeChatError(w, r, http.StatusInternalServerError, "Couldn't look up that profile.")
			return
		}
		if profile == nil {
			h.writeChatError(w, r, http.StatusNotFound, "No profile found with that handle.")
			return
		}
		if profile.ID == profileID {
			h.writeChatError(w, r, http.StatusBadRequest, "You can't start a DM with yourself.")
			return
		}
		other = profile.ID
	}

	createdChat, err := h.chatService.FindOrCreateDMInCircle(r.Context(), profileID, other, circleID)
	if err != nil {
		if errors.Is(err, chat.ErrCircleAccessDenied) {
			h.writeChatError(w, r, http.StatusForbidden, "You can only message members of a circle you both belong to.")
			return
		}
		h.writeServiceError(w, err)
		return
	}

	redirect := "/chat/" + createdChat.ID
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirect)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// writeChatError renders an inline error for HTMX requests (swapped into the
// modal's error slot) and falls back to plain text for non-HTMX callers.
func (h *ChatHandler) writeChatError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, `<p class="chat-form-error" role="alert">%s</p>`, template.HTMLEscapeString(msg))
		return
	}
	http.Error(w, msg, status)
}

// handleCreateGroup creates a new group chat with the supplied participants.
// Requires `circle_id`; every participant (including the creator) must be
// an active member of that circle. The circle only gates creation, the
// resulting chat is not stored as circle-scoped.
func (h *ChatHandler) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}

	var (
		name         string
		circleID     string
		participants []string
	)

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		payload := struct {
			Name                  string   `json:"name"`
			CircleID              string   `json:"circle_id"`
			ParticipantProfileIDs []string `json:"participant_profile_ids"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		name = payload.Name
		circleID = payload.CircleID
		participants = payload.ParticipantProfileIDs
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}
		name = r.FormValue("name")
		circleID = r.FormValue("circle_id")
		participants = r.Form["participant_profile_ids"]
	}

	circleID = strings.TrimSpace(circleID)
	if circleID == "" {
		h.writeChatError(w, r, http.StatusBadRequest, "Pick a circle first.")
		return
	}

	createdChat, err := h.chatService.CreateGroupChatInCircle(r.Context(), profileID, name, participants, circleID)
	if err != nil {
		if errors.Is(err, chat.ErrCircleAccessDenied) {
			h.writeChatError(w, r, http.StatusForbidden, "Every participant must be an active member of the selected circle.")
			return
		}
		if errors.Is(err, chat.ErrInvalidParticipants) {
			h.writeChatError(w, r, http.StatusBadRequest, "Select at least one other member.")
			return
		}
		h.writeServiceError(w, err)
		return
	}

	redirect := "/chat/" + createdChat.ID
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirect)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// handleCreateCircleChat creates a circle-scoped chat for active members.
func (h *ChatHandler) handleCreateCircleChat(w http.ResponseWriter, r *http.Request) {
	_, profileID, ok := h.requireProfile(w, r)
	if !ok {
		return
	}
	circleID := r.PathValue("circleID")
	if circleID == "" {
		http.Error(w, "Missing circle id", http.StatusBadRequest)
		return
	}

	name := formOrJSONValue(r, "name")
	chat, err := h.chatService.CreateCircleChat(r.Context(), profileID, circleID, name)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	redirect := "/chat/" + chat.ID
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirect)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// fanOutMessage renders the recipient-side row once per unique ID and
// publishes the same fragment to every subscribed tab via the hub.
func (h *ChatHandler) fanOutMessage(message *domain.Message, chatID string, recipients []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sender, err := h.profileService.GetByID(ctx, message.SenderProfileID)
	if err != nil || sender == nil {
		h.logger.Warn("failed to load sender for SSE fan-out",
			zap.String("sender_profile_id", message.SenderProfileID),
			zap.Error(err),
		)
		return
	}

	var buf bytes.Buffer
	view := messageToView(*message, *sender, false, chatID)
	markAsStandaloneRun(&view)
	if err := templates.GetTemplates().ChatMessageRow.ExecuteTemplate(&buf, "chat-message-row", view); err != nil {
		h.logger.Error("failed to render sse message row", zap.Error(err))
		return
	}

	event := chat.Event{
		Name:   "msg-" + chatID,
		ID:     message.ID,
		Data:   sanitizeSSEData(buf.String()),
		ChatID: chatID,
	}
	h.hub.PublishAll(recipients, event)
}

// sanitizeSSEData collapses newlines in a single HTML fragment so it renders
// cleanly through `data:` framing (multiple `data:` lines concatenate with
// newlines which browsers honor, but we want a compact single-line event).
func sanitizeSSEData(html string) string {
	trimmed := strings.ReplaceAll(html, "\r", "")
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	return strings.TrimSpace(trimmed)
}

func (h *ChatHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chat.ErrChatNotFound):
		http.Error(w, "Chat not found", http.StatusNotFound)
	case errors.Is(err, chat.ErrNotAuthorized), errors.Is(err, chat.ErrCircleAccessDenied):
		http.Error(w, "Forbidden", http.StatusForbidden)
	case errors.Is(err, chat.ErrBlocked):
		http.Error(w, "Conversation not allowed", http.StatusConflict)
	case errors.Is(err, chat.ErrEmptyMessage), errors.Is(err, chat.ErrInvalidParticipants), errors.Is(err, chat.ErrCircleRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		h.logger.Error("chat service error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *ChatHandler) requireProfile(w http.ResponseWriter, r *http.Request) (*domain.User, string, bool) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, "", false
	}
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return nil, "", false
	}
	return user, *session.ActiveProfileID, true
}

func (h *ChatHandler) resolveTheme(r *http.Request, user *domain.User, profileID string) models.ThemeSettings {
	fallback := models.ThemeSettings{Mode: "system", Radius: "6"}
	if h.prefsService == nil || user == nil {
		return fallback
	}
	session := auth.GetSession(r.Context())
	var activeProfileID *string
	if session != nil && session.ActiveProfileID != nil {
		activeProfileID = session.ActiveProfileID
	} else {
		activeProfileID = &profileID
	}
	theme, err := h.prefsService.GetEffectiveTheme(r.Context(), user.ID, activeProfileID)
	if err != nil || theme == nil {
		return fallback
	}
	return models.ThemeSettings{
		Mode:   theme.Mode,
		Radius: theme.Radius,
	}
}

func extractActiveChatID(path string) string {
	if !strings.HasPrefix(path, "/chat/") {
		return ""
	}
	trimmed := strings.TrimPrefix(path, "/chat/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return trimmed[:idx]
	}
	return trimmed
}

func formOrJSONValue(r *http.Request, key string) string {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		payload := map[string]string{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		return strings.TrimSpace(payload[key])
	}
	if err := r.ParseForm(); err == nil {
		return strings.TrimSpace(r.FormValue(key))
	}
	return ""
}

// hydrateConversations turns inbox entries into view models, resolving
// profile names/avatars for DM counterparts and groups.
func (h *ChatHandler) hydrateConversations(ctx context.Context, inbox []chat.InboxEntry, viewerProfileID string) ([]models.Conversation, error) {
	if len(inbox) == 0 {
		return []models.Conversation{}, nil
	}

	profileIDs := make(map[string]struct{})
	chatParticipants := make(map[string][]domain.ChatParticipant, len(inbox))
	for _, entry := range inbox {
		participants, err := h.chatService.Participants(ctx, entry.Chat.ID)
		if err != nil {
			return nil, fmt.Errorf("participants for %s: %w", entry.Chat.ID, err)
		}
		chatParticipants[entry.Chat.ID] = participants
		for _, p := range participants {
			profileIDs[p.ProfileID] = struct{}{}
		}
		if entry.LastMessage != nil {
			profileIDs[entry.LastMessage.SenderProfileID] = struct{}{}
		}
	}

	profiles, err := h.loadProfiles(ctx, profileIDs)
	if err != nil {
		return nil, err
	}

	conversations := make([]models.Conversation, 0, len(inbox))
	for _, entry := range inbox {
		conv := buildConversationView(entry.Chat, chatParticipants[entry.Chat.ID], entry.LastMessage, entry.UnreadCount, viewerProfileID, profiles)
		conversations = append(conversations, conv)
	}
	return conversations, nil
}

func (h *ChatHandler) buildConversation(ctx context.Context, chat *domain.Chat, last *domain.Message, unread int, viewerProfileID string) (models.Conversation, error) {
	participants, err := h.chatService.Participants(ctx, chat.ID)
	if err != nil {
		return models.Conversation{}, err
	}
	profileIDs := make(map[string]struct{})
	for _, p := range participants {
		profileIDs[p.ProfileID] = struct{}{}
	}
	if last != nil {
		profileIDs[last.SenderProfileID] = struct{}{}
	}
	profiles, err := h.loadProfiles(ctx, profileIDs)
	if err != nil {
		return models.Conversation{}, err
	}
	return buildConversationView(*chat, participants, last, unread, viewerProfileID, profiles), nil
}

func buildConversationView(
	chat domain.Chat,
	participants []domain.ChatParticipant,
	last *domain.Message,
	unread int,
	viewerProfileID string,
	profiles map[string]domain.Profile,
) models.Conversation {
	conv := models.Conversation{
		ID:          chat.ID,
		IsGroup:     chat.IsGroup() || chat.IsCircleScoped(),
		UnreadCount: unread,
	}
	if chat.IsCircleScoped() {
		conv.IsCircleScoped = true
		if chat.CircleID != nil {
			conv.CircleID = *chat.CircleID
		}
	}

	name := chat.Name
	avatar := ""
	var otherParticipants []models.User

	for _, p := range participants {
		prof, ok := profiles[p.ProfileID]
		if !ok {
			continue
		}
		displayName := profileDisplayName(prof)
		if p.ProfileID != viewerProfileID {
			otherParticipants = append(otherParticipants, models.User{
				ID:     prof.ID,
				Handle: "@" + prof.Handle,
				Name:   displayName,
				Avatar: prof.AvatarURL,
			})
		}
	}

	if chat.IsDirect() {
		if len(otherParticipants) > 0 {
			name = otherParticipants[0].Name
			avatar = otherParticipants[0].Avatar
		}
	} else {
		conv.Participants = otherParticipants
		if name == "" {
			name = "Group chat"
		}
	}

	conv.Name = name
	conv.Avatar = avatar

	if last != nil {
		senderName := "Unknown"
		if prof, ok := profiles[last.SenderProfileID]; ok {
			senderName = profileDisplayName(prof)
		}
		preview := last.Content
		if preview == "" {
			preview = "(empty)"
		}
		if !chat.IsDirect() && last.SenderProfileID != viewerProfileID {
			preview = senderName + ": " + preview
		}
		conv.LastMessage = preview
		conv.LastTime = relativeTime(last.CreatedAt)
	}
	return conv
}

func (h *ChatHandler) hydrateMessages(ctx context.Context, messages []domain.Message, viewerProfileID string) ([]models.Message, error) {
	if len(messages) == 0 {
		return []models.Message{}, nil
	}
	profileIDs := make(map[string]struct{})
	for _, m := range messages {
		profileIDs[m.SenderProfileID] = struct{}{}
	}
	profiles, err := h.loadProfiles(ctx, profileIDs)
	if err != nil {
		return nil, err
	}
	views := make([]models.Message, 0, len(messages))
	for _, m := range messages {
		sender := profiles[m.SenderProfileID]
		views = append(views, messageToView(m, sender, m.SenderProfileID == viewerProfileID, m.ChatID))
	}
	return views, nil
}

func (h *ChatHandler) loadProfiles(ctx context.Context, ids map[string]struct{}) (map[string]domain.Profile, error) {
	if len(ids) == 0 {
		return map[string]domain.Profile{}, nil
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	profiles, err := h.profileService.GetByIDs(ctx, idList)
	if err != nil {
		return nil, fmt.Errorf("load profiles: %w", err)
	}
	out := make(map[string]domain.Profile, len(profiles))
	for _, p := range profiles {
		out[p.ID] = p
	}
	return out, nil
}

func messageToView(m domain.Message, sender domain.Profile, isOwn bool, chatID string) models.Message {
	return models.Message{
		ID:        m.ID,
		Content:   m.Content,
		Timestamp: m.CreatedAt.Format("3:04 PM"),
		CreatedAt: m.CreatedAt,
		Sender: models.User{
			ID:     sender.ID,
			Handle: "@" + sender.Handle,
			Name:   profileDisplayName(sender),
			Avatar: sender.AvatarURL,
		},
		IsOwn:  isOwn,
		IsRead: true,
		Type:   messageTypeOrText(m.MessageType),
	}
}

// annotateMessageGrouping stamps BucketLabel / IsRunStart / IsRunEnd on the
// view-model slice so the template can render date dividers and collapse
// consecutive messages from the same sender sent within runWindow.
// The slice must be in chronological order (oldest first).
func annotateMessageGrouping(loc *time.Location, msgs []models.Message) []models.Message {
	if len(msgs) == 0 {
		return msgs
	}
	if loc == nil {
		loc = time.Local
	}
	const runWindow = 5 * time.Minute
	now := time.Now().In(loc)
	for i := range msgs {
		local := msgs[i].CreatedAt.In(loc)
		newBucket := i == 0
		if i > 0 {
			prevLocal := msgs[i-1].CreatedAt.In(loc)
			if !sameLocalDay(prevLocal, local) {
				newBucket = true
			}
		}
		if newBucket {
			msgs[i].BucketLabel = bucketLabelFor(local, now)
		}
		newRun := i == 0 || newBucket || msgs[i-1].Sender.ID != msgs[i].Sender.ID ||
			local.Sub(msgs[i-1].CreatedAt.In(loc)) > runWindow
		msgs[i].IsRunStart = newRun
		if newRun && i > 0 {
			msgs[i-1].IsRunEnd = true
		}
	}
	msgs[len(msgs)-1].IsRunEnd = true
	return msgs
}

// markAsStandaloneRun flags a single appended message so the template treats
// it as a run of one (shows avatar/name + timestamp) without prepending a
// bucket divider. Used for POST and SSE fan-out paths where we lack enough
// context to decide if a new date divider is warranted.
func markAsStandaloneRun(m *models.Message) {
	m.IsRunStart = true
	m.IsRunEnd = true
	m.BucketLabel = ""
}

func sameLocalDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func bucketLabelFor(local, now time.Time) string {
	if sameLocalDay(local, now) {
		return "Today"
	}
	if sameLocalDay(local, now.AddDate(0, 0, -1)) {
		return "Yesterday"
	}
	cutoff := now.AddDate(0, 0, -6)
	if local.After(cutoff) {
		return local.Format("Monday")
	}
	return local.Format("Jan 2, 2006")
}

func messageTypeOrText(t string) string {
	if t == "" {
		return domain.MessageTypeText
	}
	return t
}

func profileDisplayName(p domain.Profile) string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	if p.Name != "" {
		return p.Name
	}
	if p.Handle != "" {
		return p.Handle
	}
	return "Unknown"
}

func relativeTime(t time.Time) string {
	delta := time.Since(t)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	case delta < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

// Ensure html/template import stays satisfied if future helpers render directly.
var _ = template.HTML("")
