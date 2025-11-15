package events

import (
	"context"
	"time"

	"circles.diy/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	query := `
		INSERT INTO events (
			id, circle_id, organizer_profile_id, title, description, location, timezone,
			start_time, end_time, requires_ticket, capacity, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.CircleID,
		event.OrganizerProfileID,
		event.Title,
		event.Description,
		event.Location,
		event.Timezone,
		event.StartTime,
		event.EndTime,
		event.RequiresTicket,
		event.Capacity,
		event.CreatedAt,
		event.UpdatedAt,
	)

	return err
}

func (r *PostgresRepository) GetEventByID(ctx context.Context, eventID string) (*domain.Event, error) {
	query := `
		SELECT id, circle_id, organizer_profile_id, title, description, location, timezone,
		       start_time, end_time, requires_ticket, capacity, created_at, updated_at, deleted_at
		FROM events
		WHERE id = $1`

	var event domain.Event
	var description, location, timezone *string
	var endTime *time.Time
	var capacity *int

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.CircleID,
		&event.OrganizerProfileID,
		&event.Title,
		&description,
		&location,
		&timezone,
		&event.StartTime,
		&endTime,
		&event.RequiresTicket,
		&capacity,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if description != nil {
		event.Description = *description
	}
	if location != nil {
		event.Location = *location
	}
	if timezone != nil {
		event.Timezone = *timezone
	}
	if endTime != nil {
		event.EndTime = *endTime
	}
	if capacity != nil {
		event.Capacity = *capacity
	}

	return &event, nil
}

