package dashboard

import (
	"context"
	"html/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetCoordinationNeeds(ctx context.Context, profileID string) ([]CoordinationNeed, error) {
	query := `
		SELECT ecn.event_id, ecn.type, ecn.message, p.handle, c.name, ecn.created_at
		FROM event_coordination_needs ecn
		JOIN events e ON e.id = ecn.event_id
		JOIN circles c ON c.id = e.circle_id
		JOIN circle_memberships cm ON cm.circle_id = c.id
		JOIN profiles p ON p.id = ecn.author_profile_id
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND ecn.resolved_at IS NULL
		AND e.start_time > NOW()
		ORDER BY ecn.created_at DESC
		LIMIT 5`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var needs []CoordinationNeed
	for rows.Next() {
		var n CoordinationNeed
		if err := rows.Scan(&n.EventID, &n.Type, &n.Message, &n.AuthorName, &n.CircleName, &n.CreatedAt); err != nil {
			return nil, err
		}
		needs = append(needs, n)
	}

	return needs, rows.Err()
}

func (r *PostgresRepository) GetUnreadMessagesSummary(ctx context.Context, profileID string) (*MessagesSummary, error) {
	query := `
		WITH unread_messages AS (
			SELECT DISTINCT m.sender_profile_id, m.chat_id, p.handle
			FROM messages m
			JOIN chat_participants cp ON cp.chat_id = m.chat_id
			LEFT JOIN message_reads mr ON mr.message_id = m.id AND mr.profile_id = $1
			JOIN profiles p ON p.id = m.sender_profile_id
			WHERE cp.profile_id = $1
			AND m.sender_profile_id != $1
			AND mr.id IS NULL
			AND m.deleted_at IS NULL
		),
		sender_circles AS (
			SELECT um.sender_profile_id, um.handle,
			       STRING_AGG(DISTINCT c.name, ', ') as shared_circles
			FROM unread_messages um
			JOIN circle_memberships cm1 ON cm1.profile_id = um.sender_profile_id
			JOIN circle_memberships cm2 ON cm2.circle_id = cm1.circle_id AND cm2.profile_id = $1
			JOIN circles c ON c.id = cm1.circle_id
			WHERE cm1.state = 'active' AND cm2.state = 'active'
			GROUP BY um.sender_profile_id, um.handle
		)
		SELECT COUNT(DISTINCT um.sender_profile_id) as sender_count,
		       COUNT(*) as message_count
		FROM unread_messages um`

	var summary MessagesSummary
	var senderCount, messageCount int
	if err := r.pool.QueryRow(ctx, query, profileID).Scan(&senderCount, &messageCount); err != nil {
		if err == pgx.ErrNoRows {
			return &MessagesSummary{}, nil
		}
		return nil, err
	}

	summary.UnreadCount = messageCount

	if senderCount > 0 {
		sendersQuery := `
			WITH unread_messages AS (
				SELECT DISTINCT m.sender_profile_id, p.handle
				FROM messages m
				JOIN chat_participants cp ON cp.chat_id = m.chat_id
				LEFT JOIN message_reads mr ON mr.message_id = m.id AND mr.profile_id = $1
				JOIN profiles p ON p.id = m.sender_profile_id
				WHERE cp.profile_id = $1
				AND m.sender_profile_id != $1
				AND mr.id IS NULL
				AND m.deleted_at IS NULL
			)
			SELECT um.handle, STRING_AGG(DISTINCT c.name, ', ') as shared_circles
			FROM unread_messages um
			JOIN circle_memberships cm1 ON cm1.profile_id = um.sender_profile_id
			JOIN circle_memberships cm2 ON cm2.circle_id = cm1.circle_id AND cm2.profile_id = $1
			JOIN circles c ON c.id = cm1.circle_id
			WHERE cm1.state = 'active' AND cm2.state = 'active'
			GROUP BY um.handle
			LIMIT 5`

		rows, err := r.pool.Query(ctx, sendersQuery, profileID)
		if err != nil {
			return &summary, nil
		}
		defer rows.Close()

		for rows.Next() {
			var sender MessageSender
			var circlesStr string
			if err := rows.Scan(&sender.Name, &circlesStr); err != nil {
				continue
			}
			sender.SharedCircles = []string{circlesStr}
			summary.Senders = append(summary.Senders, sender)
		}
	}

	return &summary, nil
}

