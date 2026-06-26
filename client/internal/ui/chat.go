package ui

import (
	"context"
	"io"
	"messenger/client/internal/api"
	"messenger/client/internal/platform"
	"messenger/client/internal/state"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	maxImagePreviewHeight = float32(220)
	rowHeaderHeight       = float32(24)
	rowButtonHeight       = float32(36)
	rowPadding            = float32(12)
	dateColumnWidth       = float32(128)
)

// ChatArea is the right panel with message history and input.
type ChatArea struct {
	state        *state.AppState
	list         *widget.List
	input        *chatEntry
	send         *widget.Button
	attach       *widget.Button
	dropBound    bool
	imageCache   map[int64]fyne.Resource
	imageLoading map[int64]bool
	imageMu      sync.Mutex
}

type chatEntry struct {
	widget.Entry
	onPaste func(fyne.Clipboard) bool
}

func newChatEntry(onPaste func(fyne.Clipboard) bool) *chatEntry {
	e := &chatEntry{onPaste: onPaste}
	e.ExtendBaseWidget(e)
	e.SetPlaceHolder("Type a message...")
	return e
}

func (e *chatEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if paste, ok := shortcut.(*fyne.ShortcutPaste); ok && e.onPaste != nil {
		if e.onPaste(paste.Clipboard) {
			return
		}
	}
	e.Entry.TypedShortcut(shortcut)
}

// NewChatArea creates a new chat area.
func NewChatArea(s *state.AppState) *ChatArea {
	ca := &ChatArea{
		state:        s,
		imageCache:   make(map[int64]fyne.Resource),
		imageLoading: make(map[int64]bool),
	}
	ca.setupList()
	ca.setupInput()

	s.OnMessagesUpdate = func() {
		ca.refreshMessages()
	}

	return ca
}

func (ca *ChatArea) refreshMessages() {
	ca.refreshListHeights()
	ca.list.Refresh()
	if len(ca.state.Messages) > 0 {
		ca.list.ScrollToBottom()
	}
}

func (ca *ChatArea) setupList() {
	ca.list = widget.NewList(
		func() int {
			return len(ca.state.Messages)
		},
		func() fyne.CanvasObject {
			return newMessageListItem()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			msg := ca.state.Messages[id]
			item := obj.(*messageListItem)
			date := item.textRow.date
			body := item.textRow.body
			preview := item.preview
			downloadBtn := item.download

			date.SetText(formatMessageTime(msg.CreatedAt))
			body.Wrapping = fyne.TextWrapWord
			preview.hidePreview()

			if msg.Attachment != nil && isImageAttachment(msg.Attachment) {
				if msg.Body == "" {
					body.Hide()
					body.SetText("")
				} else {
					body.Show()
					body.SetText(msg.Body)
				}

				attachmentID := msg.Attachment.ID
				filename := msg.Attachment.Filename
				preview.showPreview(
					fyne.NewSize(240, maxImagePreviewHeight),
					func() { ca.showImageFullscreen(attachmentID, filename) },
				)
				ca.applyImagePreview(preview, attachmentID, filename)

				downloadBtn.SetText("Download " + filename)
				downloadBtn.Show()
				downloadBtn.OnTapped = func() {
					ca.downloadAttachment(attachmentID, filename)
				}
				return
			}

			body.Show()
			body.SetText(formatMessageBody(msg))

			if msg.Attachment != nil {
				downloadBtn.SetText("Download " + msg.Attachment.Filename)
				downloadBtn.Show()
				attachmentID := msg.Attachment.ID
				filename := msg.Attachment.Filename
				downloadBtn.OnTapped = func() {
					ca.downloadAttachment(attachmentID, filename)
				}
				return
			}

			downloadBtn.Hide()
			downloadBtn.OnTapped = nil
		},
	)
}

