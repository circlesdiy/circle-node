package templates

import (
	"circles.diy/internal/models"
	"circles.diy/internal/utils"
)

func GetMockChatData() models.ChatPageData {
	return models.ChatPageData{
		BaseData: models.BaseData{
			Title:     "Chat",
			ActiveNav: "chat",
			Theme: models.ThemeSettings{
				Mode:   "system",
				Radius: "0",
			},
			CSRFToken: utils.GenerateCSRFToken(),
		},
		Conversations: []models.Conversation{
			{
				ID:          "1",
				Name:        "circles.diy Design",
				Avatar:      "https://images.unsplash.com/photo-1581291518857-4e27b48ff24e?w=48&h=48&fit=crop",
				LastMessage: "Maria: The color palette looks great! 🎨",
				LastTime:    "2m ago",
				UnreadCount: 3,
				IsOnline:    true,
				IsGroup:     true,
				Participants: []models.User{

					{ID: "maria", Handle: "@maria", Name: "Maria Chen", Avatar: "https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?w=32&h=32&fit=crop&crop=face"},
					{ID: "alex", Handle: "@alex", Name: "Alex Ramirez", Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face"},
					{ID: "jordan", Handle: "@jordan", Name: "Jordan Kim", Avatar: "https://images.unsplash.com/photo-1494790108755-2616b2e6ead5?w=32&h=32&fit=crop&crop=face"},
				},
			},
			{
				ID:          "2",
				Name:        "Emma Wilson",
				Avatar:      "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=48&h=48&fit=crop&crop=face",
				LastMessage: "Thanks for the feedback on the prototype!",
				LastTime:    "15m ago",
				UnreadCount: 0,
				IsOnline:    true,
				IsGroup:     false,
			},
			{
				ID:          "3",
				Name:        "Dev Team",
				Avatar:      "https://images.unsplash.com/photo-1522202176988-66273c2fd55f?w=48&h=48&fit=crop&crop=face",
				LastMessage: "Sam: Ready for the demo tomorrow? 💻",
				LastTime:    "1h ago",
				UnreadCount: 1,
				IsOnline:    false,
				IsGroup:     true,
				Participants: []models.User{

					{ID: "sam", Handle: "@sam", Name: "Sam Rodriguez", Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face"},
					{ID: "riley", Handle: "@riley", Name: "Riley Park", Avatar: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=32&h=32&fit=crop&crop=face"},
				},
			},
			{
				ID:          "4",
				Name:        "Marcus Thompson",
				Avatar:      "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=48&h=48&fit=crop&crop=face",
				LastMessage: "Great job on the presentation!",
				LastTime:    "2h ago",
				UnreadCount: 0,
				IsOnline:    false,
				IsGroup:     false,
			},
			{
				ID:          "5",
				Name:        "Book Club Circle",
				Avatar:      "https://images.unsplash.com/photo-1481627834876-b7833e8f5570?w=48&h=48&fit=crop&crop=face",
				LastMessage: "Lisa: Anyone else loving this chapter? 📖",
				LastTime:    "3h ago",
				UnreadCount: 0,
				IsOnline:    true,
				IsGroup:     true,
			},
		},
		ActiveChat: &models.Conversation{
			ID:       "1",
			Name:     "circles.diy Design",
			Avatar:   "https://images.unsplash.com/photo-1581291518857-4e27b48ff24e?w=48&h=48&fit=crop&crop=face",
			IsOnline: true,
			IsGroup:  true,
			Participants: []models.User{

				{ID: "maya", Handle: "@maria", Name: "Maria Chen", Avatar: "https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?w=48&h=48&fit=crop&crop=face"},
				{ID: "alex", Handle: "@alex", Name: "Alex Ramirez", Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face"},
				{ID: "jordan", Handle: "@jordan", Name: "Jordan Kim", Avatar: "https://images.unsplash.com/photo-1494790108755-2616b2e6ead5?w=32&h=32&fit=crop&crop=face"},
			},
		},
		Messages: []models.Message{
			{
				ID:        "1",
				Content:   "Hey everyone! I've been working on the new design system. What do you think about these color combinations?",
				Timestamp: "10:30 AM",
				Sender: models.User{

					ID:     "maya",
					Handle: "@maria",
					Name:   "Maria Chen",
					Avatar: "https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?w=48&h=48&fit=crop&crop=face",
				},
				IsOwn:  false,
				IsRead: true,
				Type:   "text",
			},
			{
				ID:        "2",
				Content:   "Those shades are perfect! Really captures the minimal vibe.",
				Timestamp: "10:32 AM",
				Sender: models.User{

					ID:     "alex",
					Handle: "@alex",
					Name:   "Alex Ramirez",
					Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face",
				},
				IsOwn:  false,
				IsRead: true,
				Type:   "text",
			},
			{
				ID:        "3",
				Content:   "I love the direction this is taking! The contrast ratios look accessible af 🔥",
				Timestamp: "10:35 AM",
				Sender: models.User{

					ID:     "current_user",
					Handle: "@you",
					Name:   "You",
					Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face",
				},
				IsOwn:  true,
				IsRead: true,
				Type:   "text",
			},
			{
				ID:        "5",
				Content:   "Perfect! This is exactly what I had in mind. Should we schedule a call to discuss implementation?",
				Timestamp: "10:45 AM",
				Sender: models.User{

					ID:     "jordan",
					Handle: "@jordan",
					Name:   "Jordan Kim",
					Avatar: "https://images.unsplash.com/photo-1543610892-0b1f7e6d8ac1?w=32&h=32&fit=crop&crop=face",
				},
				IsOwn:  false,
				IsRead: true,
				Type:   "text",
			},
			{
				ID:        "6",
				Content:   "Great idea! I'm free this afternoon. How about we gather @ 2 PM?",
				Timestamp: "10:46 AM",
				Sender: models.User{

					ID:     "current_user",
					Handle: "@you",
					Name:   "You",
					Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face",
				},
				IsOwn:  true,
				IsRead: true,
				Type:   "text",
			},
			{
				ID:        "7",
				Content:   "Thanks guys 🫶 Sounds good, talk soon!",
				Timestamp: "10:48 AM",
				Sender: models.User{

					ID:     "maya",
					Handle: "@maria",
					Name:   "Maria Chen",
					Avatar: "https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?w=48&h=48&fit=crop&crop=face",
				},
				IsOwn:  false,
				IsRead: false,
				Type:   "text",
			},
		},
		Contacts: []models.Contact{
			{
				User: models.User{

					ID:     "maya",
					Handle: "@maria",
					Name:   "Maria Chen",
					Avatar: "https://images.unsplash.com/photo-1502823403499-6ccfcf4fb453?w=48&h=48&fit=crop&crop=face",
				},
				IsOnline:     true,
				Relationship: "circle_member",
			},
			{
				User: models.User{

					ID:     "alex",
					Handle: "@alex",
					Name:   "Alex Ramirez",
					Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face",
				},
				IsOnline:     true,
				Relationship: "friend",
			},
			{
				User: models.User{

					ID:     "emma",
					Handle: "@emma",
					Name:   "Emma Wilson",
					Avatar: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=32&h=32&fit=crop&crop=face",
				},
				IsOnline:     true,
				Relationship: "friend",
			},
			{
				User: models.User{

					ID:     "marcus",
					Handle: "@marcus",
					Name:   "Marcus Thompson",
					Avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=32&h=32&fit=crop&crop=face",
				},
				IsOnline:     false,
				LastSeen:     "2h ago",
				Relationship: "circle_member",
			},
		},
	}
}

func GetMockMarketplaceData() models.MarketplacePageData {
	return models.MarketplacePageData{
		BaseData: models.BaseData{
			Title:     "Marketplace",
			ActiveNav: "marketplace",
			Theme: models.ThemeSettings{
				Mode:   "system",
				Radius: "0",
			},
			CSRFToken: utils.GenerateCSRFToken(),
		},
		FeaturedItems: []models.MarketplaceItem{
			{
				ID:          "featured-1",
				Title:       "Handcrafted Oak Coffee Table",
				Description: "Beautiful oak coffee table crafted using traditional joinery techniques. Features mortise and tenon construction with natural oil finish. Perfect centerpiece for any living room.",
				Price:       "$850",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1707749522150-e3b1b5f3e079?w=600&h=400&fit=crop&crop=center",
					Alt: "Oak coffee table",
				},
				Location: "Alexandria, NSW",
				Distance: "2.1km",
				TimeAgo:  "3h ago",
				Seller: models.User{

					ID:     "maia",
					Handle: "@maia",
					Name:   "Maia Makes",
					Avatar: "https://images.unsplash.com/photo-1653508242641-09fdb7339942?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Woodworking",
				Category:    "furniture",
				Tags:        []string{"handmade", "oak", "furniture", "traditional"},
				Condition:   "new",
				IsAvailable: true,
				ViewCount:   47,
				IsFeatured:  true,
			},
			{
				ID:          "featured-2",
				Title:       "Professional DJ Decks & Mixer",
				Description: "Two Technics SL1210 Vinyl Turntables and a standalone mixer. Perfect for aspiring DJs or professionals. Includes original box and cables.",
				Price:       "$1200",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1619723525755-8eb7fd5b96f6?w=600&h=400&fit=crop",
					Alt: "DJ mixing desk",
				},
				Location: "Marrickville, NSW",
				Distance: "3.2km",
				TimeAgo:  "5h ago",
				Seller: models.User{

					ID:     "dj_nova",
					Handle: "@dj_nova",
					Name:   "Nova",
					Avatar: "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Music Production",
				Category:    "electronics",
				Tags:        []string{"music", "dj", "vinyl", "professional", "technics"},
				Condition:   "like-new",
				IsAvailable: true,
				ViewCount:   89,
				IsFeatured:  true,
			},
		},
		Items: []models.MarketplaceItem{
			{
				ID:          "1",
				Title:       "Arduino Starter Kit Bundle",
				Description: "Complete Arduino Uno starter kit with breadboard, jumper wires, LEDs, resistors, sensors and project guide. Perfect for beginners learning electronics.",
				Price:       "$65",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?w=400&h=300&fit=crop",
					Alt: "Arduino starter kit",
				},
				Location: "Chippendale, NSW",
				Distance: "1.8km",
				TimeAgo:  "2d ago",
				Seller: models.User{

					ID:     "sara_pcb",
					Handle: "@sara_pcb",
					Name:   "Sara Electronics",
					Avatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "DIY Electronics",
				Category:    "electronics",
				Tags:        []string{"arduino", "beginner", "electronics", "kit"},
				Condition:   "new",
				IsAvailable: true,
				ViewCount:   23,
				IsFeatured:  false,
			},
			{
				ID:          "2",
				Title:       "Vintage Leather Jacket",
				Description: "Authentic 1980s leather jacket in excellent condition. Size medium. Classic biker style with original zippers and lining intact.",
				Price:       "$180",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1551028719-00167b16eac5?w=400&h=300&fit=crop",
					Alt: "Vintage leather jacket",
				},
				Location: "Newtown, NSW",
				Distance: "2.7km",
				TimeAgo:  "1d ago",
				Seller: models.User{

					ID:     "vintage_hunter",
					Handle: "@vintage_hunter",
					Name:   "Riley Vintage",
					Avatar: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Sydney Artists",
				Category:    "clothing",
				Tags:        []string{"vintage", "leather", "1980s", "fashion"},
				Condition:   "good",
				IsAvailable: true,
				ViewCount:   56,
				IsFeatured:  false,
			},
			{
				ID:          "3",
				Title:       "Organic Seedling Bundle",
				Description: "Mix of organic vegetable seedlings ready for transplanting. Includes tomatoes, lettuce, basil, and herbs. Perfect for starting your home garden.",
				Price:       "Trade",
				PriceType:   "trade",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=400&h=300&fit=crop",
					Alt: "Garden seedlings",
				},
				Location: "Marrickville, NSW",
				Distance: "3.1km",
				TimeAgo:  "6h ago",
				Seller: models.User{

					ID:     "green_thumb",
					Handle: "@green_thumb",
					Name:   "Mary Gardens",
					Avatar: "https://images.unsplash.com/photo-1565980100090-3c8b3ec27c43?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Sustainable Living",
				Category:    "garden",
				Tags:        []string{"organic", "seedlings", "vegetables", "sustainable"},
				Condition:   "new",
				IsAvailable: true,
				ViewCount:   34,
				IsFeatured:  false,
			},
			{
				ID:          "4",
				Title:       "Photography Studio Lights",
				Description: "Professional 3-light kit with softboxes and stands. Great for portrait photography or product shots. Includes carrying case.",
				Price:       "$320",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1525199896530-b1d87c75c887?w=400&h=300&fit=crop",
					Alt: "Photography lighting equipment",
				},
				Location: "Surry Hills, NSW",
				Distance: "4.2km",
				TimeAgo:  "4d ago",
				Seller: models.User{

					ID:     "shutterbug",
					Handle: "@shutterbug",
					Name:   "Sarah Kim",
					Avatar: "https://images.unsplash.com/photo-1495745966610-2a67f2297e5e?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Photography",
				Category:    "electronics",
				Tags:        []string{"photography", "lighting", "professional", "studio"},
				Condition:   "good",
				IsAvailable: true,
				ViewCount:   67,
				IsFeatured:  false,
			},
			{
				ID:          "5",
				Title:       "Handmade Ceramic Dinnerware Set",
				Description: "Beautiful 6-piece ceramic dinnerware set. Each piece is hand-thrown and glazed with a unique earth-tone finish. Dishwasher safe.",
				Price:       "$280",
				PriceType:   "sale",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1740811619883-6a3615f3acbb?w=400&h=300&fit=crop",
					Alt: "Ceramic dinnerware",
				},
				Location: "Stanwell Park, NSW",
				Distance: "12.8km",
				TimeAgo:  "1w ago",
				Seller: models.User{

					ID:     "clay_artist",
					Handle: "@clay_artist",
					Name:   "Emma Potter",
					Avatar: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "Sydney Artists",
				Category:    "art",
				Tags:        []string{"ceramic", "handmade", "dinnerware", "art"},
				Condition:   "new",
				IsAvailable: true,
				ViewCount:   78,
				IsFeatured:  false,
			},
			{
				ID:          "6",
				Title:       "Electric Bike - Needs Battery",
				Description: "Solid electric bike frame and motor, but needs a new battery pack. Great project for someone handy with electronics. Includes charger.",
				Price:       "$150",
				PriceType:   "negotiable",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1571068316344-75bc76f77890?w=400&h=300&fit=crop",
					Alt: "Electric bicycle",
				},
				Location: "Redfern, NSW",
				Distance: "3.8km",
				TimeAgo:  "3d ago",
				Seller: models.User{

					ID:     "fix_it_felix",
					Handle: "@fix_it_felix",
					Name:   "Felix Rodriguez",
					Avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=48&h=48&fit=crop&crop=face",
				},
				Circle:      "DIY Electronics",
				Category:    "transport",
				Tags:        []string{"electric", "bike", "repair", "project"},
				Condition:   "fair",
				IsAvailable: true,
				ViewCount:   45,
				IsFeatured:  false,
			},
		},
		Categories: []models.MarketplaceCategory{
			{ID: "furniture", Name: "Furniture", Icon: "🪑", Count: 8},
			{ID: "electronics", Name: "Electronics", Icon: "⚡", Count: 12},
			{ID: "art", Name: "Art & Crafts", Icon: "🎨", Count: 15},
			{ID: "clothing", Name: "Clothing", Icon: "👕", Count: 6},
			{ID: "garden", Name: "Garden & Plants", Icon: "🌱", Count: 9},
			{ID: "tools", Name: "Tools", Icon: "🔧", Count: 7},
			{ID: "books", Name: "Books", Icon: "📚", Count: 4},
			{ID: "transport", Name: "Transport", Icon: "🚲", Count: 3},
		},
		PopularLocations: []models.Location{
			{Name: "Alexandria, NSW", Count: 8},
			{Name: "Marrickville, NSW", Count: 12},
			{Name: "Newtown, NSW", Count: 15},
			{Name: "Chippendale, NSW", Count: 6},
			{Name: "Surry Hills, NSW", Count: 9},
			{Name: "Redfern, NSW", Count: 7},
		},
		TotalItems:   64,
		ItemsPerPage: 12,
		CurrentPage:  1,
		HasMore:      true,
		Filters: models.MarketplaceFilter{
			PriceTypes:  []string{"sale", "trade", "free", "negotiable"},
			Categories:  []string{"furniture", "electronics", "art", "clothing", "garden", "tools", "books", "transport"},
			Conditions:  []string{"new", "like-new", "good", "fair", "poor"},
			MaxDistance: 25,
		},
		ActiveFilters: make(map[string]interface{}),
	}
}