func (r *PostgresRepository) GetUpcomingEvents(ctx context.Context, profileID string, days int) ([]EventWithDetails, error) {
	query := `
		SELECT e.id, e.title, e.start_time, c.id, c.name,
		       COALESCE(c.icon, ''), COALESCE(c.icon_bg_color, ''), COALESCE(c.avatar_url, ''),
		       COUNT(DISTINCT er.id) FILTER (WHERE er.status = 'going') as attendee_count,
		       COALESCE(user_rsvp.status, 'not_responded') as user_status
		FROM events e
		JOIN circles c ON c.id = e.circle_id
		JOIN circle_memberships cm ON cm.circle_id = c.id
		LEFT JOIN event_rsvps er ON er.event_id = e.id AND er.status = 'going'
		LEFT JOIN event_rsvps user_rsvp ON user_rsvp.event_id = e.id AND user_rsvp.profile_id = $1
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND e.deleted_at IS NULL
		AND e.start_time BETWEEN NOW() AND NOW() + INTERVAL '1 day' * $2
		GROUP BY e.id, e.title, e.start_time, c.id, c.name, c.icon, c.icon_bg_color, c.avatar_url, user_rsvp.status
		ORDER BY e.start_time ASC
		LIMIT 10`

	rows, err := r.pool.Query(ctx, query, profileID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventWithDetails
	for rows.Next() {
		var e EventWithDetails
		var circleBgColor string
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &e.CircleID, &e.CircleName,
			&e.CircleIcon, &circleBgColor, &e.CircleAvatar, &e.AttendeeCount, &e.RSVPStatus); err != nil {
			return nil, err
		}
		e.CircleBgColor = template.CSS(circleBgColor)

		coordQuery := `
			SELECT ecn.type, ecn.message, p.handle
			FROM event_coordination_needs ecn
			JOIN profiles p ON p.id = ecn.author_profile_id
			WHERE ecn.event_id = $1
			AND ecn.resolved_at IS NULL
			ORDER BY ecn.created_at DESC
			LIMIT 1`

		var coord CoordinationNeed
		if err := r.pool.QueryRow(ctx, coordQuery, e.ID).Scan(&coord.Type, &coord.Message, &coord.AuthorName); err == nil {
			e.Coordination = &coord
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (r *PostgresRepository) GetUserCirclesWithActivity(ctx context.Context, profileID string) ([]CircleActivity, error) {
	query := `
		SELECT c.id, c.name, COALESCE(c.icon, ''), COALESCE(c.icon_bg_color, ''),
		       COALESCE(c.avatar_url, ''), COALESCE(c.banner_url, ''),
		       COALESCE(MAX(a.occurred_at), c.created_at) as last_activity,
		       next_event.start_time as next_event_time
		FROM circles c
		JOIN circle_memberships cm ON cm.circle_id = c.id
		LEFT JOIN activities a ON a.circle_id = c.id AND a.occurred_at > NOW() - INTERVAL '30 days'
		LEFT JOIN LATERAL (
			SELECT start_time
			FROM events
			WHERE circle_id = c.id
			AND deleted_at IS NULL
			AND start_time > NOW()
			ORDER BY start_time ASC
			LIMIT 1
		) next_event ON true
		WHERE cm.profile_id = $1
		AND cm.state = 'active'
		AND c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.icon, c.icon_bg_color, c.avatar_url, c.banner_url, next_event.start_time
		ORDER BY last_activity DESC
		LIMIT 10`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var circles []CircleActivity
	for rows.Next() {
		var c CircleActivity
		var iconBgColor string
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &iconBgColor, &c.AvatarURL, &c.BannerURL,
			&c.LastActivity, &c.NextEventTime); err != nil {
			return nil, err
		}
		c.IconBgColor = template.CSS(iconBgColor)
		circles = append(circles, c)
	}

	return circles, rows.Err()
}

func (r *PostgresRepository) GetSerendipityRecommendations(ctx context.Context, profileID string) (*Recommendation, error) {
	return nil, nil
}

func (r *PostgresRepository) GetRecentUpdates(ctx context.Context, profileID string, since time.Duration) ([]Activity, error) {
	return nil, nil
}

func (r *PostgresRepository) GetPendingInvitations(ctx context.Context, profileID string) ([]PendingInvitation, error) {
	query := `
		SELECT cm.id, c.id, c.name, COALESCE(c.icon, ''), COALESCE(c.icon_bg_color, ''),
		       COALESCE(c.avatar_url, ''), cm.created_at,
		       owner_profile.id, COALESCE(owner_profile.name, owner_profile.handle, 'Someone'),
		       COALESCE(owner_profile.handle, '')
		FROM circle_memberships cm
		JOIN circles c ON c.id = cm.circle_id
		JOIN profiles owner_profile ON owner_profile.id = c.owner_profile_id
		WHERE cm.profile_id = $1
		AND cm.state = 'invited'
		AND c.deleted_at IS NULL
		ORDER BY cm.created_at DESC
		LIMIT 20`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []PendingInvitation
	for rows.Next() {
		var inv PendingInvitation
		var iconBgColor string
		if err := rows.Scan(&inv.MembershipID, &inv.CircleID, &inv.CircleName, &inv.CircleIcon,
			&iconBgColor, &inv.CircleAvatar, &inv.InvitedAt,
			&inv.InviterID, &inv.InviterName, &inv.InviterHandle); err != nil {
			return nil, err
		}
		inv.CircleBgColor = template.CSS(iconBgColor)
		invitations = append(invitations, inv)
	}

	return invitations, rows.Err()
}
