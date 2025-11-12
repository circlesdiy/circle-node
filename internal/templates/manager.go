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
	Circles         *template.Template
	Chat            *template.Template
	Gather          *template.Template
	EventCreate     *template.Template
	EventDetail     *template.Template
	Marketplace     *template.Template

	// Components
	Modal       *template.Template
	Toast       *template.Template
	CircleCard  *template.Template
	CircleForm  *template.Template
	EmptyState  *template.Template
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

	templates.Circles, err = parseTemplateFromEmbedded("circles", "html/pages/circles.html")
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

	return nil
}

// GetTemplates returns the global templates instance
func GetTemplates() *Templates {
	return templates
}
