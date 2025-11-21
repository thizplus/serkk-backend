package di

import (
	"context"
	"gofiber-template/application/serviceimpl"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/postgres"
	"gofiber-template/infrastructure/redis"
	"gofiber-template/infrastructure/storage"
	"gofiber-template/infrastructure/websocket"
	"gofiber-template/infrastructure/workers"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/pkg/ai"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/database"
	"gofiber-template/pkg/health"
	"gofiber-template/pkg/logger"
	"gofiber-template/pkg/metrics"
	"gofiber-template/pkg/scheduler"
	"gorm.io/gorm"
	"log"
)

type Container struct {
	// Configuration
	Config *config.Config

	// Monitoring & Logging
	Logger        *logger.Logger
	HealthChecker *health.HealthChecker
	Metrics       *metrics.Metrics

	// Infrastructure
	DB                 *gorm.DB
	RedisClient        *redis.RedisClient
	RedisService       *redis.RedisService
	FeedCacheService   *redis.FeedCacheService
	BunnyStorage       storage.BunnyStorage
	BunnyStreamService *storage.BunnyStreamService
	R2Storage          storage.R2Storage
	MediaUploadService *storage.MediaUploadService
	EventScheduler     scheduler.EventScheduler
	ChatHub            *websocket.ChatHub
	NotificationHub    *websocket.NotificationHub
	VideoEncoderWorker *workers.VideoEncoderWorker

	// Repositories - Legacy
	UserRepository repositories.UserRepository
	TaskRepository repositories.TaskRepository
	FileRepository repositories.FileRepository
	JobRepository  repositories.JobRepository

	// Repositories - Social Media
	PostRepository                 repositories.PostRepository
	CommentRepository              repositories.CommentRepository
	VoteRepository                 repositories.VoteRepository
	FollowRepository               repositories.FollowRepository
	SavedPostRepository            repositories.SavedPostRepository
	NotificationRepository         repositories.NotificationRepository
	NotificationSettingsRepository repositories.NotificationSettingsRepository
	PushSubscriptionRepository     repositories.PushSubscriptionRepository
	TagRepository                  repositories.TagRepository
	SearchHistoryRepository        repositories.SearchHistoryRepository
	MediaRepository                repositories.MediaRepository

	// Repositories - Chat System
	ConversationRepository repositories.ConversationRepository
	MessageRepository      repositories.MessageRepository
	BlockRepository        repositories.BlockRepository

	// Repositories - Auto-Post
	AutoPostSettingRepository repositories.AutoPostSettingRepository
	AutoPostLogRepository     repositories.AutoPostLogRepository

	// Services - Legacy
	UserService services.UserService
	TaskService services.TaskService
	FileService services.FileService
	JobService  services.JobService

	// Services - Social Media
	PostService         services.PostService
	CommentService      services.CommentService
	VoteService         services.VoteService
	FollowService       services.FollowService
	SavedPostService    services.SavedPostService
	NotificationService services.NotificationService
	PushService         services.PushService
	TagService          services.TagService
	SearchService       services.SearchService
	MediaService        services.MediaService
	OAuthService        services.OAuthService

	// Services - Chat System
	ConversationService services.ConversationService
	MessageService      services.MessageService
	BlockService        services.BlockService

	// Services - Upload
	FileUploadService services.FileUploadService

	// Services - Auto-Post
	AutoPostService       services.AutoPostService
	SimpleAutoPostService services.SimpleAutoPostService
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Initialize() error {
	if err := c.initConfig(); err != nil {
		return err
	}

	// Initialize monitoring first so we can use logger throughout
	if err := c.initMonitoring(); err != nil {
		return err
	}

	if err := c.initInfrastructure(); err != nil {
		return err
	}

	if err := c.initRepositories(); err != nil {
		return err
	}

	// ⭐ Initialize WebSocket hubs BEFORE services (to avoid circular dependency)
	if err := c.initNotificationHub(); err != nil {
		return err
	}

	if err := c.initServices(); err != nil {
		return err
	}

	if err := c.initChatHub(); err != nil {
		return err
	}

	if err := c.initScheduler(); err != nil {
		return err
	}

	if err := c.initVideoEncoderWorker(); err != nil {
		return err
	}

	return nil
}

func (c *Container) initConfig() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	c.Config = cfg
	log.Println("✓ Configuration loaded")
	return nil
}

