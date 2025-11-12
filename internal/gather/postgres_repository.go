package gather

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetUserCircleEvents(ctx context.Context, profileID string) ([]CircleWithEvents, error) {
	// First, get all circles for the user
	circlesQuery := `
		SELECT c.id, c.name, c.icon, c.icon_bg_color
		FROM circles c
		JOIN circle_memberships cm ON cm.circle_id = c.id
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND c.deleted_at IS NULL
		ORDER BY c.name`

	circleRows, err := r.pool.Query(ctx, circlesQuery, profileID)
	if err != nil {
		return nil, err
	}
	defer circleRows.Close()

	var circles []CircleWithEvents
	for circleRows.Next() {
		var circle CircleWithEvents
		var icon, iconBgColor *string
		if err := circleRows.Scan(&circle.CircleID, &circle.CircleName, &icon, &iconBgColor); err != nil {
			return nil, err
		}
		if icon != nil {
			circle.Icon = *icon
		}
		if iconBgColor != nil {
			circle.IconBgColor = *iconBgColor
		}
		circles = append(circles, circle)
	}

	if err := circleRows.Err(); err != nil {
		return nil, err
	}

	// For each circle, get its events
	for i := range circles {
		events, hasUnresolved, err := r.getCircleEvents(ctx, circles[i].CircleID, profileID)
		if err != nil {
			return nil, err
		}
		circles[i].Events = events
		circles[i].HasUnresolved = hasUnresolved
	}

	return circles, nil
}

func (r *PostgresRepository) getCircleEvents(ctx context.Context, circleID, profileID string) ([]EventItem, bool, error) {
	query := `
		SELECT
			e.id,
			e.title,
			e.start_time,
			e.end_time,
			e.location,
			COALESCE(er.status, '') as rsvp_status,
			e.organizer_profile_id,
			p.handle as organizer_name,
			COALESCE(ecn.type, '') as coordination_type,
			COALESCE(ecn.message, '') as coordination_message,
			CASE WHEN ecn.resolved_at IS NULL AND ecn.id IS NOT NULL THEN false ELSE true END as coordination_resolved
		FROM events e
		JOIN profiles p ON p.id = e.organizer_profile_id
		LEFT JOIN event_rsvps er ON er.event_id = e.id AND er.profile_id = $2
		LEFT JOIN LATERAL (
			SELECT id, type, message, resolved_at
			FROM event_coordination_needs
			WHERE event_id = e.id
			ORDER BY created_at DESC
			LIMIT 1
		) ecn ON true
		WHERE e.circle_id = $1
		AND e.deleted_at IS NULL
		AND e.start_time > NOW()
		ORDER BY e.start_time ASC`

	rows, err := r.pool.Query(ctx, query, circleID, profileID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var events []EventItem
	hasUnresolved := false

	for rows.Next() {
		var event EventItem
		var location *string
		var coordType, coordMessage string
		var coordResolved bool

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartTime,
			&event.EndTime,
			&location,
			&event.UserRSVPStatus,
			&event.OrganizerProfileID,
			&event.OrganizerName,
			&coordType,
			&coordMessage,
			&coordResolved,
		); err != nil {
			return nil, false, err
		}

		if location != nil {
			event.Location = *location
		}

		event.CoordinationType = coordType
		event.CoordinationMessage = coordMessage
		event.CoordinationResolved = coordResolved

		if coordType != "" && !coordResolved {
			hasUnresolved = true
		}

		// Get attendee count
		count, err := r.GetEventAttendeeCount(ctx, event.ID)
		if err == nil {
			event.AttendeeCount = count
		}

		events = append(events, event)
	}

	return events, hasUnresolved, rows.Err()
}

func (r *PostgresRepository) GetDomainEvents(ctx context.Context, profileID string) ([]DomainEvents, error) {
	// This is a placeholder for domain-based event discovery
	// In a full implementation, this would query based on user's domain interests
	// For now, return empty to avoid errors
	return []DomainEvents{}, nil
}

