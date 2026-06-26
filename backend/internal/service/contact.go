package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

// ContactService handles contact requests and confirmed connections.
type ContactService struct {
	users    repository.UserStore
	contacts repository.ContactStore
}

// NewContactService creates a new contact service.
func NewContactService(users repository.UserStore, contacts repository.ContactStore) *ContactService {
	return &ContactService{
		users:    users,
		contacts: contacts,
	}
}

// ContactRequestItem is a DTO for contact requests.
type ContactRequestItem struct {
	ID           int64                       `json:"id"`
	FromUserID   int64                       `json:"from_user_id"`
	ToUserID     int64                       `json:"to_user_id"`
	Status       domain.ContactRequestStatus `json:"status"`
	PeerEmail    string                      `json:"peer_email,omitempty"`
	PeerUsername string                      `json:"peer_username,omitempty"`
	CreatedAt    time.Time                   `json:"created_at"`
	RespondedAt  *time.Time                  `json:"responded_at,omitempty"`
}

// ContactItem is a DTO for confirmed contacts.
type ContactItem struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Invite sends a contact request to another user by email or username.
func (s *ContactService) Invite(ctx context.Context, fromID int64, identifier string) (ContactRequestItem, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if identifier == "" {
		return ContactRequestItem{}, errors.New("identifier is required")
	}

	var (
		peer repository.UserRecord
		err  error
	)

	if strings.Contains(identifier, "@") {
		peer, err = s.users.FindByEmail(ctx, identifier)
	} else {
		peer, err = s.users.FindByUsername(ctx, identifier)
	}

	if errors.Is(err, repository.ErrNotFound) {
		return ContactRequestItem{}, repository.ErrNotFound
	}
	if err != nil {
		return ContactRequestItem{}, err
	}

	if peer.ID == fromID {
		return ContactRequestItem{}, errors.New("cannot invite yourself")
	}

	// Check if already friends
	areFriends, err := s.contacts.AreContacts(ctx, fromID, peer.ID)
	if err != nil {
		return ContactRequestItem{}, err
	}
	if areFriends {
		return ContactRequestItem{}, domain.ErrAlreadyFriends
	}

	record, err := s.contacts.CreateRequest(ctx, fromID, peer.ID)
	if err != nil {
		return ContactRequestItem{}, err
	}

	return ContactRequestItem{
		ID:          record.ID,
		FromUserID:  record.FromUserID,
		ToUserID:    record.ToUserID,
		Status:      record.Status,
		CreatedAt:   record.CreatedAt,
		RespondedAt: record.RespondedAt,
	}, nil
}

// AcceptRequest approves a contact invitation.
func (s *ContactService) AcceptRequest(ctx context.Context, userID, requestID int64) error {
	req, err := s.contacts.GetRequest(ctx, requestID)
	if err != nil {
		return err
	}

	if req.ToUserID != userID {
		return errors.New("unauthorized to accept this request")
	}
	if req.Status != domain.ContactRequestPending {
		return errors.New("request is already processed")
	}

	if err := s.contacts.UpdateRequestStatus(ctx, requestID, domain.ContactRequestAccepted); err != nil {
		return err
	}

	return s.contacts.AddContact(ctx, req.FromUserID, req.ToUserID)
}

// RejectRequest denies a contact invitation.
func (s *ContactService) RejectRequest(ctx context.Context, userID, requestID int64) error {
	req, err := s.contacts.GetRequest(ctx, requestID)
	if err != nil {
		return err
	}

	if req.ToUserID != userID {
		return errors.New("unauthorized to reject this request")
	}
	if req.Status != domain.ContactRequestPending {
		return errors.New("request is already processed")
	}

	return s.contacts.UpdateRequestStatus(ctx, requestID, domain.ContactRequestRejected)
}

// ListRequests returns all requests involving the user.
func (s *ContactService) ListRequests(ctx context.Context, userID int64) ([]ContactRequestItem, error) {
	records, err := s.contacts.ListRequests(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]ContactRequestItem, 0, len(records))
	for _, r := range records {
		item := ContactRequestItem{
			ID:          r.ID,
			FromUserID:  r.FromUserID,
			ToUserID:    r.ToUserID,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
			RespondedAt: r.RespondedAt,
		}
		if peer, err := s.peerForRequest(ctx, userID, r); err == nil {
			item.PeerEmail = peer.Email
			item.PeerUsername = peer.Username
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ContactService) peerForRequest(ctx context.Context, userID int64, r repository.ContactRequestRecord) (repository.UserRecord, error) {
	peerID := r.FromUserID
	if peerID == userID {
		peerID = r.ToUserID
	}
	return s.users.FindByID(ctx, peerID)
}

// ListContacts returns the user's confirmed friends.
func (s *ContactService) ListContacts(ctx context.Context, userID int64) ([]ContactItem, error) {
	records, err := s.contacts.ListContacts(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]ContactItem, 0, len(records))
	for _, r := range records {
		peer, err := s.users.FindByID(ctx, r.ContactID)
		if errors.Is(err, repository.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		items = append(items, ContactItem{
			ID:       peer.ID,
			Email:    peer.Email,
			Username: peer.Username,
		})
	}
	return items, nil
}
