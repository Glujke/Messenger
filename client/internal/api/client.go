package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the messenger backend API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

// New creates an API client for the given server base URL.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UserProfile describes the authenticated user.
type UserProfile struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// GetMe returns the current user's profile.
func (c *Client) GetMe(ctx context.Context, token string) (UserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/me", nil)
	if err != nil {
		return UserProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UserProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return UserProfile{}, fmt.Errorf("%s", errResp.Error)
		}
		return UserProfile{}, fmt.Errorf("failed to get profile: %s", resp.Status)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return UserProfile{}, err
	}
	return profile, nil
}

// UpdateUsername changes the authenticated user's username.
func (c *Client) UpdateUsername(ctx context.Context, token, username string) (UserProfile, error) {
	body, _ := json.Marshal(map[string]string{"username": username})

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/auth/me", bytes.NewBuffer(body))
	if err != nil {
		return UserProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return UserProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return UserProfile{}, fmt.Errorf("%s", errResp.Error)
		}
		return UserProfile{}, fmt.Errorf("failed to update username: %s", resp.Status)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return UserProfile{}, err
	}
	return profile, nil
}

// ChangePassword updates the authenticated user's password.
func (c *Client) ChangePassword(ctx context.Context, token, oldPassword, newPassword string) error {
	body, _ := json.Marshal(map[string]string{
		"old_password": oldPassword,
		"new_password": newPassword,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/password", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("failed to change password: %s", resp.Status)
	}
	return nil
}

// Login authenticates a user and returns a JWT token.
func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return "", fmt.Errorf("%s", errResp.Error)
		}
		return "", fmt.Errorf("login failed: %s", resp.Status)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// Register creates a new user account.
func (c *Client) Register(ctx context.Context, email, username, password string) error {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"username": username,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/register", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("registration failed: %s", resp.Status)
	}

	return nil
}

// Room represents a chat room in the API.
type Room struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name,omitempty"`
	PeerID    int64  `json:"peer_id,omitempty"`
	PeerEmail string `json:"peer_email,omitempty"`
}

// Attachment describes a file attached to a message.
type Attachment struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// Message represents a chat message in the API.
type Message struct {
	ID         int64       `json:"id"`
	RoomID     int64       `json:"room_id"`
	SenderID   int64       `json:"sender_id"`
	Type       string      `json:"type"`
	Body       string      `json:"body"`
	Attachment *Attachment `json:"attachment,omitempty"`
	CreatedAt  string      `json:"created_at"`
}

// ContactRequest represents a contact invitation (incoming or outgoing).
type ContactRequest struct {
	ID           int64  `json:"id"`
	FromUserID   int64  `json:"from_user_id"`
	ToUserID     int64  `json:"to_user_id"`
	Status       string `json:"status"`
	PeerEmail    string `json:"peer_email,omitempty"`
	PeerUsername string `json:"peer_username,omitempty"`
	CreatedAt    string `json:"created_at"`
	RespondedAt  string `json:"responded_at,omitempty"`
}

// Contact represents a confirmed friend.
type Contact struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// GetRooms returns the list of rooms for the authenticated user.
func (c *Client) GetRooms(ctx context.Context, token string) ([]Room, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/rooms", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("failed to get rooms: %s", resp.Status)
	}

	var result struct {
		Rooms []Room `json:"rooms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Rooms, nil
}

// DefaultMessageLimit is the page size for message history requests.
const DefaultMessageLimit = 50

// GetMessages returns message history for a room.
// Pass limit=0 to use the server default. Pass beforeID>0 to load older messages.
func (c *Client) GetMessages(ctx context.Context, token string, roomID int64, limit int, beforeID int64) ([]Message, error) {
	url := fmt.Sprintf("%s/rooms/%d/messages", c.baseURL, roomID)
	if limit > 0 || beforeID > 0 {
		q := make([]string, 0, 2)
		if limit > 0 {
			q = append(q, fmt.Sprintf("limit=%d", limit))
		}
		if beforeID > 0 {
			q = append(q, fmt.Sprintf("before_id=%d", beforeID))
		}
		url += "?" + strings.Join(q, "&")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("failed to get messages: %s", resp.Status)
	}

	var result struct {
		Messages []Message `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Messages, nil
}

