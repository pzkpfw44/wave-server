package handlers

import (
	"context"

	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/service"
)

// Handler is a container for all handlers
type Handler struct {
	Auth    *AuthHandler
	Message *MessageHandler
	Contact *ContactHandler
	Key     *KeyHandler
	Account *AccountHandler
	logger  *zap.Logger
}

// NewHandler creates a new Handler with all handlers
func NewHandler(db *repository.Database, cfg *config.Config, logger *zap.Logger) *Handler {
	// Create repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	contactRepo := repository.NewContactRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Create services
	userService := service.NewUserService(userRepo, logger)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg, logger)
	messageService := service.NewMessageService(messageRepo, userRepo, logger)
	contactService := service.NewContactService(contactRepo, logger)
	accountService := service.NewAccountService(userRepo, contactRepo, messageRepo, tokenRepo, logger)

	// Schedule token cleanup with a proper background context
	ctx := context.Background()
	authService.ScheduleTokenCleanup(ctx)

	// Create handlers
	authHandler := NewAuthHandler(authService, userService, cfg, logger)
	messageHandler := NewMessageHandler(messageService, userService, logger)
	contactHandler := NewContactHandler(contactService, logger)
	keyHandler := NewKeyHandler(userService, logger)
	accountHandler := NewAccountHandler(accountService, authService, logger)

	return &Handler{
		Auth:    authHandler,
		Message: messageHandler,
		Contact: contactHandler,
		Key:     keyHandler,
		Account: accountHandler,
		logger:  logger,
	}
}
