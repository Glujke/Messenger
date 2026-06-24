package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/service"
)

// AttachmentUploader stores uploaded files for a room.
type AttachmentUploader interface {
	Upload(ctx context.Context, roomID, uploaderID int64, filename, contentType string, reader io.Reader, declaredSize int64) (service.UploadResult, error)
}

// AttachmentDownloader opens uploaded files.
type AttachmentDownloader interface {
	Open(ctx context.Context, attachmentID, userID int64) (service.OpenResult, error)
}

// AttachmentsHandler serves attachment upload and download endpoints.
type AttachmentsHandler struct {
	uploader AttachmentUploader
	download AttachmentDownloader
}

// NewAttachmentsHandler creates an attachments HTTP handler.
func NewAttachmentsHandler(uploader AttachmentUploader, download AttachmentDownloader) *AttachmentsHandler {
	return &AttachmentsHandler{
		uploader: uploader,
		download: download,
	}
}

// ServeUpload handles POST /rooms/{id}/attachments.
func (h *AttachmentsHandler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	roomID, err := parseRoomID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid room id"})
		return
	}

	if err := r.ParseMultipartForm(domain.MaxUploadBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	result, err := h.uploader.Upload(r.Context(), roomID, caller.ID, header.Filename, contentType, file, header.Size)
	if errors.Is(err, domain.ErrUploadTooLarge) || errors.Is(err, domain.ErrUploadEmpty) {
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

	writeJSON(w, http.StatusCreated, result)
}

// ServeDownload handles GET /attachments/{id}.
func (h *AttachmentsHandler) ServeDownload(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	attachmentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || attachmentID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid attachment id"})
		return
	}

	result, err := h.download.Open(r.Context(), attachmentID, caller.ID)
	if errors.Is(err, service.ErrAttachmentNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
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
	defer result.Reader.Close()

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+result.Filename+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(result.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, result.Reader)
}