func (c *Container) initMonitoring() error {
	// Initialize Logger based on environment
	var loggerConfig logger.Config
	if c.Config.App.Env == "development" {
		loggerConfig = logger.DevelopmentConfig()
	} else {
		loggerConfig = logger.ProductionConfig()
	}

	c.Logger = logger.New(loggerConfig)
	logger.Init(loggerConfig) // Initialize global logger
	log.Println("✓ Logger initialized")

	// Initialize Metrics
	c.Metrics = metrics.NewMetrics()
	log.Println("✓ Metrics initialized")

	// Health Checker will be initialized after infrastructure
	// to add database health check
	return nil
}

func (c *Container) initInfrastructure() error {
	// Initialize Database
	dbConfig := postgres.DatabaseConfig{
		Host:     c.Config.Database.Host,
		Port:     c.Config.Database.Port,
		User:     c.Config.Database.User,
		Password: c.Config.Database.Password,
		DBName:   c.Config.Database.DBName,
		SSLMode:  c.Config.Database.SSLMode,
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		return err
	}
	c.DB = db
	log.Println("✓ Database connected")

	// Configure connection pool
	var poolConfig database.PoolConfig
	if c.Config.App.Env == "production" {
		poolConfig = database.ProductionPoolConfig()
	} else {
		poolConfig = database.DevelopmentPoolConfig()
	}
	if err := database.ConfigureConnectionPool(db, poolConfig); err != nil {
		log.Printf("Warning: Failed to configure connection pool: %v", err)
	} else {
		log.Println("✓ Database connection pool configured")
	}

	// Run migrations
	if err := postgres.Migrate(db); err != nil {
		return err
	}
	log.Println("✓ Database migrated")

	// Initialize Health Checker with database check
	c.HealthChecker = health.NewHealthChecker(c.Config.App.Version)
	c.HealthChecker.AddChecker(health.NewDatabaseChecker(db))
	c.HealthChecker.AddChecker(health.NewMemoryChecker(500)) // 500MB threshold
	log.Println("✓ Health checker initialized")

	// Initialize Redis
	redisConfig := redis.RedisConfig{
		Host:     c.Config.Redis.Host,
		Port:     c.Config.Redis.Port,
		Password: c.Config.Redis.Password,
		DB:       c.Config.Redis.DB,
	}
	c.RedisClient = redis.NewRedisClient(redisConfig)

	// Test Redis connection
	if err := c.RedisClient.Ping(context.Background()); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("✓ Redis connected")
	}

	// Initialize RedisService
	c.RedisService = redis.NewRedisService(c.RedisClient)
	log.Println("✓ RedisService initialized")

	// Initialize FeedCacheService
	c.FeedCacheService = redis.NewFeedCacheService(c.RedisClient)
	log.Println("✓ FeedCacheService initialized")

	// Initialize Bunny Storage
	bunnyConfig := storage.BunnyConfig{
		StorageZone: c.Config.Bunny.StorageZone,
		AccessKey:   c.Config.Bunny.AccessKey,
		BaseURL:     c.Config.Bunny.BaseURL,
		CDNUrl:      c.Config.Bunny.CDNUrl,
	}
	c.BunnyStorage = storage.NewBunnyStorage(bunnyConfig)
	log.Println("✓ Bunny Storage initialized")

	// Initialize Bunny Stream
	c.BunnyStreamService = storage.NewBunnyStreamService(
		c.Config.Bunny.StreamAPIKey,
		c.Config.Bunny.StreamLibraryID,
		c.Config.Bunny.StreamCDNURL,
	)
	log.Println("✓ Bunny Stream initialized")

	// Initialize R2 Storage (if configured)
	if c.Config.R2.AccountID != "" && c.Config.R2.AccessKeyID != "" {
		r2Storage, err := storage.NewR2Storage(
			c.Config.R2.AccountID,
			c.Config.R2.AccessKeyID,
			c.Config.R2.SecretAccessKey,
			c.Config.R2.BucketName,
			c.Config.R2.PublicURL,
		)
		if err != nil {
			log.Printf("Warning: R2 Storage initialization failed: %v", err)
		} else {
			c.R2Storage = r2Storage
			log.Println("✓ R2 Storage initialized")
		}
	} else {
		log.Println("ℹ R2 Storage not configured (skipped)")
	}

	// Initialize MediaUploadService
	c.MediaUploadService = storage.NewMediaUploadService(c.BunnyStorage, c.BunnyStreamService)
	log.Println("✓ MediaUploadService initialized")

	return nil
}

