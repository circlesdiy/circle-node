package templates

import (
	"circles.diy/internal/models"
	"circles.diy/internal/utils"
)

func GetMockDashboardData() models.DashboardData {
	return models.DashboardData{
		BaseData: models.BaseData{
			Title:     "Dashboard",
			ActiveNav: "dashboard",
			Theme: models.ThemeSettings{
				Mode:   "system",
				Radius: "0",
			},
			CSRFToken: utils.GenerateCSRFToken(),
		},
		Stats: models.DashboardStats{
			ActiveCircles:       5,
			ActiveCirclesChange: "2 need your attention",
			UpcomingGatherings:  4,
			GatheringsNext:      "Next: D&B Workshop in 2 days",
			UnreadMessages:      3,
			MessagesFrom:        "From 2 circles",
		},
		QuickCircles: []models.QuickCircle{
			{
				ID:          "1",
				Name:        "Music crew",
				Icon:        "🎵",
				IconBgColor: "var(--warm-accent)",
				MemberCount: 8,
				Status:      "Active",
			},
			{
				ID:          "2",
				Name:        "Housemates",
				Icon:        "🏠",
				IconBgColor: "var(--cool-accent)",
				MemberCount: 5,
				Status:      "Active",
			},
			{
				ID:          "3",
				Name:        "Climbing",
				Icon:        "🧗",
				IconBgColor: "var(--earth-accent)",
				MemberCount: 12,
				Status:      "Active",
			},
			{
				ID:          "4",
				Name:        "Coffee nerds",
				Icon:        "☕",
				IconBgColor: "var(--sage-accent)",
				MemberCount: 4,
				Status:      "Quiet",
			},
		},
		ActivityFeed: []models.ActivityItem{
			{
				ID:       "1",
				Icon:     `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256"><path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Z"></path></svg>`,
				Text:     "<strong>Sarah</strong> asked about transport to D&B Workshop",
				Meta:     "Music crew • 2m ago",
				ActionBy: "Sarah",
			},
			{
				ID:   "2",
				Icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256"><path d="M208,32H184V24a8,8,0,0,0-16,0v8H88V24a8,8,0,0,0-16,0v8H48A16,16,0,0,0,32,48V208"></path></svg>`,
				Text: "Weekend climb session confirmed",
				Meta: "Climbing • 1h ago",
			},
			{
				ID:       "3",
				Icon:     `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256"><path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Z"></path></svg>`,
				Text:     "<strong>Alex</strong> updated the cleaning schedule",
				Meta:     "Housemates • 3h ago",
				ActionBy: "Alex",
			},
		},
		Feed: []models.FeedItem{
			{
				ID: "1",
				User: models.User{

					ID:     "maia",
					Handle: "@maia",
					Avatar: "https://images.unsplash.com/photo-1653508242641-09fdb7339942?w=48&h=48&fit=crop&crop=face",
				},
				Content: "Just finished this oak coffee table! Happy to step out of my comfort-zone and share some joinery! This piece is available 💜💸",
				TimeAgo: "2m ago",
				Circle:  "Woodworking",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1707749522150-e3b1b5f3e079?w=600&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NHx8b2FrJTIwdGFibGV8ZW58MHx8MHx8fDI%3D",
					Alt: "Oak coffee table project",
				},
				CanBuy: true,
				Replies: []models.Reply{
					{
						ID: "f1_r1",
						User: models.User{

							ID:     "wood_enthusiast",
							Handle: "@wood_enthusiast",
							Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=32&h=32&fit=crop&crop=face",
						},
						Content:   "Absolutely gorgeous! The vintage vibe is delightful!",
						TimeAgo:   "1m ago",
						Timestamp: "2025-09-12T13:29:00Z",
					},
				},
			},
			{
				ID: "2",
				User: models.User{

					ID:     "heathtyler",
					Handle: "@heathtyler",
					Avatar: "https://images.unsplash.com/photo-1581391528803-54be77ce23e3?w=48&h=48&fit=crop&crop=face",
				},
				Content: "Warming up the barbeque and just got couple cases of the finest bread-water. Keen to see you all... remember 7PM dont be late!",
				TimeAgo: "30m ago",
				Circle:  "The Crop Circle",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1664463758574-e640a7a998d4?q=80&w=600&auto=format&fit=crop",
					Alt: "BBQ gathering setup",
				},
				CanBuy: false,
				Replies: []models.Reply{
					{
						ID: "f1_r1",
						User: models.User{

							ID:     "bookworm",
							Handle: "@bookworm",
							Avatar: "https://images.unsplash.com/photo-1635830609300-ad974ad25d6a?w=32&h=32&fit=crop&crop=face",
						},
						Content:   "Keen as a number of beans 🫘!!",
						TimeAgo:   "15m ago",
						Timestamp: "2025-09-12T13:15:00Z",
						Replies: []models.Reply{
							{
								ID: "f1_r1_r1",
								User: models.User{

									ID:     "heathtyler",
									Handle: "@heathtyler",
									Avatar: "https://images.unsplash.com/photo-1581391528803-54be77ce23e3?w=48&h=48&fit=crop&crop=face",
								},
								Content:   "The magical fruit ✨",
								TimeAgo:   "15m ago",
								Timestamp: "2025-09-12T13:15:00Z",
								Replies: []models.Reply{
									{
										ID: "f1_r1_r1_r1",
										User: models.User{

											ID:     "bookworm",
											Handle: "@bookworm",
											Avatar: "https://images.unsplash.com/photo-1635830609300-ad974ad25d6a?w=32&h=32&fit=crop&crop=face",
										},
										Content:   "Thats deep!",
										TimeAgo:   "15m ago",
										Timestamp: "2025-09-12T13:15:00Z",
										Replies: []models.Reply{
											{
												ID: "f1_r1_r1_r1_r1",
												User: models.User{

													ID:     "heathtyler",
													Handle: "@heathtyler",
													Avatar: "https://images.unsplash.com/photo-1581391528803-54be77ce23e3?w=48&h=48&fit=crop&crop=face",
												},
												Content: "So deep imma need a shovel!",
												TimeAgo: "15m ago",
												Replies: []models.Reply{
													{
														ID: "f1_r1_r1_r1_r1_r1",
														User: models.User{

															ID:     "bookworm",
															Handle: "@bookworm",
															Avatar: "https://images.unsplash.com/photo-1635830609300-ad974ad25d6a?w=32&h=32&fit=crop&crop=face",
														},
														Content: `▒▒▒▄▄▄▄▄▄▄▄▄
░▄███████▀▀▀▀▀▀███████▄
░▐████▀▒▒▒▒▒▒▒▒▒▀██████▄
░███▀▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▀█████
░▐██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒████▌
░▐█▌▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒████▌
░░█▒▄▀▀▀▀▀▄▒▒▄▀▀▀▀▀▄▒▐███▌
░░░▐░░░▄▄░░▌▐░░░▄▄░░▌▐███▌
░▄▀▌░░░▀▀░░▌▐░░░▀▀░░▌▒▀▒█▌
░▌▒▀▄░░░░▄▀▒▒▀▄░░░▄▀▒▒▄▀▒▌
░▀▄▐▒▀▀▀▀▒▒▒▒▒▒▀▀▀▒▒▒▒▒▒█
░░░▀▌▒▄██▄▄▄▄████▄▒▒▒▒█▀
░░░░▄██████████████▒▒▐▌
░░░▀███▀▀████▀█████▀▒▌
░░░░░▌▒▒▒▄▒▒▒▄▒▒▒▒▒▒▐
░░░░░▌▒▒▒▒▀▀▀▒▒▒▒▒▒▒▐`,
														TimeAgo:   "15m ago",
														Timestamp: "2025-09-12T13:15:00Z",
														Replies: []models.Reply{
															{
																ID: "f1_r1_r1_r1_r1_r1_r1",
																User: models.User{

																	ID:     "heathtyler",
																	Handle: "@heathtyler",
																	Avatar: "https://images.unsplash.com/photo-1581391528803-54be77ce23e3?w=48&h=48&fit=crop&crop=face",
																},
																Content: "LOL",
																TimeAgo: "15m ago",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				ID: "3",
				User: models.User{

					ID:     "sara_pcb",
					Handle: "@sara_pcb",
					Avatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?q=80&w=48&h=48&auto=format&fit=crop",
				},
				Content: "New tutorial series starting: \"Arduino for Beginners\". First session covers basic circuits and programming fundamentals.",
				TimeAgo: "1h ago",
				Circle:  "DIY Electronics",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?w=600&fit=crop&crop=center",
					Alt: "Arduino tutorial setup",
				},
				CanBuy: false,
			},
			{
				ID: "4",
				User: models.User{

					ID:     "zucc",
					Handle: "@zucc",
					Avatar: "https://media.tenor.com/y1mYLo66EuoAAAAM/zucky.gifs",
				},
				Content: "circles.diy has changed the game forever. \n\nFuck, I wish I'd thought of that.",
				TimeAgo: "1h ago",
				Circle:  "Communication Software",
			},
		},
		FeedOffset: 4,
		Circles: []models.Circle{
			{ID: "1", Name: "Woodworking", Thumbnail: "https://images.unsplash.com/photo-1504148455328-c376907d081c?w=48&h=48&fit=crop&crop=center", LastActivity: "2m ago"},
			{ID: "2", Name: "The Crop Circle", Thumbnail: "https://images.unsplash.com/photo-1611402858501-d3de70f3c67e?q=80&w=48&auto=format&fit=crop", LastActivity: "15m ago"},
			{ID: "3", Name: "DIY Electronics", Thumbnail: "https://images.unsplash.com/photo-1518611012118-696072aa579a?w=48&h=48&fit=crop&crop=center", LastActivity: "1h ago"},
		},
		Discussions: []models.Discussion{
			{ID: "1", Title: "Sustainable Materials: Where to Source?", Preview: "Looking for suppliers of ethically sourced hardwoods. What are your go-to sources for...", Circle: "Woodworking", TimeAgo: "23 replies • 45m ago"},
			{ID: "2", Title: "Pricing Creative Work: Community Wisdom", Preview: "How do you approach pricing custom commissions? Struggling to find the balance between...", Circle: "Local Artists", TimeAgo: "8 replies • 2h ago"},
		},
		Events: []models.Event{
			{ID: "1", Title: "Workshop: Intro to Drum & Bass Mixing", Time: "2:00 PM", Day: "31", Month: "Aug"},
			{ID: "2", Title: "Monthly Showcase", Time: "7:00 PM", Day: "02", Month: "Sep"},
		},
		Ripples: []models.Ripple{
			{
				ID: "1",
				User: models.User{

					ID:     "maia",
					Handle: "@maia",
					Avatar: "https://images.unsplash.com/photo-1653508242641-09fdb7339942?w=32&h=32&fit=crop&crop=face",
				},
				Content:     "Quick progress shot of the dovetail joints 🔧",
				ContentType: "image",
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1504148455328-c376907d081c?w=300&h=200&fit=crop&crop=center",
					Alt: "Dovetail joint detail on oak wood",
				},
				ExpiresIn: "2h left",
				Circle:    "Woodworking",
			},
			{
				ID: "2",
				User: models.User{

					ID:     "sara_pcb",
					Handle: "@sara_pcb",
					Avatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=32&h=32&fit=crop&crop=face",
				},
				Content:     "30-second tip: soldering tiny SMD components",
				ContentType: "video",
				Video: &models.MediaItem{
					URL: "https://sample-videos.com/zip/10/mp4/SampleVideo_360x240_1mb.mp4",
					Alt: "SMD soldering technique demonstration",
				},
				ExpiresIn: "6h left",
				Circle:    "DIY Electronics",
			},
			{
				ID: "4",
				User: models.User{

					ID:     "green_thumb",
					Handle: "@green_thumb",
					Avatar: "https://images.unsplash.com/photo-1565980100090-3c8b3ec27c43?w=32&h=32&fit=crop&crop=face",
				},
				Content:     "Found this amazing article on companion planting",
				ContentType: "link",
				Link: &models.LinkPreview{
					URL:         "https://example.com/companion-planting-guide",
					Title:       "The Complete Guide to Companion Planting",
					Description: "Learn which vegetables grow best together and create a thriving garden ecosystem.",
					Image:       "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=200&h=120&fit=crop",
					Domain:      "gardenguide.com",
				},
				ExpiresIn: "1d left",
				Circle:    "Sustainable Living",
			},
			{
				ID: "5",
				User: models.User{

					ID:     "craft_collective",
					Handle: "@craft_collective",
					Avatar: "https://images.unsplash.com/photo-1511367461989-f85a21fda167?w=32&h=32&fit=crop&crop=face",
				},
				Content:     "Pro tip: pre-drill hardwood to avoid splitting. Learned this the hard way! 🪵",
				ContentType: "text",
				ExpiresIn:   "18h left",
				Circle:      "Woodworking",
			},
		},
		MarketplaceItems: []models.MarketplaceItem{
			{
				ID:       "1",
				Title:    "Handcrafted Oak Coffee Table",
				Price:    "$850",
				Image:    &models.MediaItem{URL: "https://images.unsplash.com/photo-1707749522150-e3b1b5f3e079?w=128&fit=crop&crop=center", Alt: "Cordless drill"},
				Location: "Alexandria, NSW",
				TimeAgo:  "3h ago",
			},
			{
				ID:       "2",
				Title:    "Professional DJ Decks & Mixer",
				Price:    "1200",
				Image:    &models.MediaItem{URL: "https://images.unsplash.com/photo-1619723525755-8eb7fd5b96f6?w=128&fit=crop", Alt: "Ceramic wheel"},
				Location: "Marickville, NSW",
				TimeAgo:  "1d ago",
			},
			{
				ID:       "3",
				Title:    "Oak Lumber Bundle",
				Price:    "$120",
				Image:    &models.MediaItem{URL: "https://images.unsplash.com/photo-1702195789139-4897ff9b0083?w=128&fit=crop", Alt: "Oak lumber"},
				Location: "Granville, NSW",
				TimeAgo:  "2d ago",
			},
		},
		Impact: []models.ImpactItem{
			{Label: "Contributions", Value: "47"},
			{Label: "Discussions", Value: "23"},
			{Label: "Circle Tithe", Value: "$5/month"},
		},
	}
}
func GetMockCirclesPageData() models.CirclesPageData {
	return models.CirclesPageData{
		BaseData: models.BaseData{
			Title:     "My Circles",
			ActiveNav: "circles",
			Theme: models.ThemeSettings{
				Mode:   "system",
				Radius: "0",
			},
			CSRFToken: utils.GenerateCSRFToken(),
		},
		Circles: []models.Circle{
			{
				ID:           "1",
				Name:         "Woodworking",
				Description:  "Traditional craftsmanship meets modern techniques. Share projects, ask questions, and connect with fellow makers.",
				Thumbnail:    "https://images.unsplash.com/photo-1702195789139-4897ff9b0083?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1497219055242-93359eeed651?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "47",
				OnlineCount:  "8",
				UserRole:     "admin",
				JoinedDate:   "6 months ago",
				LastActivity: "2m ago",
				Active:       true,
			},
			{
				ID:           "2",
				Name:         "The Crop Circle",
				Description:  "The official circle for Alienated Collective, a sydney based group of local ravers, doofers, party and music enthusiasts.",
				Thumbnail:    "https://images.unsplash.com/photo-1611402858501-d3de70f3c67e?q=80&w=80&h=80&auto=format&fit=crop",
				Banner:       "https://images.unsplash.com/photo-1594623930572-300a3011d9ae?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "12",
				OnlineCount:  "3",
				UserRole:     "member",
				JoinedDate:   "3 months ago",
				LastActivity: "15m ago",
				Active:       true,
			},
			{
				ID:           "3",
				Name:         "DIY Electronics",
				Description:  "Arduino, Raspberry Pi, circuit design, and everything in between. From beginner tutorials to advanced projects.",
				Thumbnail:    "https://images.unsplash.com/photo-1603732551658-5fabbafa84eb?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "156",
				OnlineCount:  "24",
				UserRole:     "member",
				JoinedDate:   "8 months ago",
				LastActivity: "1h ago",
				Active:       true,
			},
			{
				ID:           "4",
				Name:         "Sydney Artists",
				Description:  "Supporting Sydney's creative community through collaboration, critique, and celebration of local talent.",
				Thumbnail:    "https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1663505819040-00bbd0814fab?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "89",
				OnlineCount:  "0",
				UserRole:     "member",
				JoinedDate:   "4 months ago",
				LastActivity: "2h ago",
				Active:       false,
			},
			{
				ID:           "5",
				Name:         "Communication Software",
				Description:  "Discussing the future of digital communication, platform design, and better community building tools.",
				Thumbnail:    "https://images.unsplash.com/photo-1526045612212-70caf35c14df?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "34",
				OnlineCount:  "3",
				UserRole:     "owner",
				JoinedDate:   "1 year ago",
				LastActivity: "1h ago",
				Active:       true,
			},
		},
		RecentActivity: []models.CircleActivity{
			{
				ID:       "1",
				CircleID: "1",
				Type:     "post",
				Title:    "New oak coffee table project completed",
				Content:  "Just finished this oak coffee table! Happy to step out of my comfort-zone...",
				User:     "@maia",
				TimeAgo:  "2m ago",
			},
			{
				ID:       "2",
				CircleID: "2",
				Type:     "event",
				Title:    "BBQ gathering tonight",
				Content:  "Warming up the barbeque and just got couple cases of the finest bread-water...",
				User:     "@heathtyler",
				TimeAgo:  "15m ago",
			},
			{
				ID:       "3",
				CircleID: "3",
				Type:     "announcement",
				Title:    "New Arduino tutorial series starting",
				Content:  "First session covers basic circuits and programming fundamentals...",
				User:     "@sara_pcb",
				TimeAgo:  "1h ago",
			},
			{
				ID:       "4",
				CircleID: "4",
				Type:     "member_joined",
				Title:    "New member joined Sydney Artists",
				Content:  "@alex_painter has joined the circle",
				User:     "System",
				TimeAgo:  "3h ago",
			},
			{
				ID:       "5",
				CircleID: "1",
				Type:     "post",
				Title:    "Workshop layout considerations",
				Content:  "Spending today selecting timber for the next commission...",
				User:     "@maia",
				TimeAgo:  "1d ago",
			},
		},
		Stats: models.CircleStats{
			TotalPosts:     127,
			ActiveMembers:  52,
			RecentActivity: "Very High",
			WeeklyGrowth:   "+12%",
			EngagementRate: "85%",
		},
		FeaturedCircles: []models.Circle{
			{
				ID:           "6",
				Name:         "Sustainable Living",
				Description:  "Practical tips for reducing environmental impact through daily choices and community action.",
				Thumbnail:    "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "234",
				OnlineCount:  "18",
				UserRole:     "",
				JoinedDate:   "",
				LastActivity: "",
				Active:       true,
			},
			{
				ID:           "7",
				Name:         "Urban Beekeeping",
				Description:  "City-based beekeeping community sharing techniques, equipment, and harvest stories.",
				Thumbnail:    "https://images.unsplash.com/photo-1558642452-9d2a7deb7f62?w=80&h=80&fit=crop&crop=center",
				Banner:       "https://images.unsplash.com/photo-1558642452-9d2a7deb7f62?w=400&h=200&fit=crop&crop=center",
				MemberCount:  "78",
				OnlineCount:  "6",
				UserRole:     "",
				JoinedDate:   "",
				LastActivity: "",
				Active:       true,
			},
		},
	}
}

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

func GetMockGatherData() models.GatherPageData {
	return models.GatherPageData{
		BaseData: models.BaseData{
			Title:     "Gather",
			ActiveNav: "gather",
			Theme: models.ThemeSettings{
				Mode:   "system",
				Radius: "0",
			},
			CSRFToken: utils.GenerateCSRFToken(),
		},
		CircleGatherings: []models.CircleGathering{
			{
				CircleID:         "1",
				CircleName:       "Music crew",
				Icon:             "🎵",
				IconBgColor:      "rgba(232, 213, 196, 0.4)",
				GatheringCount:   3,
				NeedsCoordination: true,
				Gatherings: []models.GatheringItem{
					{
						ID:            "1",
						Title:         "D&B Mixing Workshop",
						TimeAgo:       "In 2 days",
						TimeRange:     "Saturday 2pm-5pm",
						RSVPStatus:    "going",
						Description:   "Learn the fundamentals of crate digging and mixing D&B sets with industry pros.",
						Location:      "Studio Complex",
						AttendeeCount: 4,
						CoordinationNeeded: &models.CoordinationNeed{
							Icon:  "🚗",
							Title: "Transport coordination",
							Text:  "Sarah asked: \"Who's driving? I can take 3 people\"",
						},
					},
					{
						ID:            "2",
						Title:         "Pitch Music Festival",
						TimeAgo:       "Next Friday",
						TimeRange:     "6pm-late",
						RSVPStatus:    "going",
						Description:   "Three stages of electronic music across genres.",
						Location:      "Carriageworks",
						AttendeeCount: 7,
					},
					{
						ID:          "3",
						Title:       "Studio Listening Session",
						TimeAgo:     "In 2 weeks",
						TimeRange:   "Sunday 4pm",
						RSVPStatus:  "maybe",
						Description: "Share your recent productions and get feedback.",
						Location:    "Alex's place",
					},
				},
			},
			{
				CircleID:         "2",
				CircleName:       "Climbing",
				Icon:             "🧗",
				IconBgColor:      "rgba(212, 197, 176, 0.4)",
				GatheringCount:   1,
				NeedsCoordination: false,
				Gatherings: []models.GatheringItem{
					{
						ID:            "4",
						Title:         "Weekend Climb @ Hardrock",
						TimeAgo:       "This Saturday",
						TimeRange:     "10am-2pm",
						RSVPStatus:    "going",
						Description:   "Regular Saturday session. All levels welcome.",
						Location:      "Hardrock Climbing, Annandale",
						AttendeeCount: 8,
					},
				},
			},
			{
				CircleID:         "3",
				CircleName:       "Coffee nerds",
				Icon:             "☕",
				IconBgColor:      "rgba(201, 213, 196, 0.4)",
				GatheringCount:   0,
				NeedsCoordination: false,
				Gatherings:       []models.GatheringItem{},
			},
		},
		DomainDiscovery: []models.DomainSection{
			{
				Name:             "Music & Events",
				Icon:             "🎸",
				SerendipityCount: 3,
				Gatherings: []models.DomainGathering{
					{
						ID:                 "5",
						Title:              "Synth Building Workshop",
						TimeAgo:            "Next Tuesday",
						TimeRange:          "6pm-9pm",
						Description:        "Build your own analog synthesizer from scratch. All materials provided.",
						Location:           "Maker Space, Redfern",
						SerendipityCircles: 2,
					},
				},
			},
			{
				Name: "Sustainability",
				Icon: "🌱",
				Gatherings: []models.DomainGathering{
					{
						ID:          "6",
						Title:       "Community Garden Build Day",
						TimeAgo:     "In 4 days",
						TimeRange:   "Saturday 9am-3pm",
						Description: "Help build raised beds and plant vegetables for our community garden.",
						Location:    "Marrickville Community Centre",
					},
				},
			},
		},
		UserCircles: []models.CircleOption{
			{ID: "1", Name: "Music crew"},
			{ID: "2", Name: "Climbing"},
			{ID: "3", Name: "Coffee nerds"},
		},
		ActiveCoordination: []models.CoordinationItem{
			{Label: "Transport to D&B Workshop", Value: "2 needs", GatheringID: "1"},
			{Label: "Festival group ticket", Value: "$45 each", GatheringID: "2"},
		},
		UserHosting: []models.HostingEvent{
			{ID: "7", Title: "Woodworking Show & Tell", TimeAgo: "In 9 days"},
		},
		FeaturedEvents: []models.GatherEvent{
			{
				ID:          "1",
				Title:       "Workshop: Intro to Drum & Bass Mixing",
				Description: "Learn the fundamentals of crate digging, curation and mixing of D&B sets with industry professionals. From atmospheric to bass-face bangers, we'll cover all the essentials.",
				Host: models.User{

					ID:     "dj_nova",
					Handle: "@dj_nova",
					Name:   "Nova",
					Avatar: "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Music Production",
				DateTime: "2025-09-05T14:00:00Z",
				TimeAgo:  "in 2 days",
				Duration: "3 hours",
				Location: models.EventLocation{
					Type:    "venue",
					Name:    "Studio Complex",
					Address: "42 King Street",
					City:    "Sydney, NSW",
				},
				Type:          "in-person",
				Category:      "Workshop",
				IsTicketed:    true,
				Price:         "45",
				Currency:      "AUD",
				Capacity:      20,
				AttendeeCount: 12,
				RSVPStatus:    "going",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1619723525755-8eb7fd5b96f6?w=600&h=300&fit=crop",
					Alt: "Music production workshop setup",
				},
				Tags: []string{"music", "workshop", "beginner-friendly"},
			},
			{
				ID:          "2",
				Title:       "Community Garden Build Day",
				Description: "Join us for a hands-on day of building raised beds, planting and creating a sustainable food garden for our local community.",
				Host: models.User{

					ID:     "green_thumb",
					Handle: "@green_thumb",
					Name:   "Mary Gardens",
					Avatar: "https://images.unsplash.com/photo-1565980100090-3c8b3ec27c43?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Sustainable Living",
				DateTime: "2025-09-07T09:00:00Z",
				TimeAgo:  "in 4 days",
				Duration: "6 hours",
				Location: models.EventLocation{
					Type:    "address",
					Name:    "Marrickville Community Centre",
					Address: "166 Marrickville Road",
					City:    "Marrickville, NSW",
				},
				Type:          "in-person",
				Category:      "Community",
				IsTicketed:    false,
				Capacity:      30,
				AttendeeCount: 18,
				RSVPStatus:    "maybe",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=600&h=300&fit=crop",
					Alt: "Community garden volunteers",
				},
				Tags: []string{"community", "gardening", "sustainability"},
			},
		},
		UpcomingEvents: []models.GatherEvent{
			{
				ID:          "3",
				Title:       "Virtual Reality Art Exhibition",
				Description: "Explore immersive digital art installations by local artists. Experience the future of creative expression.",
				Host: models.User{

					ID:     "pixel_artist",
					Handle: "@pixel_artist",
					Name:   "Alex Chen",
					Avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Digital Arts",
				DateTime: "2025-09-08T18:00:00Z",
				TimeAgo:  "in 5 days",
				Duration: "2 hours",
				Location: models.EventLocation{
					Type:       "online",
					Name:       "VR Gallery Space",
					OnlineLink: "https://vr.circles.diy/gallery",
				},
				Type:          "online",
				Category:      "Art",
				IsTicketed:    true,
				Price:         "25",
				Currency:      "AUD",
				Capacity:      50,
				AttendeeCount: 23,
				RSVPStatus:    "not_responded",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1637664067012-241eb64b7675?w=600&h=300&fit=crop",
					Alt: "VR art installation",
				},
				Tags: []string{"art", "vr", "digital"},
			},
			{
				ID:          "4",
				Title:       "Electronics Repair Café",
				Description: "Bring your broken electronics and learn to repair them with our expert volunteers. Reduce waste and learn valuable skills.",
				Host: models.User{

					ID:     "fix_it_felix",
					Handle: "@fix_it_felix",
					Name:   "Felix Rodriguez",
					Avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "DIY Electronics",
				DateTime: "2025-09-10T10:00:00Z",
				TimeAgo:  "in 1 week",
				Duration: "4 hours",
				Location: models.EventLocation{
					Type:    "venue",
					Name:    "Maker Space",
					Address: "78 Cleveland Street",
					City:    "Chippendale, NSW",
				},
				Type:          "in-person",
				Category:      "Workshop",
				IsTicketed:    false,
				Capacity:      15,
				AttendeeCount: 8,
				RSVPStatus:    "not_responded",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1576613109753-27804de2cba8?w=600&h=300&fit=crop",
					Alt: "Electronics repair workspace",
				},
				Tags: []string{"repair", "electronics", "sustainability"},
			},
			{
				ID:          "5",
				Title:       "Photography Walk: Urban Architecture",
				Description: "Explore the city's architectural gems through your lens. We'll meet at Central Station and walk through various districts capturing both modern and heritage buildings.",
				Host: models.User{

					ID:     "shutterbug",
					Handle: "@shutterbug",
					Name:   "Sarah Kim",
					Avatar: "https://images.unsplash.com/photo-1495745966610-2a67f2297e5e?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Photography",
				DateTime: "2025-09-12T14:00:00Z",
				TimeAgo:  "in 1 week",
				Duration: "3 hours",
				Location: models.EventLocation{
					Type:    "venue",
					Name:    "Central Station",
					Address: "Eddy Avenue",
					City:    "Sydney, NSW",
				},
				Type:          "in-person",
				Category:      "Social",
				IsTicketed:    false,
				Price:         "0",
				Currency:      "AUD",
				Capacity:      20,
				AttendeeCount: 8,
				RSVPStatus:    "attending",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1449824913935-59a10b8d2000?w=600&h=300&fit=crop",
					Alt: "Urban architecture photography",
				},
				Tags: []string{"photography", "architecture", "walk"},
			},
			{
				ID:          "6",
				Title:       "Climate Solutions Workshop",
				Description: "Learn about practical climate action strategies. Join us in-person or online to discuss renewable energy, sustainable living, climate resilience and community initiatives.",
				Host: models.User{

					ID:     "green_future",
					Handle: "@green_future",
					Name:   "Jordan Martinez",
					Avatar: "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Climate Action",
				DateTime: "2025-09-15T15:30:00Z",
				TimeAgo:  "in 11 days",
				Duration: "2.5 hours",
				Location: models.EventLocation{
					Type:       "hybrid",
					Name:       "Community Centre Hall A",
					Address:    "45 Green Square Road",
					City:       "Green Square, NSW",
					OnlineLink: "https://meet.circles.diy/climate-workshop",
				},
				Type:          "hybrid",
				Category:      "Workshop",
				IsTicketed:    false,
				Price:         "0",
				Currency:      "AUD",
				Capacity:      80,
				AttendeeCount: 35,
				RSVPStatus:    "maybe",
				IsHost:        false,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1552799446-159ba9523315?w=600&h=300&fit=crop",
					Alt: "Climate action workshop",
				},
				Tags: []string{"climate", "sustainability", "workshop"},
			},
		},
		MyEvents: []models.GatherEvent{
			{
				ID:          "5",
				Title:       "Woodworking Show & Tell",
				Description: "Monthly gathering to share recent projects, techniques, and connect with fellow woodworkers. BYO project photos!",
				Host: models.User{

					ID:     "maia",
					Handle: "@maia",
					Name:   "Maia Makes",
					Avatar: "https://images.unsplash.com/photo-1653508242641-09fdb7339942?w=48&h=48&fit=crop&crop=face",
				},
				Circle:   "Woodworking",
				DateTime: "2025-09-12T19:00:00Z",
				TimeAgo:  "in 9 days",
				Duration: "2 hours",
				Location: models.EventLocation{
					Type:    "venue",
					Name:    "The Workshop",
					Address: "123 Workshop Lane",
					City:    "Alexandria, NSW",
				},
				Type:          "in-person",
				Category:      "Social",
				IsTicketed:    false,
				Capacity:      25,
				AttendeeCount: 11,
				RSVPStatus:    "going",
				IsHost:        true,
				Image: &models.MediaItem{
					URL: "https://images.unsplash.com/photo-1702195789139-4897ff9b0083?w=600&h=300&fit=crop",
					Alt: "Woodworking tools and projects",
				},
				Tags: []string{"woodworking", "showcase", "networking"},
				Announcements: []models.Announcement{
					{
						ID:      "1",
						Title:   "Don't forget your project photos!",
						Content: "Reminder to bring photos of your recent work to share with the group. We love seeing what everyone's been creating!",
						Author: models.User{

							ID:     "maia",
							Handle: "@maia",
							Name:   "Maia Makes",
							Avatar: "https://images.unsplash.com/photo-1653508242641-09fdb7339942?w=32&h=32&fit=crop&crop=face",
						},
						TimeAgo: "2 days ago",
					},
				},
			},
		},
		EventCategories: []models.EventCategory{
			{ID: "workshop", Name: "Workshops", Icon: "🔨", Count: 3},
			{ID: "social", Name: "Social", Icon: "🍻", Count: 1},
			{ID: "community", Name: "Community", Icon: "🤝", Count: 1},
			{ID: "art", Name: "Art & Culture", Icon: "🎨", Count: 1},
			{ID: "tech", Name: "Technology", Icon: "💻", Count: 0},
			{ID: "outdoor", Name: "Outdoor", Icon: "🌲", Count: 0},
		},
		PopularLocations: []models.EventLocation{
			{Type: "venue", Name: "The Workshop", City: "Alexandria, NSW"},
			{Type: "venue", Name: "Community Hub", City: "Newtown, NSW"},
			{Type: "venue", Name: "Maker Space", City: "Chippendale, NSW"},
			{Type: "online", Name: "Virtual Meetup"},
		},
	}
}
