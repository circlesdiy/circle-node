package content

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"circles.diy/internal/domain"
)

type ReactionRepository struct {
	pool *pgxpool.Pool
}

func NewReactionRepository(pool *pgxpool.Pool) *ReactionRepository {
	return &ReactionRepository{
		pool: pool,
	}
}

func (r *ReactionRepository) CreateReaction(ctx context.Context, reaction *domain.Reaction) error {
	query := `
		INSERT INTO reactions (
			id, target_type, target_id, profile_id, key, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		reaction.ID,
		reaction.TargetType,
		reaction.TargetID,
		reaction.ProfileID,
		reaction.Key,
		reaction.CreatedAt,
	)

	return err
}

func (r *ReactionRepository) GetReactionByID(ctx context.Context, reactionID string) (*domain.Reaction, error) {
	query := `
		SELECT id, target_type, target_id, profile_id, key, created_at
		FROM reactions
		WHERE id = $1
	`

	var reaction domain.Reaction

	err := r.pool.QueryRow(ctx, query, reactionID).Scan(
		&reaction.ID,
		&reaction.TargetType,
		&reaction.TargetID,
		&reaction.ProfileID,
		&reaction.Key,
		&reaction.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &reaction, nil
}

func (r *ReactionRepository) GetReactionsByTarget(ctx context.Context, targetType, targetID string) ([]*domain.Reaction, error) {
	query := `
		SELECT id, target_type, target_id, profile_id, key, created_at
		FROM reactions
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*domain.Reaction
	for rows.Next() {
		var reaction domain.Reaction

		err := rows.Scan(
			&reaction.ID,
			&reaction.TargetType,
			&reaction.TargetID,
			&reaction.ProfileID,
			&reaction.Key,
			&reaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		reactions = append(reactions, &reaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (r *ReactionRepository) GetReactionsByProfileID(ctx context.Context, profileID string) ([]*domain.Reaction, error) {
	query := `
		SELECT id, target_type, target_id, profile_id, key, created_at
		FROM reactions
		WHERE profile_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*domain.Reaction
	for rows.Next() {
		var reaction domain.Reaction

		err := rows.Scan(
			&reaction.ID,
			&reaction.TargetType,
			&reaction.TargetID,
			&reaction.ProfileID,
			&reaction.Key,
			&reaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		reactions = append(reactions, &reaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (r *ReactionRepository) GetReactionByProfileAndTarget(ctx context.Context, profileID, targetType, targetID string) (*domain.Reaction, error) {
	query := `
		SELECT id, target_type, target_id, profile_id, key, created_at
		FROM reactions
		WHERE profile_id = $1 AND target_type = $2 AND target_id = $3
	`

	var reaction domain.Reaction

	err := r.pool.QueryRow(ctx, query, profileID, targetType, targetID).Scan(
		&reaction.ID,
		&reaction.TargetType,
		&reaction.TargetID,
		&reaction.ProfileID,
		&reaction.Key,
		&reaction.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &reaction, nil
}

func (r *ReactionRepository) DeleteReaction(ctx context.Context, reactionID string) error {
	query := `
		DELETE FROM reactions
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, reactionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("reaction not found")
	}

	return nil
}

func (r *ReactionRepository) GetReactionCounts(ctx context.Context, targetType, targetID string) (map[string]int, error) {
	query := `
		SELECT key, COUNT(*) as count
		FROM reactions
		WHERE target_type = $1 AND target_id = $2
		GROUP BY key
	`

	rows, err := r.pool.Query(ctx, query, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var key string
		var count int

		err := rows.Scan(&key, &count)
		if err != nil {
			return nil, err
		}

		counts[key] = count
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}
