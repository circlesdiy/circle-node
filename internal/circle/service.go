package circle

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles circle business logic
type Service struct {
	repo    domain.CircleRepository
	storage domain.FileStorage
	logger  *zap.Logger
}

// NewService creates a new circle service
func NewService(repo domain.CircleRepository, storage domain.FileStorage, logger *zap.Logger) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

// CreateCircle creates a new circle with automatic owner membership
func (s *Service) CreateCircle(ctx context.Context, name, description, visibility string, autoModEnabled bool, ownerProfileID string) (*domain.Circle, error) {
	// Validate inputs
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("circle name cannot be empty")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("circle name cannot exceed 255 characters")
	}

	// Validate visibility
	if !isValidVisibility(visibility) {
		return nil, fmt.Errorf("invalid visibility: %s", visibility)
	}

	s.logger.Debug("creating circle",
		zap.String("name", name),
		zap.String("owner_profile_id", ownerProfileID),
		zap.String("visibility", visibility),
	)

	// Create circle entity
	now := time.Now().UTC()
	circle := &domain.Circle{
		ID:             uuid.New().String(),
		OwnerProfileID: ownerProfileID,
		Name:           name,
		Description:    description,
		Visibility:     visibility,
		AutoModEnabled: autoModEnabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create owner membership
	membership := &domain.CircleMembership{
		ID:        uuid.New().String(),
		CircleID:  circle.ID,
		ProfileID: ownerProfileID,
		State:     domain.MembershipStateActive,
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Atomic creation
	if err := s.repo.CreateCircleWithOwnership(ctx, circle, membership); err != nil {
		s.logger.Error("failed to create circle",
			zap.String("circle_id", circle.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create circle: %w", err)
	}

	s.logger.Info("circle created successfully",
		zap.String("circle_id", circle.ID),
		zap.String("name", name),
		zap.String("owner_profile_id", ownerProfileID),
	)

	return circle, nil
}

// GetCircleByID retrieves a circle by ID
func (s *Service) GetCircleByID(ctx context.Context, id string) (*domain.Circle, error) {
	s.logger.Debug("getting circle by ID",
		zap.String("circle_id", id),
	)

	circle, err := s.repo.GetCircleByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get circle",
			zap.String("circle_id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get circle: %w", err)
	}

	if circle == nil {
		s.logger.Debug("circle not found",
			zap.String("circle_id", id),
		)
		return nil, nil
	}

	return circle, nil
}

// UpdateCircle updates a circle (owner only)
func (s *Service) UpdateCircle(ctx context.Context, circleID string, updates map[string]interface{}, requesterProfileID string) error {
	// Get existing circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check ownership
	if circle.OwnerProfileID != requesterProfileID {
		s.logger.Warn("unauthorized circle update attempt",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
		)
		return fmt.Errorf("only the owner can update the circle")
	}

	s.logger.Debug("updating circle",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID),
	)

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("circle name cannot be empty")
		}
		if len(name) > 255 {
			return fmt.Errorf("circle name cannot exceed 255 characters")
		}
		circle.Name = name
	}

	if description, ok := updates["description"].(string); ok {
		circle.Description = description
	}

	if visibility, ok := updates["visibility"].(string); ok {
		if !isValidVisibility(visibility) {
			return fmt.Errorf("invalid visibility: %s", visibility)
		}
		circle.Visibility = visibility
	}

	if autoMod, ok := updates["auto_mod_enabled"].(bool); ok {
		circle.AutoModEnabled = autoMod
	}

	circle.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCircle(ctx, circle); err != nil {
		s.logger.Error("failed to update circle",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update circle: %w", err)
	}

	s.logger.Info("circle updated successfully",
		zap.String("circle_id", circleID),
	)

	return nil
}

// DeleteCircle soft deletes a circle (owner only)
func (s *Service) DeleteCircle(ctx context.Context, circleID, requesterProfileID string) error {
	// Get existing circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check ownership
	if circle.OwnerProfileID != requesterProfileID {
		s.logger.Warn("unauthorized circle deletion attempt",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
		)
		return fmt.Errorf("only the owner can delete the circle")
	}

	s.logger.Debug("deleting circle",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID),
	)

	if err := s.repo.DeleteCircle(ctx, circleID); err != nil {
		s.logger.Error("failed to delete circle",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete circle: %w", err)
	}

	s.logger.Info("circle deleted successfully",
		zap.String("circle_id", circleID),
	)

	return nil
}

