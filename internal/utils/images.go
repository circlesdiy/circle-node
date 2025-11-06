package utils

const (
	// AvatarFallbackURL is the default avatar image for users and profiles
	AvatarFallbackURL = "https://placehold.co/32x32/FF8E7A/31343C"

	// BannerFallbackURL is the default banner image for profiles and circles
	BannerFallbackURL = "https://placehold.co/800x400/7ADEFF/31343C"
)

// GetAvatarURL returns the avatar URL or fallback if empty
func GetAvatarURL(avatarURL string) string {
	if avatarURL == "" {
		return AvatarFallbackURL
	}
	return avatarURL
}

// GetBannerURL returns the banner URL or fallback if empty
func GetBannerURL(bannerURL string) string {
	if bannerURL == "" {
		return BannerFallbackURL
	}
	return bannerURL
}

// GetThumbnailURL returns the thumbnail URL or fallback to avatar fallback if empty
func GetThumbnailURL(thumbnailURL string) string {
	if thumbnailURL == "" {
		return AvatarFallbackURL
	}
	return thumbnailURL
}