func (c *Container) initRepositories() error {
	// Legacy repositories
	c.UserRepository = postgres.NewUserRepository(c.DB)
	c.TaskRepository = postgres.NewTaskRepository(c.DB)
	c.FileRepository = postgres.NewFileRepository(c.DB)
	c.JobRepository = postgres.NewJobRepository(c.DB)

	// Social media repositories
	c.PostRepository = postgres.NewPostRepository(c.DB)
	c.CommentRepository = postgres.NewCommentRepository(c.DB)
	c.VoteRepository = postgres.NewVoteRepository(c.DB)
	c.FollowRepository = postgres.NewFollowRepository(c.DB)
	c.SavedPostRepository = postgres.NewSavedPostRepository(c.DB)
	c.NotificationRepository = postgres.NewNotificationRepository(c.DB)
	c.NotificationSettingsRepository = postgres.NewNotificationSettingsRepository(c.DB)
	c.PushSubscriptionRepository = postgres.NewPushSubscriptionRepository(c.DB)
	c.TagRepository = postgres.NewTagRepository(c.DB)
	c.SearchHistoryRepository = postgres.NewSearchHistoryRepository(c.DB)
	c.MediaRepository = postgres.NewMediaRepository(c.DB)

	// Chat system repositories
	c.ConversationRepository = postgres.NewConversationRepository(c.DB)
	c.MessageRepository = postgres.NewMessageRepository(c.DB)
	c.BlockRepository = postgres.NewBlockRepository(c.DB)

	// Auto-Post repositories
	c.AutoPostSettingRepository = postgres.NewAutoPostSettingRepository(c.DB)
	c.AutoPostLogRepository = postgres.NewAutoPostLogRepository(c.DB)

	log.Println("✓ Repositories initialized (20 repositories)")
	return nil
}

