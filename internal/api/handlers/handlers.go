package handlers

import (
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

	// Create handlers
	return &Handler{
		Auth: &AuthHandler{
			authService: authService,
			userService: userService,
			config:      cfg,
			logger:      logger,
		},
		Message: NewMessageHandler(messageService, userService, logger),
		Contact: NewContactHandler(contactService, logger),
		Key:     NewKeyHandler(userService, logger),
		Account: NewAccountHandler(accountService, authService, logger),
		logger:  logger,
	}
}
