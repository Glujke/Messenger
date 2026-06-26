package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"messenger/backend/internal/config"
	"messenger/backend/internal/domain"
	"messenger/backend/internal/handler"
	"messenger/backend/internal/platform/db"
	"messenger/backend/internal/platform/storage"
	"messenger/backend/internal/repository/postgres"
	"messenger/backend/internal/service"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	fileStore, err := storage.NewDiskStore(cfg.UploadDir)
	if err != nil {
		logger.Error("upload storage initialization failed", "error", err)
		os.Exit(1)
	}

	store := postgres.NewStore(database)
	hub := service.NewHub()
	authService := service.NewAuthService(store, cfg.JWTSecret, cfg.JWTTTL)
	authHandler := handler.NewAuthHandler(authService, authService)
	profileService := service.NewProfileService(store)
	meHandler := handler.NewMeHandler(profileService)
	passwordHandler := handler.NewPasswordHandler(profileService)
	roomService := service.NewRoomService(store, store, store)
	roomsHandler := handler.NewRoomsHandler(roomService, roomService, roomService)
	attachmentService := service.NewAttachmentService(store, store, fileStore, cfg.MaxUploadBytes)
	attachmentsHandler := handler.NewAttachmentsHandler(attachmentService, attachmentService)
	contactService := service.NewContactService(store, store)
	contactsHandler := handler.NewContactsHandler(contactService)
	messageService := service.NewMessageService(store, store, store, hub, domain.NewXOREncrypter(cfg.EncryptionKey))
	messagesHandler := handler.NewMessagesHandler(messageService, messageService)
	wsHandler := handler.NewWSHandler(hub, cfg.JWTSecret, store, logger)
	router := handler.NewRouter(logger, store, authHandler, meHandler, passwordHandler, contactsHandler, roomsHandler, messagesHandler, attachmentsHandler, wsHandler, cfg.JWTSecret)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api stopped")
}