func (r *PostgresRepository) GetUserCoordinationNeeds(ctx context.Context, profileID string) ([]CoordinationNeedItem, error) {
	query := `
		SELECT
			ecn.event_id,
			e.title,
			ecn.type,
			ecn.message
		FROM event_coordination_needs ecn
		JOIN events e ON e.id = ecn.event_id
		JOIN circles c ON c.id = e.circle_id
		JOIN circle_memberships cm ON cm.circle_id = c.id
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND ecn.resolved_at IS NULL
		AND e.start_time > NOW()
		AND e.deleted_at IS NULL
		ORDER BY ecn.created_at DESC
		LIMIT 5`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var needs []CoordinationNeedItem
	for rows.Next() {
		var need CoordinationNeedItem
		if err := rows.Scan(&need.EventID, &need.EventTitle, &need.Type, &need.Value); err != nil {
			return nil, err
		}

		// Create human-readable labels
		switch need.Type {
		case "transport":
			need.Label = "Transportation"
		case "tickets":
			need.Label = "Tickets"
		case "supplies":
			need.Label = "Supplies"
		default:
			need.Label = need.Type
		}

		needs = append(needs, need)
	}

	return needs, rows.Err()
}

func (r *PostgresRepository) GetUserHostedEvents(ctx context.Context, profileID string) ([]HostedEvent, error) {
	query := `
		SELECT id, title, start_time
		FROM events
		WHERE organizer_profile_id = $1
		AND deleted_at IS NULL
		AND start_time > NOW()
		ORDER BY start_time ASC
		LIMIT 5`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []HostedEvent
	for rows.Next() {
		var event HostedEvent
		if err := rows.Scan(&event.ID, &event.Title, &event.StartTime); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *PostgresRepository) GetUserCircles(ctx context.Context, profileID string) ([]Circle, error) {
	query := `
		SELECT c.id, c.name, COALESCE(c.icon, '')
		FROM circles c
		JOIN circle_memberships cm ON cm.circle_id = c.id
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND c.deleted_at IS NULL
		ORDER BY c.name`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var circles []Circle
	for rows.Next() {
		var circle Circle
		if err := rows.Scan(&circle.ID, &circle.Name, &circle.Icon); err != nil {
			return nil, err
		}
		circles = append(circles, circle)
	}

	return circles, rows.Err()
}

func (r *PostgresRepository) GetEventDetails(ctx context.Context, eventID, profileID string) (*EventDetails, error) {
	// Get event basic info
	query := `
		SELECT
			e.id,
			e.title,
			e.start_time,
			e.end_time,
			COALESCE(e.location, ''),
			e.organizer_profile_id,
			p.handle,
			c.name,
			COALESCE(er.status, '')
		FROM events e
		JOIN profiles p ON p.id = e.organizer_profile_id
		JOIN circles c ON c.id = e.circle_id
		LEFT JOIN event_rsvps er ON er.event_id = e.id AND er.profile_id = $2
		WHERE e.id = $1
		AND e.deleted_at IS NULL`

	var details EventDetails
	var event EventItem
	var circleName string

	err := r.pool.QueryRow(ctx, query, eventID, profileID).Scan(
		&event.ID,
		&event.Title,
		&event.StartTime,
		&event.EndTime,
		&event.Location,
		&event.OrganizerProfileID,
		&event.OrganizerName,
		&circleName,
		&event.UserRSVPStatus,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get attendee count
	count, err := r.GetEventAttendeeCount(ctx, eventID)
	if err == nil {
		event.AttendeeCount = count
	}

	details.Event = event
	details.CircleName = circleName
	details.OrganizerName = event.OrganizerName

	return &details, nil
}

func (r *PostgresRepository) UpdateRSVP(ctx context.Context, eventID, profileID, status string) error {
	query := `
		INSERT INTO event_rsvps (event_id, profile_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (event_id, profile_id)
		DO UPDATE SET status = $3, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, eventID, profileID, status)
	return err
}

func (r *PostgresRepository) GetEventAttendeeCount(ctx context.Context, eventID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM event_rsvps
		WHERE event_id = $1
		AND status IN ('attending', 'maybe')`

	var count int
	err := r.pool.QueryRow(ctx, query, eventID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
