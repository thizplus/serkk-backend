package handlers

import (
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/redis"
	"gofiber-template/infrastructure/storage"
	chatWebsocket "gofiber-template/infrastructure/websocket"
	websocketHandler "gofiber-template/interfaces/api/websocket"
	"gofiber-template/pkg/config"
)

// Services contains all the services needed for handlers
type Services struct {
	UserService         services.UserService
	TaskService         services.TaskService
	FileService         services.FileService
	JobService          services.JobService
	PostService         services.PostService
	CommentService      services.CommentService
	VoteService         services.VoteService
	FollowService       services.FollowService
	SavedPostService    services.SavedPostService
	NotificationService services.NotificationService
	TagService          services.TagService
	SearchService       services.SearchService
	MediaService        services.MediaService
	OAuthService        services.OAuthService
	PushService         services.PushService
	ConversationService services.ConversationService
	MessageService      services.MessageService
	BlockService        services.BlockService
	FileUploadService   services.FileUploadService
}

// Handlers contains all HTTP handlers
type Handlers struct {
	UserHandler         *UserHandler
	ProfileHandler      *ProfileHandler
	TaskHandler         *TaskHandler
	FileHandler         *FileHandler
	JobHandler          *JobHandler
	PostHandler         *PostHandler
	CommentHandler      *CommentHandler
	VoteHandler         *VoteHandler
	FollowHandler       *FollowHandler
	SavedPostHandler    *SavedPostHandler
	NotificationHandler *NotificationHandler
	TagHandler          *TagHandler
	SearchHandler       *SearchHandler
	MediaHandler        *MediaHandler
	OAuthHandler        *OAuthHandler
	SEOHandler          *SEOHandler
	PushHandler         *PushHandler
	ConversationHandler *ConversationHandler
	MessageHandler        *MessageHandler
	BlockHandler          *BlockHandler
	ChatWSHandler         *websocketHandler.ChatWebSocketHandler
	NotificationWSHandler *websocketHandler.NotificationWebSocketHandler
	FileUploadHandler     *FileUploadHandler
	PresignedUploadHandler *PresignedUploadHandler
	WebhookHandler        *WebhookHandler
}

// NewHandlers creates a new instance of Handlers with all dependencies
func NewHandlers(services *Services, cfg *config.Config, chatWSHandler *websocketHandler.ChatWebSocketHandler, notificationWSHandler *websocketHandler.NotificationWebSocketHandler, chatHub *chatWebsocket.ChatHub, notificationHub *chatWebsocket.NotificationHub, conversationRepo repositories.ConversationRepository, mediaUploadService *storage.MediaUploadService, r2Storage storage.R2Storage, mediaRepo repositories.MediaRepository, redisService interface{}) *Handlers {
	return &Handlers{
		UserHandler:           NewUserHandler(services.UserService),
		ProfileHandler:        NewProfileHandler(services.UserService),
		TaskHandler:           NewTaskHandler(services.TaskService),
		FileHandler:           NewFileHandler(services.FileService),
		JobHandler:            NewJobHandler(services.JobService),
		PostHandler:           NewPostHandler(services.PostService),
		CommentHandler:        NewCommentHandler(services.CommentService),
		VoteHandler:           NewVoteHandler(services.VoteService),
		FollowHandler:         NewFollowHandler(services.FollowService),
		SavedPostHandler:      NewSavedPostHandler(services.SavedPostService),
		NotificationHandler:   NewNotificationHandler(services.NotificationService),
		TagHandler:            NewTagHandler(services.TagService),
		SearchHandler:         NewSearchHandler(services.SearchService),
		MediaHandler:          NewMediaHandler(services.MediaService),
		OAuthHandler:          NewOAuthHandler(services.OAuthService, cfg),
		SEOHandler:            NewSEOHandler(services.PostService, cfg),
		PushHandler:           NewPushHandler(services.PushService),
		ConversationHandler:   NewConversationHandler(services.ConversationService, conversationRepo, chatHub),
		MessageHandler:        NewMessageHandler(services.MessageService, services.MediaService, mediaUploadService, chatHub),
		BlockHandler:          NewBlockHandler(services.BlockService),
		ChatWSHandler:         chatWSHandler,
		NotificationWSHandler: notificationWSHandler,
		FileUploadHandler:     NewFileUploadHandler(services.FileUploadService),
		PresignedUploadHandler: func() *PresignedUploadHandler {
			if r2Storage != nil {
				return NewPresignedUploadHandler(r2Storage, mediaRepo)
			}
			return nil
		}(),
		WebhookHandler:        NewWebhookHandler(services.MediaService, services.PostService, services.MessageService, redisService.(*redis.RedisService), notificationHub),
	}
}