// GetUserCircles retrieves all circles a user owns or is a member of
func (s *Service) GetUserCircles(ctx context.Context, profileID string) ([]domain.Circle, error) {
	s.logger.Debug("getting user circles",
		zap.String("profile_id", profileID),
	)

	// Get owned circles
	ownedCircles, err := s.repo.GetCirclesByOwnerID(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get owned circles",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get owned circles: %w", err)
	}

	// Get member circles
	memberCircles, err := s.repo.GetCirclesByMemberID(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get member circles",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get member circles: %w", err)
	}

	// Combine and deduplicate
	circleMap := make(map[string]domain.Circle)
	for _, circle := range ownedCircles {
		circleMap[circle.ID] = circle
	}
	for _, circle := range memberCircles {
		circleMap[circle.ID] = circle
	}

	circles := make([]domain.Circle, 0, len(circleMap))
	for _, circle := range circleMap {
		circles = append(circles, circle)
	}

	s.logger.Debug("retrieved user circles",
		zap.String("profile_id", profileID),
		zap.Int("count", len(circles)),
	)

	return circles, nil
}

// GetPublicCircles retrieves public circles with pagination
func (s *Service) GetPublicCircles(ctx context.Context, limit, offset int) ([]domain.Circle, error) {
	s.logger.Debug("getting public circles",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	circles, err := s.repo.GetPublicCircles(ctx, limit, offset)
	if err != nil {
		s.logger.Error("failed to get public circles",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get public circles: %w", err)
	}

	return circles, nil
}

// InviteMember invites a profile to join a circle
func (s *Service) InviteMember(ctx context.Context, circleID, profileID, inviterProfileID string) error {
	// Check if circle exists
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if inviter can invite (owner or admin)
	canInvite, err := s.CanInvite(ctx, circleID, inviterProfileID)
	if err != nil {
		return fmt.Errorf("failed to check invite permission: %w", err)
	}
	if !canInvite {
		s.logger.Warn("unauthorized invitation attempt",
			zap.String("circle_id", circleID),
			zap.String("inviter_profile_id", inviterProfileID),
		)
		return fmt.Errorf("only owners and admins can invite members")
	}

	// Check if membership already exists
	existing, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to check existing membership: %w", err)
	}

	if existing != nil {
		if existing.State == domain.MembershipStateBanned {
			return fmt.Errorf("cannot invite banned member")
		}
		if existing.State == domain.MembershipStateActive {
			return fmt.Errorf("profile is already a member")
		}
		if existing.State == domain.MembershipStateInvited {
			return fmt.Errorf("invitation already pending")
		}
	}

	s.logger.Debug("inviting member to circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
		zap.String("inviter_profile_id", inviterProfileID),
	)

	// Create invitation
	now := time.Now().UTC()
	membership := &domain.CircleMembership{
		ID:               uuid.New().String(),
		CircleID:         circleID,
		ProfileID:        profileID,
		InviterProfileID: &inviterProfileID,
		State:            domain.MembershipStateInvited,
		JoinedAt:         now, // Will be updated when accepted
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateMembership(ctx, membership); err != nil {
		s.logger.Error("failed to create invitation",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	s.logger.Info("member invited to circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// AcceptInvitation accepts a circle invitation
func (s *Service) AcceptInvitation(ctx context.Context, circleID, profileID string) error {
	// Get membership
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}
	if membership == nil {
		return fmt.Errorf("invitation not found")
	}

	if membership.State != domain.MembershipStateInvited {
		return fmt.Errorf("no pending invitation found")
	}

	s.logger.Debug("accepting invitation",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Update membership to active
	now := time.Now().UTC()
	membership.State = domain.MembershipStateActive
	membership.JoinedAt = now
	membership.UpdatedAt = now

	if err := s.repo.UpdateMembership(ctx, membership); err != nil {
		s.logger.Error("failed to accept invitation",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	s.logger.Info("invitation accepted",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// DeclineInvitation declines a circle invitation
func (s *Service) DeclineInvitation(ctx context.Context, circleID, profileID string) error {
	// Get membership
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}
	if membership == nil {
		return fmt.Errorf("invitation not found")
	}

	// Verify state is invited
	if membership.State != domain.MembershipStateInvited {
		return fmt.Errorf("no pending invitation found")
	}

	// Delete the invitation (membership record)
	if err := s.repo.DeleteMembership(ctx, membership.ID); err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}

	s.logger.Info("invitation declined",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// JoinPublicCircle allows a user to directly join a public or unlisted circle
func (s *Service) JoinPublicCircle(ctx context.Context, circleID, profileID string) error {
	// Get circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Only allow joining public or unlisted circles directly
	if circle.Visibility == domain.CircleVisibilityPrivate {
		return fmt.Errorf("cannot join private circles without an invitation")
	}

	// Check if membership already exists
	existing, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to check existing membership: %w", err)
	}

	if existing != nil {
		if existing.State == domain.MembershipStateBanned {
			return fmt.Errorf("you are banned from this circle")
		}
		if existing.State == domain.MembershipStateActive {
			return fmt.Errorf("you are already a member of this circle")
		}
		if existing.State == domain.MembershipStateInvited {
			// If they have a pending invitation, accept it instead
			return s.AcceptInvitation(ctx, circleID, profileID)
		}
		// If they previously left, allow them to rejoin by updating the existing membership
		if existing.State == domain.MembershipStateLeft {
			s.logger.Debug("rejoining circle",
				zap.String("circle_id", circleID),
				zap.String("profile_id", profileID),
			)

			now := time.Now().UTC()
			existing.State = domain.MembershipStateActive
			existing.JoinedAt = now
			existing.LeftAt = nil
			existing.UpdatedAt = now

			if err := s.repo.UpdateMembership(ctx, existing); err != nil {
				s.logger.Error("failed to rejoin circle",
					zap.String("circle_id", circleID),
					zap.String("profile_id", profileID),
					zap.Error(err),
				)
				return fmt.Errorf("failed to rejoin circle: %w", err)
			}

			s.logger.Info("rejoined circle",
				zap.String("circle_id", circleID),
				zap.String("profile_id", profileID),
			)

			return nil
		}
	}

	// Create new active membership
	s.logger.Debug("joining public circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	now := time.Now().UTC()
	membership := &domain.CircleMembership{
		ID:        uuid.New().String(),
		CircleID:  circleID,
		ProfileID: profileID,
		State:     domain.MembershipStateActive,
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateMembership(ctx, membership); err != nil {
		s.logger.Error("failed to join circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to join circle: %w", err)
	}

	s.logger.Info("joined public circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// LeaveCircle allows a member to leave a circle
func (s *Service) LeaveCircle(ctx context.Context, circleID, profileID string) error {
	// Get circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Prevent owner from leaving
	if circle.OwnerProfileID == profileID {
		return fmt.Errorf("owner cannot leave the circle, delete it instead")
	}

	// Get membership
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}
	if membership == nil {
		return fmt.Errorf("membership not found")
	}

	if membership.State != domain.MembershipStateActive {
		return fmt.Errorf("only active members can leave")
	}

	s.logger.Debug("member leaving circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Update membership state
	now := time.Now().UTC()
	membership.State = domain.MembershipStateLeft
	membership.LeftAt = &now
	membership.UpdatedAt = now

	if err := s.repo.UpdateMembership(ctx, membership); err != nil {
		s.logger.Error("failed to leave circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to leave circle: %w", err)
	}

	s.logger.Info("member left circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// BanMember bans a member from a circle (owner only)
func (s *Service) BanMember(ctx context.Context, circleID, profileID, bannerProfileID string) error {
	// Get circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if banner is owner
	if circle.OwnerProfileID != bannerProfileID {
		s.logger.Warn("unauthorized ban attempt",
			zap.String("circle_id", circleID),
			zap.String("banner_profile_id", bannerProfileID),
		)
		return fmt.Errorf("only the owner can ban members")
	}

	// Cannot ban yourself
	if profileID == bannerProfileID {
		return fmt.Errorf("cannot ban yourself")
	}

	// Get membership
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}
	if membership == nil {
		return fmt.Errorf("membership not found")
	}

	if membership.State == domain.MembershipStateBanned {
		return fmt.Errorf("member is already banned")
	}

	s.logger.Debug("banning member from circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
		zap.String("banner_profile_id", bannerProfileID),
	)

	// Update membership state
	now := time.Now().UTC()
	membership.State = domain.MembershipStateBanned
	membership.LeftAt = &now
	membership.UpdatedAt = now

	if err := s.repo.UpdateMembership(ctx, membership); err != nil {
		s.logger.Error("failed to ban member",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to ban member: %w", err)
	}

	s.logger.Info("member banned from circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// RemoveMember removes a member from a circle (owner only)
func (s *Service) RemoveMember(ctx context.Context, circleID, profileID, removerProfileID string) error {
	// Get circle
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if remover is owner
	if circle.OwnerProfileID != removerProfileID {
		s.logger.Warn("unauthorized removal attempt",
			zap.String("circle_id", circleID),
			zap.String("remover_profile_id", removerProfileID),
		)
		return fmt.Errorf("only the owner can remove members")
	}

	// Cannot remove yourself
	if profileID == removerProfileID {
		return fmt.Errorf("cannot remove yourself")
	}

	// Get membership
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return fmt.Errorf("failed to get membership: %w", err)
	}
	if membership == nil {
		return fmt.Errorf("membership not found")
	}

	s.logger.Debug("removing member from circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
		zap.String("remover_profile_id", removerProfileID),
	)

	if err := s.repo.DeleteMembership(ctx, membership.ID); err != nil {
		s.logger.Error("failed to remove member",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove member: %w", err)
	}

	s.logger.Info("member removed from circle",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	return nil
}

// GetCircleMembers retrieves members of a circle with pagination
func (s *Service) GetCircleMembers(ctx context.Context, circleID string, limit, offset int) ([]domain.CircleMembership, error) {
	s.logger.Debug("getting circle members",
		zap.String("circle_id", circleID),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	members, err := s.repo.GetCircleMembersByCircleID(ctx, circleID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get circle members",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get circle members: %w", err)
	}

	return members, nil
}

// GetPendingInvitations retrieves pending invitations for a profile
func (s *Service) GetPendingInvitations(ctx context.Context, profileID string) ([]domain.CircleMembership, error) {
	s.logger.Debug("getting pending invitations",
		zap.String("profile_id", profileID),
	)

	invitations, err := s.repo.GetPendingInvitationsByProfileID(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get pending invitations",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}

	return invitations, nil
}

// GetPendingInvitationsWithInviter retrieves pending invitations with inviter profile information
func (s *Service) GetPendingInvitationsWithInviter(ctx context.Context, profileID string) ([]domain.CircleMembershipWithInviter, error) {
	s.logger.Debug("getting pending invitations with inviter info",
		zap.String("profile_id", profileID),
	)

	invitations, err := s.repo.GetPendingInvitationsWithInviterByProfileID(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get pending invitations with inviter",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}

	return invitations, nil
}

// IsOwner checks if a profile is the owner of a circle
func (s *Service) IsOwner(ctx context.Context, circleID, profileID string) (bool, error) {
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return false, err
	}
	if circle == nil {
		return false, nil
	}
	return circle.OwnerProfileID == profileID, nil
}

// IsMember checks if a profile is an active member of a circle
func (s *Service) IsMember(ctx context.Context, circleID, profileID string) (bool, error) {
	membership, err := s.repo.GetMembershipByCircleAndProfile(ctx, circleID, profileID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}
	return membership.IsActive(), nil
}

// CanView checks if a profile can view a circle based on visibility
func (s *Service) CanView(ctx context.Context, circleID, viewerProfileID string) (bool, error) {
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return false, err
	}
	if circle == nil {
		return false, nil
	}

	// Public circles can be viewed by anyone
	if circle.IsPublic() {
		return true, nil
	}

	// Private/unlisted circles require membership
	isMember, err := s.IsMember(ctx, circleID, viewerProfileID)
	if err != nil {
		return false, err
	}

	isOwner, err := s.IsOwner(ctx, circleID, viewerProfileID)
	if err != nil {
		return false, err
	}

	return isMember || isOwner, nil
}

// CountMembers counts active members in a circle
func (s *Service) CountMembers(ctx context.Context, circleID string) (int, error) {
	count, err := s.repo.CountMembersByCircleID(ctx, circleID)
	if err != nil {
		s.logger.Error("failed to count members",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to count members: %w", err)
	}
	return count, nil
}

// IsAdmin checks if a profile is an admin of a circle
// Note: Currently we don't have a roles table implemented yet, so this returns false
// In the future, this will check the profile_roles table
func (s *Service) IsAdmin(ctx context.Context, circleID, profileID string) (bool, error) {
	// TODO: Implement admin role checking once roles are fully implemented
	// For now, check if they're a member but not owner
	return false, nil
}

// CanInvite checks if a profile can invite members (owners and admins)
func (s *Service) CanInvite(ctx context.Context, circleID, profileID string) (bool, error) {
	// Check if owner
	isOwner, err := s.IsOwner(ctx, circleID, profileID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// Check if admin
	isAdmin, err := s.IsAdmin(ctx, circleID, profileID)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

// CanEditSettings checks if a profile can edit circle settings (owner only)
func (s *Service) CanEditSettings(ctx context.Context, circleID, profileID string) (bool, error) {
	return s.IsOwner(ctx, circleID, profileID)
}

// GetMemberRole returns the role of a member in a circle (owner, admin, member)
func (s *Service) GetMemberRole(ctx context.Context, circleID, profileID string) (string, error) {
	isOwner, err := s.IsOwner(ctx, circleID, profileID)
	if err != nil {
		return "", err
	}
	if isOwner {
		return "owner", nil
	}

	isAdmin, err := s.IsAdmin(ctx, circleID, profileID)
	if err != nil {
		return "", err
	}
	if isAdmin {
		return "admin", nil
	}

	isMember, err := s.IsMember(ctx, circleID, profileID)
	if err != nil {
		return "", err
	}
	if isMember {
		return "member", nil
	}

	return "", nil
}

// UpdateAvatar uploads a new avatar and updates the circle
func (s *Service) UpdateAvatar(ctx context.Context, circleID string, file multipart.File, header *multipart.FileHeader, requesterProfileID string) error {
	s.logger.Debug("updating circle avatar",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID),
		zap.String("filename", header.Filename))

	// Get current circle to verify ownership and retrieve old avatar URL
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if requester is the owner
	if circle.OwnerProfileID != requesterProfileID {
		return fmt.Errorf("only circle owner can update avatar")
	}

	oldAvatarURL := circle.AvatarURL

	// Upload new avatar using storage interface (reuses profile storage with circleID)
	avatarURL, err := s.storage.SaveAvatar(ctx, circleID, file, header)
	if err != nil {
		s.logger.Error("failed to save circle avatar",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to save avatar: %w", err)
	}

	// Update circle with new avatar URL
	circle.AvatarURL = avatarURL
	circle.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCircle(ctx, circle); err != nil {
		// Rollback: delete newly uploaded avatar
		if deleteErr := s.storage.DeleteAvatar(ctx, circleID); deleteErr != nil {
			s.logger.Warn("failed to delete avatar after DB update failure",
				zap.String("circle_id", circleID),
				zap.Error(deleteErr))
		}
		s.logger.Error("failed to update circle with new avatar",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to update circle: %w", err)
	}

	// Delete old avatar file if it exists and is different
	if oldAvatarURL != "" && oldAvatarURL != avatarURL {
		if err := s.storage.DeleteAvatar(ctx, circleID); err != nil {
			s.logger.Warn("failed to delete old circle avatar",
				zap.String("circle_id", circleID),
				zap.String("old_url", oldAvatarURL),
				zap.Error(err))
		}
	}

	s.logger.Info("circle avatar updated successfully",
		zap.String("circle_id", circleID),
		zap.String("new_url", avatarURL))

	return nil
}

// UpdateBanner uploads a new banner and updates the circle
func (s *Service) UpdateBanner(ctx context.Context, circleID string, file multipart.File, header *multipart.FileHeader, requesterProfileID string) error {
	s.logger.Debug("updating circle banner",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID),
		zap.String("filename", header.Filename))

	// Get current circle to verify ownership and retrieve old banner URL
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if requester is the owner
	if circle.OwnerProfileID != requesterProfileID {
		return fmt.Errorf("only circle owner can update banner")
	}

	oldBannerURL := circle.BannerURL

	// Upload new banner using storage interface (reuses profile storage with circleID)
	bannerURL, err := s.storage.SaveBanner(ctx, circleID, file, header)
	if err != nil {
		s.logger.Error("failed to save circle banner",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to save banner: %w", err)
	}

	// Update circle with new banner URL
	circle.BannerURL = bannerURL
	circle.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCircle(ctx, circle); err != nil {
		// Rollback: delete newly uploaded banner
		if deleteErr := s.storage.DeleteBanner(ctx, circleID); deleteErr != nil {
			s.logger.Warn("failed to delete banner after DB update failure",
				zap.String("circle_id", circleID),
				zap.Error(deleteErr))
		}
		s.logger.Error("failed to update circle with new banner",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to update circle: %w", err)
	}

	// Delete old banner file if it exists and is different
	if oldBannerURL != "" && oldBannerURL != bannerURL {
		if err := s.storage.DeleteBanner(ctx, circleID); err != nil {
			s.logger.Warn("failed to delete old circle banner",
				zap.String("circle_id", circleID),
				zap.String("old_url", oldBannerURL),
				zap.Error(err))
		}
	}

	s.logger.Info("circle banner updated successfully",
		zap.String("circle_id", circleID),
		zap.String("new_url", bannerURL))

	return nil
}

// RemoveAvatar removes the avatar from a circle
func (s *Service) RemoveAvatar(ctx context.Context, circleID string, requesterProfileID string) error {
	s.logger.Debug("removing circle avatar",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID))

	// Get current circle to verify ownership
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if requester is the owner
	if circle.OwnerProfileID != requesterProfileID {
		return fmt.Errorf("only circle owner can remove avatar")
	}

	// Delete avatar file
	if err := s.storage.DeleteAvatar(ctx, circleID); err != nil {
		s.logger.Warn("failed to delete circle avatar file",
			zap.String("circle_id", circleID),
			zap.Error(err))
	}

	// Update circle to remove avatar URL
	circle.AvatarURL = ""
	circle.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCircle(ctx, circle); err != nil {
		s.logger.Error("failed to update circle to remove avatar",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to update circle: %w", err)
	}

	s.logger.Info("circle avatar removed successfully",
		zap.String("circle_id", circleID))
	return nil
}

// RemoveBanner removes the banner from a circle
func (s *Service) RemoveBanner(ctx context.Context, circleID string, requesterProfileID string) error {
	s.logger.Debug("removing circle banner",
		zap.String("circle_id", circleID),
		zap.String("requester_profile_id", requesterProfileID))

	// Get current circle to verify ownership
	circle, err := s.repo.GetCircleByID(ctx, circleID)
	if err != nil {
		return fmt.Errorf("failed to get circle: %w", err)
	}
	if circle == nil {
		return fmt.Errorf("circle not found")
	}

	// Check if requester is the owner
	if circle.OwnerProfileID != requesterProfileID {
		return fmt.Errorf("only circle owner can remove banner")
	}

	// Delete banner file
	if err := s.storage.DeleteBanner(ctx, circleID); err != nil {
		s.logger.Warn("failed to delete circle banner file",
			zap.String("circle_id", circleID),
			zap.Error(err))
	}

	// Update circle to remove banner URL
	circle.BannerURL = ""
	circle.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCircle(ctx, circle); err != nil {
		s.logger.Error("failed to update circle to remove banner",
			zap.String("circle_id", circleID),
			zap.Error(err))
		return fmt.Errorf("failed to update circle: %w", err)
	}

	s.logger.Info("circle banner removed successfully",
		zap.String("circle_id", circleID))
	return nil
}

// Helper functions

func isValidVisibility(visibility string) bool {
	return visibility == domain.CircleVisibilityPublic ||
		visibility == domain.CircleVisibilityPrivate ||
		visibility == domain.CircleVisibilityUnlisted
}