func (c *Container) initServices() error {
	// Legacy services
	c.UserService = serviceimpl.NewUserService(c.UserRepository, c.FollowRepository, c.Config.JWT.Secret)
	c.TaskService = serviceimpl.NewTaskService(c.TaskRepository, c.UserRepository)
	c.FileService = serviceimpl.NewFileService(c.FileRepository, c.UserRepository, c.BunnyStorage)

	// OAuth service
	c.OAuthService = serviceimpl.NewOAuthService(c.UserRepository, c.Config)

	// Social media services (order matters due to dependencies)
	// 1. No service dependencies
	c.TagService = serviceimpl.NewTagService(c.TagRepository)
	c.NotificationService = serviceimpl.NewNotificationService(
		c.NotificationRepository,
		c.NotificationSettingsRepository,
		c.UserRepository,
	)
	c.PushService = serviceimpl.NewPushService(
		c.PushSubscriptionRepository,
		c.Config,
	)

	// 2. Depends on TagService
	c.PostService = serviceimpl.NewPostService(
		c.PostRepository,
		c.UserRepository,
		c.VoteRepository,
		c.SavedPostRepository,
		c.TagService,
		c.MediaRepository,
		c.NotificationHub,
		c.RedisService,
		c.FeedCacheService,
	)

	// 3. Depends on NotificationService
	c.CommentService = serviceimpl.NewCommentService(
		c.CommentRepository,
		c.PostRepository,
		c.VoteRepository,
		c.NotificationService,
	)
	c.VoteService = serviceimpl.NewVoteService(
		c.VoteRepository,
		c.PostRepository,
		c.CommentRepository,
		c.UserRepository,
		c.NotificationService,
	)
	c.FollowService = serviceimpl.NewFollowService(
		c.FollowRepository,
		c.UserRepository,
		c.NotificationService,
	)

	// 4. Independent services
	c.SavedPostService = serviceimpl.NewSavedPostService(
		c.SavedPostRepository,
		c.PostRepository,
		c.VoteRepository,
	)
	c.SearchService = serviceimpl.NewSearchService(
		c.PostRepository,
		c.UserRepository,
		c.TagRepository,
		c.SearchHistoryRepository,
		c.VoteRepository,
		c.SavedPostRepository,
	)
	c.MediaService = serviceimpl.NewMediaService(
		c.MediaRepository,
		c.BunnyStorage,
		c.BunnyStreamService,
		c.RedisService,
	)

	// 5. Chat system services
	c.ConversationService = serviceimpl.NewConversationService(
		c.ConversationRepository,
		c.MessageRepository,
		c.BlockRepository,
		c.UserRepository,
		c.FollowRepository,
		c.RedisService,
	)
	c.MessageService = serviceimpl.NewMessageService(
		c.MessageRepository,
		c.ConversationRepository,
		c.BlockRepository,
		c.UserRepository,
		c.RedisService,
	)
	c.BlockService = serviceimpl.NewBlockService(
		c.BlockRepository,
		c.UserRepository,
	)

	// 6. Upload services
	c.FileUploadService = serviceimpl.NewFileUploadService(
		c.MediaRepository,
		c.MediaUploadService,
	)

	// 7. Auto-Post services
	openAIService := ai.NewOpenAIService(c.Config.OpenAI.APIKey, c.Config.OpenAI.Model)
	c.AutoPostService = serviceimpl.NewAutoPostService(
		c.AutoPostSettingRepository,
		c.AutoPostLogRepository,
		c.PostService,
		openAIService,
		c.Logger.GetZerolog(),
	)

	// Simple Auto-Post service (topic queue)
	c.SimpleAutoPostService = serviceimpl.NewSimpleAutoPostService(
		c.DB,
		openAIService,
		c.PostService,
	)

	// Set push service for notification service (to avoid circular dependency)
	if notifService, ok := c.NotificationService.(*serviceimpl.NotificationServiceImpl); ok {
		notifService.SetPushService(c.PushService)
	}

	log.Println("✓ Services initialized (20 services)")
	return nil
}

func (c *Container) initScheduler() error {
	c.EventScheduler = scheduler.NewEventScheduler()
	c.JobService = serviceimpl.NewJobService(c.JobRepository, c.EventScheduler)

	// Start the scheduler
	c.EventScheduler.Start()
	log.Println("✓ Event scheduler started")

	// Load and schedule existing active jobs
	ctx := context.Background()
	jobs, _, err := c.JobService.ListJobs(ctx, 0, 1000)
	if err != nil {
		log.Printf("Warning: Failed to load existing jobs: %v", err)
		return nil
	}

	activeJobCount := 0
	for _, job := range jobs {
		if job.IsActive {
			err := c.EventScheduler.AddJob(job.ID.String(), job.CronExpr, func() {
				c.JobService.ExecuteJob(ctx, job)
			})
			if err != nil {
				log.Printf("Warning: Failed to schedule job %s: %v", job.Name, err)
			} else {
				activeJobCount++
			}
		}
	}

	if activeJobCount > 0 {
		log.Printf("✓ Scheduled %d active jobs", activeJobCount)
	}

	// Schedule auto-post processing (runs every hour)
	err = c.EventScheduler.AddJob("auto-post-processor", "0 * * * *", func() {
		log.Println("🤖 Running auto-post processor...")
		if err := c.AutoPostService.ProcessAllEnabledSettings(ctx); err != nil {
			log.Printf("❌ Auto-post processor error: %v", err)
		} else {
			log.Println("✓ Auto-post processor completed")
		}
	})
	if err != nil {
		log.Printf("Warning: Failed to schedule auto-post processor: %v", err)
	} else {
		log.Println("✓ Auto-post processor scheduled (every hour)")
	}

	// Schedule simple auto-post from queue (runs every hour)
	err = c.EventScheduler.AddJob("simple-auto-post-processor", "0 * * * *", func() {
		log.Println("📝 Running simple auto-post processor...")
		if err := c.SimpleAutoPostService.ProcessNextTopic(ctx); err != nil {
			log.Printf("❌ Simple auto-post processor error: %v", err)
		} else {
			log.Println("✅ Simple auto-post processor completed")
		}
	})
	if err != nil {
		log.Printf("Warning: Failed to schedule simple auto-post processor: %v", err)
	} else {
		log.Println("✓ Simple auto-post processor scheduled (every hour)")
	}

	return nil
}

