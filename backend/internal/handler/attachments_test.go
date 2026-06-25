package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/service"
)

type mockAttachmentUploader struct {
	result service.UploadResult
	err    error
}

func (m *mockAttachmentUploader) Upload(context.Context, int64, int64, string, string, io.Reader, int64) (service.UploadResult, error) {
	return m.result, m.err
}

type mockAttachmentDownloader struct {
	result service.OpenResult
	err    error
}

func (m *mockAttachmentDownloader) Open(context.Context, int64, int64) (service.OpenResult, error) {
	return m.result, m.err
}

func TestAttachmentsHandler_ServeUpload(t *testing.T) {
	body, contentType := multipartBody(t, "file", "photo.png", "image/png", []byte("data"))
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{result: service.UploadResult{ID: 5, Filename: "photo.png", MessageType: "image"}},
		&mockAttachmentDownloader{},
	)
	rec := httptest.NewRecorder()

	h.ServeUpload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestAttachmentsHandler_ServeUpload_NotMember(t *testing.T) {
	body, contentType := multipartBody(t, "file", "photo.png", "image/png", []byte("data"))
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{err: service.ErrNotRoomMember},
		&mockAttachmentDownloader{},
	)
	rec := httptest.NewRecorder()

	h.ServeUpload(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAttachmentsHandler_ServeDownload(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/attachments/5", nil)
	req.SetPathValue("id", "5")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{},
		&mockAttachmentDownloader{
			result: service.OpenResult{
				Filename:    "photo.png",
				ContentType: "image/png",
				SizeBytes:   4,
				Reader:      io.NopCloser(bytes.NewBufferString("data")),
			},
		},
	)
	rec := httptest.NewRecorder()

	h.ServeDownload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestAttachmentsHandler_ServeDownload_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/attachments/5", nil)
	req.SetPathValue("id", "5")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{},
		&mockAttachmentDownloader{err: service.ErrAttachmentNotFound},
	)
	rec := httptest.NewRecorder()

	h.ServeDownload(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func multipartBody(t *testing.T, fieldName, filename, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_ = contentType
	return &body, writer.FormDataContentType()
}

func TestAttachmentsHandler_ServeUpload_MissingFile(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/attachments", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=bad")
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(&mockAttachmentUploader{}, &mockAttachmentDownloader{})
	rec := httptest.NewRecorder()

	h.ServeUpload(rec, req)

	if rec.Code == http.StatusCreated {
		t.Fatalf("status = %d, want error", rec.Code)
	}
}

func TestAttachmentsHandler_ServeUpload_TooLarge(t *testing.T) {
	body, contentType := multipartBody(t, "file", "big.bin", "application/octet-stream", []byte("x"))
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{err: domain.ErrUploadTooLarge},
		&mockAttachmentDownloader{},
	)
	rec := httptest.NewRecorder()

	h.ServeUpload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAttachmentsHandler_ServeDownload_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/attachments/5", nil)
	req.SetPathValue("id", "5")

	h := NewAttachmentsHandler(&mockAttachmentUploader{}, &mockAttachmentDownloader{})
	rec := httptest.NewRecorder()

	h.ServeDownload(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAttachmentsHandler_ServeDownload_InternalError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/attachments/5", nil)
	req.SetPathValue("id", "5")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))

	h := NewAttachmentsHandler(
		&mockAttachmentUploader{},
		&mockAttachmentDownloader{err: errors.New("disk down")},
	)
	rec := httptest.NewRecorder()

	h.ServeDownload(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