func (r *PostgresRepository) GetEventDetails(ctx context.Context, eventID, viewerProfileID string) (*EventDetails, error) {
	// Get event basic info with circle and organizer details
	query := `
		SELECT
			e.id, e.circle_id, e.organizer_profile_id, e.title, e.description, e.location, e.timezone,
			e.start_time, e.end_time, e.requires_ticket, e.capacity, e.created_at, e.updated_at,
			c.name as circle_name,
			p.handle as organizer_name,
			COALESCE(er.status, '') as viewer_rsvp_status
		FROM events e
		JOIN circles c ON c.id = e.circle_id
		JOIN profiles p ON p.id = e.organizer_profile_id
		LEFT JOIN event_rsvps er ON er.event_id = e.id AND er.profile_id = $2
		WHERE e.id = $1
		AND e.deleted_at IS NULL`

	var details EventDetails
	var event domain.Event
	var description, location, timezone *string
	var endTime *time.Time
	var capacity *int

	err := r.pool.QueryRow(ctx, query, eventID, viewerProfileID).Scan(
		&event.ID,
		&event.CircleID,
		&event.OrganizerProfileID,
		&event.Title,
		&description,
		&location,
		&timezone,
		&event.StartTime,
		&endTime,
		&event.RequiresTicket,
		&capacity,
		&event.CreatedAt,
		&event.UpdatedAt,
		&details.CircleName,
		&details.OrganizerName,
		&details.ViewerRSVPStatus,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if description != nil {
		event.Description = *description
	}
	if location != nil {
		event.Location = *location
	}
	if timezone != nil {
		event.Timezone = *timezone
	}
	if endTime != nil {
		event.EndTime = *endTime
	}
	if capacity != nil {
		event.Capacity = *capacity
	}

	details.Event = event

	// Get attendees
	attendees, err := r.getEventAttendees(ctx, eventID)
	if err != nil {
		return nil, err
	}
	details.Attendees = attendees

	// Get coordination needs
	needs, err := r.GetCoordinationNeeds(ctx, eventID)
	if err != nil {
		return nil, err
	}
	details.CoordinationNeeds = needs

	return &details, nil
}

func (r *PostgresRepository) getEventAttendees(ctx context.Context, eventID string) ([]Attendee, error) {
	query := `
		SELECT p.id, p.handle, er.status, p.avatar_url
		FROM event_rsvps er
		JOIN profiles p ON p.id = er.profile_id
		WHERE er.event_id = $1
		AND er.status IN ('attending', 'maybe')
		ORDER BY er.created_at ASC`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendees []Attendee
	for rows.Next() {
		var a Attendee
		if err := rows.Scan(&a.ProfileID, &a.Name, &a.RSVPStatus, &a.AvatarURL); err != nil {
			return nil, err
		}
		attendees = append(attendees, a)
	}

	return attendees, rows.Err()
}

func (r *PostgresRepository) UpdateEvent(ctx context.Context, event *domain.Event) error {
	query := `
		UPDATE events
		SET title = $2,
		    description = $3,
		    location = $4,
		    timezone = $5,
		    start_time = $6,
		    end_time = $7,
		    requires_ticket = $8,
		    capacity = $9,
		    updated_at = $10
		WHERE id = $1
		AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.Title,
		event.Description,
		event.Location,
		event.Timezone,
		event.StartTime,
		event.EndTime,
		event.RequiresTicket,
		event.Capacity,
		time.Now(),
	)

	return err
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, eventID string) error {
	query := `
		UPDATE events
		SET deleted_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *PostgresRepository) GetEventsByCircle(ctx context.Context, circleID string) ([]domain.Event, error) {
	query := `
		SELECT id, circle_id, organizer_profile_id, title, location, timezone,
		       start_time, end_time, requires_ticket, capacity, created_at, updated_at
		FROM events
		WHERE circle_id = $1
		AND deleted_at IS NULL
		AND start_time > NOW()
		ORDER BY start_time ASC`

	rows, err := r.pool.Query(ctx, query, circleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		var location, timezone *string
		var endTime *time.Time
		var capacity *int

		if err := rows.Scan(
			&event.ID,
			&event.CircleID,
			&event.OrganizerProfileID,
			&event.Title,
			&location,
			&timezone,
			&event.StartTime,
			&endTime,
			&event.RequiresTicket,
			&capacity,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if location != nil {
			event.Location = *location
		}
		if timezone != nil {
			event.Timezone = *timezone
		}
		if endTime != nil {
			event.EndTime = *endTime
		}
		if capacity != nil {
			event.Capacity = *capacity
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *PostgresRepository) CreateCoordinationNeed(ctx context.Context, need *CoordinationNeed) error {
	query := `
		INSERT INTO event_coordination_needs (event_id, type, message, author_profile_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, need.EventID, need.Type, need.Message, need.AuthorProfileID).
		Scan(&need.ID, &need.CreatedAt)

	return err
}

func (r *PostgresRepository) ResolveCoordinationNeed(ctx context.Context, needID string) error {
	query := `
		UPDATE event_coordination_needs
		SET resolved_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, needID)
	return err
}

func (r *PostgresRepository) GetCoordinationNeeds(ctx context.Context, eventID string) ([]CoordinationNeed, error) {
	query := `
		SELECT ecn.id, ecn.event_id, ecn.type, ecn.message, ecn.author_profile_id,
		       p.handle as author_name, ecn.created_at, ecn.resolved_at
		FROM event_coordination_needs ecn
		JOIN profiles p ON p.id = ecn.author_profile_id
		WHERE ecn.event_id = $1
		ORDER BY ecn.created_at DESC`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var needs []CoordinationNeed
	for rows.Next() {
		var need CoordinationNeed
		var createdAt time.Time
		var resolvedAt *time.Time

		if err := rows.Scan(
			&need.ID,
			&need.EventID,
			&need.Type,
			&need.Message,
			&need.AuthorProfileID,
			&need.AuthorName,
			&createdAt,
			&resolvedAt,
		); err != nil {
			return nil, err
		}

		need.CreatedAt = createdAt.Format(time.RFC3339)
		if resolvedAt != nil {
			resolvedAtStr := resolvedAt.Format(time.RFC3339)
			need.ResolvedAt = &resolvedAtStr
		}

		needs = append(needs, need)
	}

	return needs, rows.Err()
}
