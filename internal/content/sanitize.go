package content

import (
	"html"
	"regexp"
	"strings"

	"circles.diy/internal/domain"
)

// SanitizeInput sanitizes user input based on body format
func SanitizeInput(body, bodyFormat string) string {
	switch bodyFormat {
	case domain.BodyFormatHTML:
		return SanitizeHTML(body)
	case domain.BodyFormatMarkdown:
		return SanitizeMarkdown(body)
	case domain.BodyFormatPlaintext:
		return SanitizePlaintext(body)
	default:
		return SanitizePlaintext(body)
	}
}

// SanitizePlaintext sanitizes plaintext input by escaping HTML
func SanitizePlaintext(text string) string {
	// Escape HTML entities
	text = html.EscapeString(text)

	// Trim excessive whitespace
	text = strings.TrimSpace(text)

	return text
}

// SanitizeMarkdown sanitizes markdown input
// We allow most markdown syntax but prevent XSS through script tags and dangerous HTML
func SanitizeMarkdown(text string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	text = scriptRegex.ReplaceAllString(text, "")

	// Remove inline script handlers (onclick, onload, etc.)
	handlerRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	text = handlerRegex.ReplaceAllString(text, "")

	// Remove javascript: protocol
	jsProtocolRegex := regexp.MustCompile(`(?i)javascript:`)
	text = jsProtocolRegex.ReplaceAllString(text, "")

	// Remove data: protocol (can be used for XSS)
	dataProtocolRegex := regexp.MustCompile(`(?i)data:`)
	text = dataProtocolRegex.ReplaceAllString(text, "")

	// Remove iframe tags
	iframeRegex := regexp.MustCompile(`(?i)<iframe[\s\S]*?</iframe>`)
	text = iframeRegex.ReplaceAllString(text, "")

	// Remove object/embed tags
	objectRegex := regexp.MustCompile(`(?i)<(object|embed)[\s\S]*?</(object|embed)>`)
	text = objectRegex.ReplaceAllString(text, "")

	return text
}

// SanitizeHTML performs strict HTML sanitization
// For now, we escape all HTML. In future, we can use a library like bluemonday
// to allow specific safe HTML tags
func SanitizeHTML(text string) string {
	// For MVP, we'll escape all HTML to prevent XSS
	// In production, integrate a proper HTML sanitization library
	return html.EscapeString(text)
}

// TruncateText truncates text to a maximum length with ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// Find the last space before maxLength to avoid cutting words
	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")

	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// ValidateBodyFormat checks if a body format is valid
func ValidateBodyFormat(bodyFormat string) bool {
	switch bodyFormat {
	case domain.BodyFormatMarkdown, domain.BodyFormatPlaintext, domain.BodyFormatHTML:
		return true
	default:
		return false
	}
}

// ValidateVisibility checks if a visibility setting is valid
func ValidateVisibility(visibility string) bool {
	switch visibility {
	case domain.PostVisibilityPublic, domain.PostVisibilityMembersOnly, domain.PostVisibilityPrivate:
		return true
	default:
		return false
	}
}

// ExtractPreview extracts a preview from markdown or plaintext content
// Removes markdown formatting to create a clean preview
func ExtractPreview(body, bodyFormat string, maxLength int) string {
	preview := body

	if bodyFormat == domain.BodyFormatMarkdown {
		// Remove markdown headers
		headerRegex := regexp.MustCompile(`(?m)^#+\s*`)
		preview = headerRegex.ReplaceAllString(preview, "")

		// Remove markdown links but keep link text: [text](url) -> text
		linkRegex := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
		preview = linkRegex.ReplaceAllString(preview, "$1")

		// Remove markdown images: ![alt](url) -> ""
		imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\([^\)]+\)`)
		preview = imageRegex.ReplaceAllString(preview, "")

		// Remove markdown bold/italic: **text** or *text* -> text
		boldItalicRegex := regexp.MustCompile(`\*+([^\*]+)\*+`)
		preview = boldItalicRegex.ReplaceAllString(preview, "$1")

		// Remove markdown code blocks: ```code``` -> ""
		codeBlockRegex := regexp.MustCompile("```[\\s\\S]*?```")
		preview = codeBlockRegex.ReplaceAllString(preview, "")

		// Remove inline code: `code` -> code
		inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
		preview = inlineCodeRegex.ReplaceAllString(preview, "$1")

		// Remove markdown blockquotes
		blockquoteRegex := regexp.MustCompile(`(?m)^>\s*`)
		preview = blockquoteRegex.ReplaceAllString(preview, "")

		// Remove horizontal rules
		hrRegex := regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
		preview = hrRegex.ReplaceAllString(preview, "")
	}

	// Collapse multiple newlines into single space
	newlineRegex := regexp.MustCompile(`\s+`)
	preview = newlineRegex.ReplaceAllString(preview, " ")

	// Trim whitespace
	preview = strings.TrimSpace(preview)

	// Truncate to max length
	return TruncateText(preview, maxLength)
}

// StripHTML removes all HTML tags from text
func StripHTML(text string) string {
	// Remove all HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	text = htmlRegex.ReplaceAllString(text, "")

	// Decode HTML entities
	text = html.UnescapeString(text)

	return strings.TrimSpace(text)
}

// IsEmptyContent checks if content is effectively empty after sanitization
func IsEmptyContent(body string) bool {
	// Remove whitespace
	trimmed := strings.TrimSpace(body)

	// Check if empty
	if trimmed == "" {
		return true
	}

	// Check if only contains newlines, spaces, tabs
	whitespaceOnly := regexp.MustCompile(`^[\s\n\r\t]*$`)
	return whitespaceOnly.MatchString(trimmed)
}
