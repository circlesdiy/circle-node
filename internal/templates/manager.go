package templates

import (
	"html/template"
)

type Templates struct {
	// Auth pages
	AuthLogin    *template.Template
	AuthRegister *template.Template

	// App pages
	Dashboard       *template.Template
	ProfilePublic   *template.Template
	ProfileInternal *template.Template
	ProfileEdit     *template.Template
	Circles         *template.Template
	CircleDetail    *template.Template
	Chat            *template.Template
	Gather          *template.Template
	EventCreate     *template.Template
	EventDetail     *template.Template
	Marketplace     *template.Template

	// Components
	Modal               *template.Template
	Toast               *template.Template
	CircleCard          *template.Template
	CircleForm          *template.Template
	CircleTabChat       *template.Template
	CircleTabFiles      *template.Template
	CircleTabGatherings *template.Template
	CircleTabMembers    *template.Template
	CircleTabSettings   *template.Template
	InviteMemberForm    *template.Template
	EmptyState          *template.Template
	PostComposer        *template.Template
	PostCard            *template.Template
	CommentCard         *template.Template
	CommentList         *template.Template
	CommentForm         *template.Template

	// Chat fragments (returned for HTMX/SSE updates)
	ChatConversationList *template.Template
	ChatConversationRow  *template.Template
	ChatMessageList      *template.Template
	ChatMessageRow       *template.Template
	ChatNewModal         *template.Template
	ChatNewCircles       *template.Template
	ChatNewMembers       *template.Template
}

var templates *Templates

func InitTemplates() error {
	templates = &Templates{}

	var err error

	// Parse auth templates
	templates.AuthLogin, err = parseTemplateFromEmbedded("auth-login", "html/pages/auth-login.html")
	if err != nil {
		return err
	}

	templates.AuthRegister, err = parseTemplateFromEmbedded("auth-register", "html/pages/auth-register.html")
	if err != nil {
		return err
	}

	// Parse all app templates using embedded file system
	templates.Dashboard, err = parseTemplateFromEmbedded("dashboard", "html/pages/dashboard.html")
	if err != nil {
		return err
	}

	templates.ProfilePublic, err = parseTemplateFromEmbedded("profile", "html/pages/profile-public.html")
	if err != nil {
		return err
	}

	templates.ProfileInternal, err = parseTemplateFromEmbedded("profile-internal", "html/pages/profile-internal.html")
	if err != nil {
		return err
	}

	templates.ProfileEdit, err = parseTemplateFromEmbedded("profile-edit", "html/pages/profile-edit.html")
	if err != nil {
		return err
	}

	templates.Circles, err = parseTemplateFromEmbedded("circles", "html/pages/circles.html")
	if err != nil {
		return err
	}

	templates.CircleDetail, err = parseTemplateFromEmbedded("circle-detail", "html/pages/circle-detail.html")
	if err != nil {
		return err
	}

	templates.Chat, err = parseTemplateFromEmbedded("chat", "html/pages/chat.html")
	if err != nil {
		return err
	}

	templates.Gather, err = parseTemplateFromEmbedded("gather", "html/pages/gather.html")
	if err != nil {
		return err
	}

	templates.EventCreate, err = parseTemplateFromEmbedded("event-create", "html/pages/event-create.html")
	if err != nil {
		return err
	}

	templates.EventDetail, err = parseTemplateFromEmbedded("event-detail", "html/pages/event-detail.html")
	if err != nil {
		return err
	}

	templates.Marketplace, err = parseTemplateFromEmbedded("marketplace", "html/pages/marketplace.html")
	if err != nil {
		return err
	}

	// Parse component templates
	templates.Modal, err = parseTemplateFromEmbedded("modal", "html/components/modal.html")
	if err != nil {
		return err
	}

	templates.Toast, err = parseTemplateFromEmbedded("toast", "html/components/toast.html")
	if err != nil {
		return err
	}

	templates.CircleCard, err = parseTemplateFromEmbedded("circle-card", "html/components/circle-card.html")
	if err != nil {
		return err
	}

	templates.CircleForm, err = parseTemplateFromEmbedded("circle-form", "html/components/circle-form.html")
	if err != nil {
		return err
	}

	templates.CircleTabChat, err = parseTemplateFromEmbedded("circle-tab-chat", "html/components/circle-tab-chat.html")
	if err != nil {
		return err
	}

	templates.CircleTabFiles, err = parseTemplateFromEmbedded("circle-tab-files", "html/components/circle-tab-files.html")
	if err != nil {
		return err
	}

	templates.CircleTabGatherings, err = parseTemplateFromEmbedded("circle-tab-gatherings", "html/components/circle-tab-gatherings.html")
	if err != nil {
		return err
	}

	templates.CircleTabMembers, err = parseTemplateFromEmbedded("circle-tab-members", "html/components/circle-tab-members.html")
	if err != nil {
		return err
	}

	templates.CircleTabSettings, err = parseTemplateFromEmbedded("circle-tab-settings", "html/components/circle-tab-settings.html")
	if err != nil {
		return err
	}

	templates.InviteMemberForm, err = parseTemplateFromEmbedded("invite-member-form", "html/components/invite-member-form.html")
	if err != nil {
		return err
	}

	templates.PostComposer, err = parseTemplateFromEmbedded("post-composer", "html/components/post-composer.html")
	if err != nil {
		return err
	}

	templates.PostCard, err = parseTemplateFromEmbedded("post-card", "html/components/post-card.html")
	if err != nil {
		return err
	}

	templates.CommentCard, err = parseTemplateFromEmbedded("comment-card", "html/components/comment-card.html")
	if err != nil {
		return err
	}

	templates.CommentList, err = parseTemplateFromEmbedded("comment-list", "html/components/comment-list.html")
	if err != nil {
		return err
	}

	templates.CommentForm, err = parseTemplateFromEmbedded("comment-form", "html/components/comment-form.html")
	if err != nil {
		return err
	}

	templates.ChatConversationList, err = parseTemplateFromEmbedded("chat-conversation-list", "")
	if err != nil {
		return err
	}

	templates.ChatConversationRow, err = parseTemplateFromEmbedded("chat-conversation-row", "")
	if err != nil {
		return err
	}

	templates.ChatMessageList, err = parseTemplateFromEmbedded("chat-message-list", "")
	if err != nil {
		return err
	}

	templates.ChatMessageRow, err = parseTemplateFromEmbedded("chat-message-row", "")
	if err != nil {
		return err
	}

	templates.ChatNewModal, err = parseTemplateFromEmbedded("chat-new-modal", "")
	if err != nil {
		return err
	}

	templates.ChatNewCircles, err = parseTemplateFromEmbedded("chat-new-circles", "")
	if err != nil {
		return err
	}

	templates.ChatNewMembers, err = parseTemplateFromEmbedded("chat-new-members", "")
	if err != nil {
		return err
	}

	return nil
}

// GetTemplates returns the global templates instance
func GetTemplates() *Templates {
	return templates
}
