package middleware

import (
	"context"
	"net/http"
)

// Context keys for HTMX request data
type contextKey string

const (
	htmxRequestKey        contextKey = "htmx_request"
	htmxBoostedKey        contextKey = "htmx_boosted"
	htmxCurrentURLKey     contextKey = "htmx_current_url"
	htmxPromptKey         contextKey = "htmx_prompt"
	htmxTargetKey         contextKey = "htmx_target"
	htmxTriggerNameKey    contextKey = "htmx_trigger_name"
	htmxTriggerKey        contextKey = "htmx_trigger"
	htmxHistoryRestoreKey contextKey = "htmx_history_restore"
)

// HTMXMiddleware adds HTMX request context to the request
func HTMXMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			ctx = context.WithValue(ctx, htmxRequestKey, true)

			// Extract HTMX-specific headers
			if boosted := r.Header.Get("HX-Boosted"); boosted != "" {
				ctx = context.WithValue(ctx, htmxBoostedKey, boosted == "true")
			}

			if currentURL := r.Header.Get("HX-Current-URL"); currentURL != "" {
				ctx = context.WithValue(ctx, htmxCurrentURLKey, currentURL)
			}

			if prompt := r.Header.Get("HX-Prompt"); prompt != "" {
				ctx = context.WithValue(ctx, htmxPromptKey, prompt)
			}

			if target := r.Header.Get("HX-Target"); target != "" {
				ctx = context.WithValue(ctx, htmxTargetKey, target)
			}

			if triggerName := r.Header.Get("HX-Trigger-Name"); triggerName != "" {
				ctx = context.WithValue(ctx, htmxTriggerNameKey, triggerName)
			}

			if trigger := r.Header.Get("HX-Trigger"); trigger != "" {
				ctx = context.WithValue(ctx, htmxTriggerKey, trigger)
			}

			if historyRestore := r.Header.Get("HX-History-Restore-Request"); historyRestore == "true" {
				ctx = context.WithValue(ctx, htmxHistoryRestoreKey, true)
			}
		} else {
			ctx = context.WithValue(ctx, htmxRequestKey, false)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsHTMXRequest checks if the current request is an HTMX request
func IsHTMXRequest(r *http.Request) bool {
	if val, ok := r.Context().Value(htmxRequestKey).(bool); ok {
		return val
	}
	return false
}

// IsHTMXBoosted checks if the request is an HTMX boosted request
func IsHTMXBoosted(r *http.Request) bool {
	if val, ok := r.Context().Value(htmxBoostedKey).(bool); ok {
		return val
	}
	return false
}

// GetHTMXCurrentURL returns the current URL from HTMX headers
func GetHTMXCurrentURL(r *http.Request) string {
	if val, ok := r.Context().Value(htmxCurrentURLKey).(string); ok {
		return val
	}
	return ""
}

// GetHTMXPrompt returns the prompt response from HTMX
func GetHTMXPrompt(r *http.Request) string {
	if val, ok := r.Context().Value(htmxPromptKey).(string); ok {
		return val
	}
	return ""
}

// GetHTMXTarget returns the target element ID from HTMX
func GetHTMXTarget(r *http.Request) string {
	if val, ok := r.Context().Value(htmxTargetKey).(string); ok {
		return val
	}
	return ""
}

// GetHTMXTriggerName returns the name of the triggered element
func GetHTMXTriggerName(r *http.Request) string {
	if val, ok := r.Context().Value(htmxTriggerNameKey).(string); ok {
		return val
	}
	return ""
}

// GetHTMXTrigger returns the ID of the triggered element
func GetHTMXTrigger(r *http.Request) string {
	if val, ok := r.Context().Value(htmxTriggerKey).(string); ok {
		return val
	}
	return ""
}

// IsHTMXHistoryRestore checks if this is a history restore request
func IsHTMXHistoryRestore(r *http.Request) bool {
	if val, ok := r.Context().Value(htmxHistoryRestoreKey).(bool); ok {
		return val
	}
	return false
}

// SetHTMXTrigger sets a client-side event trigger
func SetHTMXTrigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}

// SetHTMXTriggerAfterSettle sets a client-side event trigger after settle
func SetHTMXTriggerAfterSettle(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger-After-Settle", event)
}

// SetHTMXTriggerAfterSwap sets a client-side event trigger after swap
func SetHTMXTriggerAfterSwap(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger-After-Swap", event)
}

// SetHTMXRedirect forces a client-side redirect
func SetHTMXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

// SetHTMXRefresh forces a client-side page refresh
func SetHTMXRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
}

// SetHTMXLocation allows client-side navigation without full page reload
func SetHTMXLocation(w http.ResponseWriter, path string) {
	w.Header().Set("HX-Location", path)
}

// SetHTMXPushURL updates the browser URL
func SetHTMXPushURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-Url", url)
}

// SetHTMXReplaceURL replaces the current URL in browser history
func SetHTMXReplaceURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Replace-Url", url)
}

// SetHTMXReswap allows you to specify how the response will be swapped
func SetHTMXReswap(w http.ResponseWriter, swapType string) {
	w.Header().Set("HX-Reswap", swapType)
}

// SetHTMXRetarget allows you to target a different element
func SetHTMXRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// SetHTMXReselect allows you to choose which part of the response to swap
func SetHTMXReselect(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Reselect", selector)
}
