package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
	"messenger/backend/internal/service"
)

// ContactManager handles contact-related operations.
type ContactManager interface {
	Invite(ctx context.Context, fromID int64, identifier string) (service.ContactRequestItem, error)
	AcceptRequest(ctx context.Context, userID, requestID int64) error
	RejectRequest(ctx context.Context, userID, requestID int64) error
	ListRequests(ctx context.Context, userID int64) ([]service.ContactRequestItem, error)
	ListContacts(ctx context.Context, userID int64) ([]service.ContactItem, error)
}

// ContactsHandler serves contact endpoints.
type ContactsHandler struct {
	manager ContactManager
}

// NewContactsHandler creates a contacts HTTP handler.
func NewContactsHandler(manager ContactManager) *ContactsHandler {
	return &ContactsHandler{
		manager: manager,
	}
}

type inviteRequest struct {
	Identifier string `json:"identifier"` // email or username
}

type contactRequestsResponse struct {
	Requests []service.ContactRequestItem `json:"requests"`
}

type contactsResponse struct {
	Contacts []service.ContactItem `json:"contacts"`
}

// ServeHTTP implements http.Handler.
func (h *ContactsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/contacts/invite":
		h.serveInvite(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/contacts/requests":
		h.serveListRequests(w, r)
	case r.Method == http.MethodPost && (len(r.URL.Path) > 19 && r.URL.Path[:19] == "/contacts/requests/"):
		// Simple routing for /contacts/requests/{id}/accept|reject
		h.serveRequestAction(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/contacts":
		h.serveListContacts(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *ContactsHandler) serveInvite(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	item, err := h.manager.Invite(r.Context(), caller.ID, req.Identifier)
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		return
	}
	if errors.Is(err, domain.ErrAlreadyFriends) || errors.Is(err, domain.ErrRequestPending) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *ContactsHandler) serveListRequests(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	items, err := h.manager.ListRequests(r.Context(), caller.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, contactRequestsResponse{Requests: items})
}

func (h *ContactsHandler) serveListContacts(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	items, err := h.manager.ListContacts(r.Context(), caller.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, contactsResponse{Contacts: items})
}

func (h *ContactsHandler) serveRequestAction(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	// Manual parsing of /contacts/requests/{id}/{action}
	// Path is like "/contacts/requests/123/accept"
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request id"})
		return
	}

	action := parts[4]
	var actionErr error
	switch action {
	case "accept":
		actionErr = h.manager.AcceptRequest(r.Context(), caller.ID, id)
	case "reject":
		actionErr = h.manager.RejectRequest(r.Context(), caller.ID, id)
	default:
		http.NotFound(w, r)
		return
	}

	if errors.Is(actionErr, domain.ErrRequestNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "request not found"})
		return
	}
	if actionErr != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
