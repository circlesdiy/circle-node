package templates

import (
	"html/template"
	"log"
)

type Templates struct {
	Dashboard       *template.Template
	ProfilePublic   *template.Template
	ProfileInternal *template.Template
	Circles         *template.Template
	Chat            *template.Template
	Gather          *template.Template
	Marketplace     *template.Template
}

var templates *Templates

func InitTemplates() error {
	templates = &Templates{}

	var err error

	// Parse all templates using embedded file system
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

	templates.Marketplace, err = parseTemplateFromEmbedded("marketplace", "html/pages/marketplace.html")
	if err != nil {
		return err
	}

	log.Println("All templates loaded successfully from embedded filesystem")
	return nil
}

// GetTemplates returns the global templates instance
func GetTemplates() *Templates {
	return templates
}