func (c *Container) initChatHub() error {
	c.ChatHub = websocket.NewChatHub(
		c.MessageService,
		c.ConversationService,
		c.BlockService,
		c.RedisService,
		c.ConversationRepository,
		c.FollowRepository,
		c.PushService,
	)

	// ⭐ Inject ChatHub back to MessageService (to avoid circular dependency)
	if messageServiceImpl, ok := c.MessageService.(*serviceimpl.MessageServiceImpl); ok {
		messageServiceImpl.SetChatHub(c.ChatHub)
		log.Println("✓ ChatHub injected to MessageService")
	}

	// Start ChatHub in background
	go c.ChatHub.Run()
	log.Println("✓ ChatHub started")

	return nil
}

func (c *Container) initNotificationHub() error {
	c.NotificationHub = websocket.NewNotificationHub()

	// Start NotificationHub in background
	go c.NotificationHub.Run()
	log.Println("✓ NotificationHub started")

	return nil
}

func (c *Container) initVideoEncoderWorker() error {
	c.VideoEncoderWorker = workers.NewVideoEncoderWorker(
		c.RedisService,
		c.BunnyStreamService,
		c.MediaRepository,
		c.NotificationService,
		c.PostService,
		c.NotificationHub,
	)

	// Start worker in background
	c.VideoEncoderWorker.Start()
	log.Println("✓ VideoEncoderWorker started")

	return nil
}

func (c *Container) Cleanup() error {
	log.Println("Starting cleanup...")

	// Stop VideoEncoderWorker
	if c.VideoEncoderWorker != nil {
		c.VideoEncoderWorker.Stop()
		log.Println("✓ VideoEncoderWorker stopped")
	}

	// Stop ChatHub
	if c.ChatHub != nil {
		c.ChatHub.Stop()
		log.Println("✓ ChatHub stopped")
	}

	// Stop scheduler
	if c.EventScheduler != nil {
		if c.EventScheduler.IsRunning() {
			c.EventScheduler.Stop()
			log.Println("✓ Event scheduler stopped")
		} else {
			log.Println("✓ Event scheduler was already stopped")
		}
	}

	// Close Redis connection
	if c.RedisClient != nil {
		if err := c.RedisClient.Close(); err != nil {
			log.Printf("Warning: Failed to close Redis connection: %v", err)
		} else {
			log.Println("✓ Redis connection closed")
		}
	}

	// Close database connection
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Warning: Failed to close database connection: %v", err)
			} else {
				log.Println("✓ Database connection closed")
			}
		}
	}

	log.Println("✓ Cleanup completed")
	return nil
}

func (c *Container) GetServices() (services.UserService, services.TaskService, services.FileService, services.JobService) {
	return c.UserService, c.TaskService, c.FileService, c.JobService
}

func (c *Container) GetConfig() *config.Config {
	return c.Config
}

func (c *Container) GetHandlerServices() *handlers.Services {
	return &handlers.Services{
		// Legacy services
		UserService: c.UserService,
		TaskService: c.TaskService,
		FileService: c.FileService,
		JobService:  c.JobService,

		// Social media services
		PostService:         c.PostService,
		CommentService:      c.CommentService,
		VoteService:         c.VoteService,
		FollowService:       c.FollowService,
		SavedPostService:    c.SavedPostService,
		NotificationService: c.NotificationService,
		PushService:         c.PushService,
		TagService:          c.TagService,
		SearchService:       c.SearchService,
		MediaService:        c.MediaService,
		OAuthService:        c.OAuthService,

		// Chat system services
		ConversationService: c.ConversationService,
		MessageService:      c.MessageService,
		BlockService:        c.BlockService,

		// Upload services
		FileUploadService: c.FileUploadService,

		// Auto-Post services
		AutoPostService: c.AutoPostService,
	}
}

func (c *Container) GetLogger() *logger.Logger {
	return c.Logger
}

func (c *Container) GetHealthChecker() *health.HealthChecker {
	return c.HealthChecker
}

func (c *Container) GetMetrics() *metrics.Metrics {
	return c.Metrics
}
