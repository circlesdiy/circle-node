package domain

import "context"

// CircleRepository defines the interface for circle-related data operations
type CircleRepository interface {
	// Circle CRUD operations
	CreateCircle(ctx context.Context, circle *Circle) error
	GetCircleByID(ctx context.Context, id string) (*Circle, error)
	GetCirclesByOwnerID(ctx context.Context, ownerProfileID string) ([]Circle, error)
	GetCirclesByMemberID(ctx context.Context, memberProfileID string) ([]Circle, error)
	GetPublicCircles(ctx context.Context, limit, offset int) ([]Circle, error)
	UpdateCircle(ctx context.Context, circle *Circle) error
	DeleteCircle(ctx context.Context, id string) error

	// Circle membership operations
	CreateMembership(ctx context.Context, membership *CircleMembership) error
	GetMembershipByID(ctx context.Context, id string) (*CircleMembership, error)
	GetMembershipByCircleAndProfile(ctx context.Context, circleID, profileID string) (*CircleMembership, error)
	GetMembershipsByCircleID(ctx context.Context, circleID string) ([]CircleMembership, error)
	GetMembershipsByProfileID(ctx context.Context, profileID string) ([]CircleMembership, error)
	GetActiveMembershipsByCircleID(ctx context.Context, circleID string) ([]CircleMembership, error)
	GetPendingInvitationsByProfileID(ctx context.Context, profileID string) ([]CircleMembership, error)
	GetCircleMembersByCircleID(ctx context.Context, circleID string, limit, offset int) ([]CircleMembership, error)
	CountMembersByCircleID(ctx context.Context, circleID string) (int, error)
	UpdateMembership(ctx context.Context, membership *CircleMembership) error
	DeleteMembership(ctx context.Context, id string) error

	// Atomic operations
	CreateCircleWithOwnership(ctx context.Context, circle *Circle, membership *CircleMembership) error

	// Event operations
	CreateEvent(ctx context.Context, event *Event) error
	GetEventByID(ctx context.Context, id string) (*Event, error)
	GetEventsByCircleID(ctx context.Context, circleID string) ([]Event, error)
	GetUpcomingEventsByCircleID(ctx context.Context, circleID string) ([]Event, error)
	UpdateEvent(ctx context.Context, event *Event) error
	DeleteEvent(ctx context.Context, id string) error

	// Event RSVP operations
	CreateEventRSVP(ctx context.Context, rsvp *EventRSVP) error
	GetEventRSVPByID(ctx context.Context, id string) (*EventRSVP, error)
	GetEventRSVPByEventAndProfile(ctx context.Context, eventID, profileID string) (*EventRSVP, error)
	GetEventRSVPsByEventID(ctx context.Context, eventID string) ([]EventRSVP, error)
	UpdateEventRSVP(ctx context.Context, rsvp *EventRSVP) error
	DeleteEventRSVP(ctx context.Context, id string) error
	CountAttendeesByEventID(ctx context.Context, eventID string) (int, error)

	// Activity operations
	CreateActivity(ctx context.Context, activity *Activity) error
	GetActivitiesByCircleID(ctx context.Context, circleID string, limit, offset int) ([]Activity, error)
	GetActivitiesByProfileID(ctx context.Context, profileID string, limit, offset int) ([]Activity, error)
}

// PermissionRepository defines the interface for role and permission operations
type PermissionRepository interface {
	// Permission operations
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermissionByID(ctx context.Context, id string) (*Permission, error)
	GetPermissionByKey(ctx context.Context, key string) (*Permission, error)
	GetAllPermissions(ctx context.Context) ([]Permission, error)

	// Role operations
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByID(ctx context.Context, id string) (*Role, error)
	GetRolesByCircleID(ctx context.Context, circleID string) ([]Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id string) error

	// RolePermission operations
	GrantPermissionToRole(ctx context.Context, rolePermission *RolePermission) error
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]Permission, error)

	// ProfileRole operations
	AssignRoleToProfile(ctx context.Context, profileRole *ProfileRole) error
	RevokeRoleFromProfile(ctx context.Context, circleID, profileID, roleID string) error
	GetRolesByProfileInCircle(ctx context.Context, circleID, profileID string) ([]Role, error)
	GetPermissionsByProfileInCircle(ctx context.Context, circleID, profileID string) ([]Permission, error)
	ProfileHasPermission(ctx context.Context, circleID, profileID, permissionKey string) (bool, error)
}

// CommerceRepository defines the interface for commerce operations
type CommerceRepository interface {
	// Order operations
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id string) (*Order, error)
	GetOrdersByProfileID(ctx context.Context, profileID string) ([]Order, error)
	UpdateOrder(ctx context.Context, order *Order) error

	// Ticket operations
	CreateTicket(ctx context.Context, ticket *Ticket) error
	GetTicketByID(ctx context.Context, id string) (*Ticket, error)
	GetTicketByQRCode(ctx context.Context, qrCode string) (*Ticket, error)
	GetTicketsByEventID(ctx context.Context, eventID string) ([]Ticket, error)
	GetTicketsByProfileID(ctx context.Context, profileID string) ([]Ticket, error)
	GetTicketsByOrderID(ctx context.Context, orderID string) ([]Ticket, error)
	UpdateTicket(ctx context.Context, ticket *Ticket) error
	MarkTicketAsUsed(ctx context.Context, ticketID string) error
	CancelTicket(ctx context.Context, ticketID string) error
}