// SendMessage sends a new text message to a room.
func (c *Client) SendMessage(ctx context.Context, token string, roomID int64, body string) (Message, error) {
	url := fmt.Sprintf("%s/rooms/%d/messages", c.baseURL, roomID)
	reqBody, _ := json.Marshal(map[string]string{
		"body": body,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return Message{}, fmt.Errorf("%s", errResp.Error)
		}
		return Message{}, fmt.Errorf("failed to send message: %s", resp.Status)
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

// SendAttachmentMessage links an uploaded attachment to a new room message.
func (c *Client) SendAttachmentMessage(ctx context.Context, token string, roomID, attachmentID int64, body string) (Message, error) {
	url := fmt.Sprintf("%s/rooms/%d/messages", c.baseURL, roomID)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"attachment_id": attachmentID,
		"body":          body,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return Message{}, fmt.Errorf("%s", errResp.Error)
		}
		return Message{}, fmt.Errorf("failed to send attachment message: %s", resp.Status)
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

// InviteContact sends a contact invitation to another user.
func (c *Client) InviteContact(ctx context.Context, token, emailOrUsername string) error {
	body, _ := json.Marshal(map[string]string{
		"identifier": emailOrUsername,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/contacts/invite", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("failed to invite contact: %s", resp.Status)
	}

	return nil
}

// GetContactRequests returns pending contact invitations.
func (c *Client) GetContactRequests(ctx context.Context, token string) ([]ContactRequest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/contacts/requests", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("failed to get contact requests: %s", resp.Status)
	}

	var result struct {
		Requests []ContactRequest `json:"requests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Requests, nil
}

// AcceptContact accepts a contact invitation.
func (c *Client) AcceptContact(ctx context.Context, token string, requestID int64) error {
	url := fmt.Sprintf("%s/contacts/requests/%d/accept", c.baseURL, requestID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("failed to accept contact: %s", resp.Status)
	}

	return nil
}

// ListContacts returns the list of confirmed contacts.
func (c *Client) ListContacts(ctx context.Context, token string) ([]Contact, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/contacts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("failed to list contacts: %s", resp.Status)
	}

	var result struct {
		Contacts []Contact `json:"contacts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Contacts, nil
}

// CreateDirectRoom opens or creates a 1:1 chat with another user.
func (c *Client) CreateDirectRoom(ctx context.Context, token string, userID int64) (int64, error) {
	body, _ := json.Marshal(map[string]int64{"user_id": userID})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/rooms/direct", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return 0, fmt.Errorf("%s", errResp.Error)
		}
		return 0, fmt.Errorf("failed to create direct room: %s", resp.Status)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// CreateGroup creates a new group room.
func (c *Client) CreateGroup(ctx context.Context, token, name string, userIDs []int64) (int64, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"name":     name,
		"user_ids": userIDs,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/rooms/group", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return 0, fmt.Errorf("%s", errResp.Error)
		}
		return 0, fmt.Errorf("failed to create group: %s", resp.Status)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.ID, nil
}

// UploadAttachment uploads a file to a room and returns the attachment ID.
func (c *Client) UploadAttachment(ctx context.Context, token string, roomID int64, filename string, r io.Reader) (int64, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return 0, err
	}
	if _, err := io.Copy(part, r); err != nil {
		return 0, err
	}
	writer.Close()

	url := fmt.Sprintf("%s/rooms/%d/attachments", c.baseURL, roomID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return 0, fmt.Errorf("%s", errResp.Error)
		}
		return 0, fmt.Errorf("failed to upload attachment: %s", resp.Status)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.ID, nil
}

// DownloadAttachment returns a reader for the attachment data.
func (c *Client) DownloadAttachment(ctx context.Context, token string, attachmentID int64) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/attachments/%d", c.baseURL, attachmentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("failed to download attachment: %s", resp.Status)
	}

	return resp.Body, nil
}

// Health checks whether the backend is reachable.
func (c *Client) Health(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("health check failed: %s", resp.Status)
	}

	return string(body), nil
}