func isImageAttachment(att *api.Attachment) bool {
	if strings.HasPrefix(att.ContentType, "image/") {
		return true
	}
	switch strings.ToLower(filepath.Ext(att.Filename)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func (ca *ChatArea) refreshListHeights() {
	for i, msg := range ca.state.Messages {
		ca.list.SetItemHeight(i, ca.computeRowHeight(msg))
	}
}

func (ca *ChatArea) computeRowHeight(msg api.Message) float32 {
	width := ca.list.Size().Width - 32
	if width < 120 {
		width = 480
	}

	textRowH := ca.textRowHeight(msg, width)

	switch {
	case msg.Attachment != nil && isImageAttachment(msg.Attachment):
		return textRowH + maxImagePreviewHeight + rowButtonHeight + rowPadding
	case msg.Attachment != nil:
		return textRowH + rowButtonHeight + rowPadding
	default:
		return textRowH + rowPadding
	}
}

func (ca *ChatArea) textRowHeight(msg api.Message, listWidth float32) float32 {
	textWidth := listWidth - dateColumnWidth - 8
	if textWidth < 80 {
		textWidth = 80
	}

	var text string
	switch {
	case msg.Attachment != nil && isImageAttachment(msg.Attachment):
		text = msg.Body
	default:
		text = formatMessageBody(msg)
	}

	bodyH := estimateWrappedTextHeight(text, textWidth)
	if bodyH < rowHeaderHeight {
		return rowHeaderHeight
	}
	return bodyH
}

func estimateWrappedTextHeight(text string, width float32) float32 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	lineHeight := theme.TextSize() * 1.35
	charWidth := theme.TextSize() * 0.55
	if charWidth <= 0 {
		charWidth = 8
	}

	lines := 0
	lineLen := float32(0)
	for _, r := range text {
		if r == '\n' {
			lines++
			lineLen = 0
			continue
		}
		lineLen += charWidth
		if lineLen >= width {
			lines++
			lineLen = charWidth
		}
	}
	lines++

	return float32(lines) * lineHeight
}

func (ca *ChatArea) applyImagePreview(preview *tappablePreview, attachmentID int64, filename string) {
	if res := ca.cachedImage(attachmentID); res != nil {
		preview.setResource(res)
		return
	}
	ca.ensureImageLoaded(attachmentID, filename)
}

func (ca *ChatArea) showImageFullscreen(attachmentID int64, filename string) {
	res := ca.cachedImage(attachmentID)
	if res == nil {
		ca.ensureImageLoaded(attachmentID, filename)
		return
	}

	fyne.Do(func() {
		img := canvas.NewImageFromResource(res)
		img.FillMode = canvas.ImageFillContain
		scroll := container.NewScroll(img)
		scroll.SetMinSize(fyne.NewSize(700, 500))
		d := dialog.NewCustom(filename, "Закрыть", scroll, ca.state.Window)
		d.Resize(fyne.NewSize(900, 700))
		d.Show()
	})
}

func (ca *ChatArea) showError(err error) {
	fyne.Do(func() {
		dialog.ShowError(err, ca.state.Window)
	})
}

func (ca *ChatArea) showInfo(title, message string) {
	fyne.Do(func() {
		dialog.ShowInformation(title, message, ca.state.Window)
	})
}

func (ca *ChatArea) cachedImage(attachmentID int64) fyne.Resource {
	ca.imageMu.Lock()
	defer ca.imageMu.Unlock()
	return ca.imageCache[attachmentID]
}

func (ca *ChatArea) ensureImageLoaded(attachmentID int64, filename string) {
	ca.imageMu.Lock()
	if ca.imageCache[attachmentID] != nil || ca.imageLoading[attachmentID] {
		ca.imageMu.Unlock()
		return
	}
	ca.imageLoading[attachmentID] = true
	ca.imageMu.Unlock()

	go func() {
		defer func() {
			ca.imageMu.Lock()
			delete(ca.imageLoading, attachmentID)
			ca.imageMu.Unlock()
		}()

		reader, err := ca.state.API.DownloadAttachment(context.Background(), ca.state.Token, attachmentID)
		if err != nil {
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return
		}

		res := fyne.NewStaticResource(filename, data)
		ca.imageMu.Lock()
		ca.imageCache[attachmentID] = res
		ca.imageMu.Unlock()

		fyne.Do(func() {
			ca.list.Refresh()
		})
	}()
}

func formatMessageBody(msg api.Message) string {
	if msg.Attachment != nil {
		prefix := "📎 " + msg.Attachment.Filename
		if msg.Body != "" {
			return prefix + "\n" + msg.Body
		}
		return prefix
	}
	return msg.Body
}

func (ca *ChatArea) setupInput() {
	ca.input = newChatEntry(ca.handleClipboardPaste)
	ca.input.OnSubmitted = func(string) {
		ca.sendMessage()
	}

	ca.send = widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		ca.sendMessage()
	})

	ca.attach = widget.NewButtonWithIcon("", theme.MailAttachmentIcon(), func() {
		ca.pickFile()
	})
}

func (ca *ChatArea) bindDropHandler() {
	if ca.dropBound {
		return
	}
	ca.dropBound = true
	ca.state.Window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			ca.uploadFromURI(uri)
		}
	})
}

func (ca *ChatArea) pickFile() {
	if ca.state.ActiveRoomID == 0 {
		return
	}
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		name := filepath.Base(reader.URI().Path())
		ca.uploadFromReader(name, reader)
	}, ca.state.Window)
}

func (ca *ChatArea) handleClipboardPaste(clipboard fyne.Clipboard) bool {
	if ca.state.ActiveRoomID == 0 {
		return false
	}

	if paths, err := platform.ClipboardFiles(); err == nil && len(paths) > 0 {
		for _, path := range paths {
			ca.uploadFromPath(path)
		}
		return true
	}

	if path, ok := platform.ClipboardTextPath(clipboard.Content()); ok {
		ca.uploadFromPath(path)
		return true
	}

	if path, ok, err := platform.ClipboardImageFile(); err != nil {
		dialog.ShowError(err, ca.state.Window)
		return true
	} else if ok {
		ca.uploadFromPath(path)
		return true
	}

	return false
}

