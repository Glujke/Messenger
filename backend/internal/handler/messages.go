package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/service"
)

// MessageSender stores a text message.
type MessageSender interface {
	SendText(ctx context.Context, roomID, senderID int64, body string) (service.MessageItem, error)
}

// MessageLister returns room messages.
type MessageLister interface {
	List(ctx context.Context, roomID, userID int64, limit int, beforeID int64) ([]service.MessageItem, error)
}

// MessagesHandler serves room message endpoints.
type MessagesHandler struct {
	sender MessageSender
	lister MessageLister
}

// NewMessagesHandler creates a messages HTTP handler.
func NewMessagesHandler(sender MessageSender, lister MessageLister) *MessagesHandler {
	return &MessagesHandler{
		sender: sender,
		lister: lister,
	}
}

type sendMessageRequest struct {
	Body string `json:"body"`
}

type messageListResponse struct {
	Messages []service.MessageItem `json:"messages"`
}

// ServeHTTP implements http.Handler.
func (h *MessagesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID, err := parseRoomID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid room id"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.serveSend(w, r, roomID)
	case http.MethodGet:
		h.serveList(w, r, roomID)
	default:
		http.NotFound(w, r)
	}
}

func (h *MessagesHandler) serveSend(w http.ResponseWriter, r *http.Request, roomID int64) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	item, err := h.sender.SendText(r.Context(), roomID, caller.ID, req.Body)
	if errors.Is(err, domain.ErrEmptyMessageBody) || errors.Is(err, domain.ErrMessageTooLong) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, service.ErrNotRoomMember) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "not a room member"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *MessagesHandler) serveList(w http.ResponseWriter, r *http.Request, roomID int64) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	limit, beforeID, err := parseMessageQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	messages, err := h.lister.List(r.Context(), roomID, caller.ID, limit, beforeID)
	if errors.Is(err, service.ErrNotRoomMember) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "not a room member"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	if messages == nil {
		messages = []service.MessageItem{}
	}

	writeJSON(w, http.StatusOK, messageListResponse{Messages: messages})
}

func parseRoomID(value string) (int64, error) {
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	roomID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || roomID <= 0 {
		return 0, strconv.ErrSyntax
	}
	return roomID, nil
}

func parseMessageQuery(r *http.Request) (int, int64, error) {
	limit := 0
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("invalid limit")
		}
		limit = parsed
	}

	var beforeID int64
	if value := r.URL.Query().Get("before_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			return 0, 0, errors.New("invalid before_id")
		}
		beforeID = parsed
	}

	return limit, beforeID, nil
}