func (ca *ChatArea) uploadFromURI(uri fyne.URI) {
	reader, err := storage.Reader(uri)
	if err != nil {
		dialog.ShowError(err, ca.state.Window)
		return
	}
	name := filepath.Base(uri.Path())
	ca.uploadFromReader(name, reader)
}

func (ca *ChatArea) uploadFromPath(path string) {
	file, err := os.Open(path)
	if err != nil {
		dialog.ShowError(err, ca.state.Window)
		return
	}
	ca.uploadFromReader(filepath.Base(path), file)
}

func (ca *ChatArea) uploadFromReader(filename string, reader io.ReadCloser) {
	if ca.state.ActiveRoomID == 0 {
		reader.Close()
		return
	}

	ca.setInputEnabled(false)
	caption := ca.input.Text

	go func() {
		defer reader.Close()
		defer ca.setInputEnabled(true)

		attachmentID, err := ca.state.API.UploadAttachment(
			context.Background(),
			ca.state.Token,
			ca.state.ActiveRoomID,
			filename,
			reader,
		)
		if err != nil {
			ca.showError(err)
			return
		}

		encryptedCaption := ""
		if strings.TrimSpace(caption) != "" {
			encryptedCaption, err = ca.state.Encrypter.Encrypt(strings.TrimSpace(caption))
			if err != nil {
				ca.showError(err)
				return
			}
		}

		msg, err := ca.state.API.SendAttachmentMessage(
			context.Background(),
			ca.state.Token,
			ca.state.ActiveRoomID,
			attachmentID,
			encryptedCaption,
		)
		if err != nil {
			ca.showError(err)
			return
		}

		fyne.Do(func() {
			ca.input.SetText("")
			ca.state.AddMessage(msg)
		})
	}()
}

func (ca *ChatArea) downloadAttachment(attachmentID int64, filename string) {
	go func() {
		dir, err := platform.DownloadsDir(ca.state.AppName)
		if err != nil {
			ca.showError(err)
			return
		}

		target := platform.UniqueFilePath(dir, filename)

		reader, err := ca.state.API.DownloadAttachment(context.Background(), ca.state.Token, attachmentID)
		if err != nil {
			ca.showError(err)
			return
		}
		defer reader.Close()

		file, err := os.Create(target)
		if err != nil {
			ca.showError(err)
			return
		}

		if _, err := io.Copy(file, reader); err != nil {
			file.Close()
			ca.showError(err)
			return
		}
		if err := file.Sync(); err != nil {
			file.Close()
			ca.showError(err)
			return
		}
		if err := file.Close(); err != nil {
			ca.showError(err)
			return
		}

		ca.showInfo("Файл сохранён", target)
	}()
}

func (ca *ChatArea) setInputEnabled(enabled bool) {
	fyne.Do(func() {
		if enabled {
			ca.input.Enable()
			ca.send.Enable()
			ca.attach.Enable()
			ca.state.Window.Canvas().Focus(ca.input)
			return
		}
		ca.input.Disable()
		ca.send.Disable()
		ca.attach.Disable()
	})
}

func (ca *ChatArea) sendMessage() {
	text := ca.input.Text
	if text == "" || ca.state.ActiveRoomID == 0 {
		return
	}

	ca.input.SetText("")
	ca.setInputEnabled(false)

	go func() {
		defer ca.setInputEnabled(true)

		encrypted, err := ca.state.Encrypter.Encrypt(text)
		if err != nil {
			ca.showError(err)
			return
		}

		msg, err := ca.state.API.SendMessage(context.Background(), ca.state.Token, ca.state.ActiveRoomID, encrypted)
		if err != nil {
			ca.showError(err)
			return
		}

		fyne.Do(func() {
			ca.state.AddMessage(msg)
		})
	}()
}

// Content returns the chat area layout.
func (ca *ChatArea) Content() fyne.CanvasObject {
	if ca.state.ActiveRoomID == 0 {
		return container.NewCenter(widget.NewLabel("Select a chat to start messaging"))
	}

	ca.bindDropHandler()
	ca.refreshListHeights()

	inputRow := container.NewBorder(nil, nil, ca.attach, ca.send, ca.input)
	dropHint := canvas.NewText("Drop files here to send", theme.Color(theme.ColorNamePlaceHolder))
	dropHint.TextSize = theme.TextSize()
	dropHint.Alignment = fyne.TextAlignCenter

	return container.NewBorder(
		container.NewPadded(dropHint),
		inputRow,
		nil,
		nil,
		ca.list,
	)
}
