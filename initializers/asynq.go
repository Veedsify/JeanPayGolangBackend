package initializers

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/jobs"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/hibiken/asynq"
)

var redisAddr = libs.GetEnvOrDefault("REDIS_ADDR", "127.0.0.1:6379")

// QueueConfig holds the configuration for the queue server
type QueueConfig struct {
	RedisAddr        string
	Concurrency      int
	StrictPriority   bool
	ShutdownTimeout  time.Duration
	HealthCheckAddr  string
	LogLevel         asynq.LogLevel
	RetryDelayFunc   asynq.RetryDelayFunc
	Queues           map[string]int
	GroupGracePeriod time.Duration
	GroupMaxDelay    time.Duration
	GroupMaxSize     int
}

// QueueServer wraps the asynq server with additional functionality
type QueueServer struct {
	server *asynq.Server
	config *QueueConfig
	mux    *asynq.ServeMux
}

// NewQueueConfig creates a new queue configuration from environment variables
func NewQueueConfig() *QueueConfig {
	return &QueueConfig{
		RedisAddr:        libs.GetEnvOrDefault("REDIS_ADDR", "127.0.0.1:6379"),
		Concurrency:      libs.GetEnvIntOrDefault("QUEUE_CONCURRENCY", 10),
		StrictPriority:   libs.GetEnvBoolOrDefault("QUEUE_STRICT_PRIORITY", false),
		ShutdownTimeout:  time.Duration(libs.GetEnvIntOrDefault("QUEUE_SHUTDOWN_TIMEOUT", 30)) * time.Second,
		HealthCheckAddr:  libs.GetEnvOrDefault("QUEUE_HEALTH_CHECK_ADDR", ":8081"),
		LogLevel:         getLogLevel(),
		RetryDelayFunc:   asynq.DefaultRetryDelayFunc,
		GroupGracePeriod: time.Duration(libs.GetEnvIntOrDefault("QUEUE_GROUP_GRACE_PERIOD", 1)) * time.Minute,
		GroupMaxDelay:    time.Duration(libs.GetEnvIntOrDefault("QUEUE_GROUP_MAX_DELAY", 10)) * time.Minute,
		GroupMaxSize:     libs.GetEnvIntOrDefault("QUEUE_GROUP_MAX_SIZE", 100),
		Queues: map[string]int{
			"critical": libs.GetEnvIntOrDefault("QUEUE_CRITICAL_PRIORITY", 6),
			"high":     libs.GetEnvIntOrDefault("QUEUE_HIGH_PRIORITY", 4),
			"default":  libs.GetEnvIntOrDefault("QUEUE_DEFAULT_PRIORITY", 3),
			"low":      libs.GetEnvIntOrDefault("QUEUE_LOW_PRIORITY", 1),
		},
	}
}

// NewQueueServer creates a new queue server with the given configuration
func NewQueueServer(config *QueueConfig) *QueueServer {
	if config == nil {
		config = NewQueueConfig()
	}

	// Create Redis client options
	redisOpt := asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: libs.GetEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       libs.GetEnvIntOrDefault("REDIS_DB", 0),
	}

	// Configure server options
	serverConfig := asynq.Config{
		Concurrency:     config.Concurrency,
		Queues:          config.Queues,
		StrictPriority:  config.StrictPriority,
		ShutdownTimeout: config.ShutdownTimeout,

		LogLevel:         config.LogLevel,
		RetryDelayFunc:   config.RetryDelayFunc,
		GroupGracePeriod: config.GroupGracePeriod,
		GroupMaxDelay:    config.GroupMaxDelay,
		GroupMaxSize:     config.GroupMaxSize,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Printf("Task failed: type=%s id=%s error=%v", task.Type(), task.ResultWriter().TaskID(), err)

			// Add custom error handling logic here
			// For example, send to monitoring service, log to database, etc.
		}),
	}

	server := asynq.NewServer(redisOpt, serverConfig)

	// Create ServeMux and register handlers
	mux := asynq.NewServeMux()
	registerHandlers(mux)

	return &QueueServer{
		server: server,
		config: config,
		mux:    mux,
	}
}

// registerHandlers registers all email job handlers
func registerHandlers(mux *asynq.ServeMux) {
	// Register email handlers
	mux.HandleFunc(jobs.TypeEmailDelivery, jobs.HandleGenericEmailTask)
	mux.HandleFunc(jobs.TypeWelcomeEmail, jobs.HandleWelcomeEmailTask)
	mux.HandleFunc(jobs.TypePasswordResetEmail, jobs.HandlePasswordResetEmailTask)
	mux.HandleFunc(jobs.TypeTransactionNotification, jobs.HandleTransactionNotificationTask)
	mux.HandleFunc(jobs.TypeEmailVerification, jobs.HandleEmailVerificationTask)
	mux.HandleFunc(jobs.TypeTwoFactorEmail, jobs.HandleTwoFactorEmailTask)
	// Activity Log
	mux.HandleFunc(jobs.TypeActivityLog, jobs.HandleActivityJobTask)
	// Notification Log
	mux.HandleFunc(jobs.TypeNotificationCreate, jobs.HandleCreateNotificationTask)
	mux.HandleFunc(jobs.TypeNotificationDelete, jobs.HandleDeleteNotificationTask)
	mux.HandleFunc(jobs.TypeNotificationUpdate, jobs.HandleUpdateNotificationTask)
	mux.HandleFunc(jobs.TypeNotificationMarkAllRead, jobs.HandleMarkAllNotificationsReadTask)
	mux.HandleFunc(jobs.TypeNotificationMarkRead, jobs.HandleMarkNotificationReadTask)
	// Add middleware for logging
	mux.Use(loggingMiddleware)
	mux.Use(metricsMiddleware)
}

// Start starts the queue server with graceful shutdown
func (qs *QueueServer) Start() error {
	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting queue server with %d workers on %s", qs.config.Concurrency, qs.config.RedisAddr)
		log.Printf("Queue priorities: %+v", qs.config.Queues)

		if err := qs.server.Run(qs.mux); err != nil {
			errChan <- err
		}
	}()

	// Start health check server if configured
	if qs.config.HealthCheckAddr != "" {
		go startHealthCheckServer(qs.config.HealthCheckAddr)
	}

	// Wait for either an error or interrupt signal
	select {
	case err := <-errChan:
		log.Printf("Queue server error: %v", err)
		return err
	case sig := <-quit:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)

		// Shutdown the server gracefully
		qs.server.Shutdown()

		log.Println("Queue server shutdown completed")
		return nil
	}
}

// Stop stops the queue server
func (qs *QueueServer) Stop() {
	qs.server.Shutdown()
}

// InitializeQueueServer initializes and starts the queue server
func InitializeQueueServer() {
	// Initialize the global email sender first
	err := services.InitializeGlobalEmailSender()
	if err != nil {
		log.Fatalf("Failed to initialize global email sender: %v", err)
	}
	log.Println("Global email sender initialized successfully")

	config := NewQueueConfig()
	server := NewQueueServer(config)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start queue server: %v", err)
	}
}

// Middleware functions

// loggingMiddleware logs information about each task
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()

		log.Printf("Processing task: type=%s id=%s", t.Type(), t.ResultWriter().TaskID())

		err := h.ProcessTask(ctx, t)

		duration := time.Since(start)
		if err != nil {
			log.Printf("Task failed: type=%s id=%s duration=%v error=%v",
				t.Type(), t.ResultWriter().TaskID(), duration, err)
		} else {
			log.Printf("Task completed: type=%s id=%s duration=%v",
				t.Type(), t.ResultWriter().TaskID(), duration)
		}

		return err
	})
}

// metricsMiddleware collects metrics about task processing
func metricsMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()

		err := h.ProcessTask(ctx, t)

		duration := time.Since(start)

		// Here you can send metrics to your monitoring system
		// For example: Prometheus, StatsD, etc.
		recordTaskMetrics(t.Type(), duration, err != nil)

		return err
	})
}

// recordTaskMetrics records task processing metrics
func recordTaskMetrics(taskType string, duration time.Duration, failed bool) {
	// Implement your metrics collection here
	// This is a placeholder for metrics collection
	log.Printf("METRICS: task_type=%s duration=%v failed=%v", taskType, duration, failed)
}

// startHealthCheckServer starts a simple HTTP health check server
func startHealthCheckServer(addr string) {
	// This is a placeholder for a health check HTTP server
	// You can implement a proper HTTP server with health check endpoints
	log.Printf("Health check server would start on %s (placeholder)", addr)
}

// getLogLevel returns the log level from environment variable
func getLogLevel() asynq.LogLevel {
	level := libs.GetEnvOrDefault("QUEUE_LOG_LEVEL", "info")
	switch level {
	case "debug":
		return asynq.DebugLevel
	case "info":
		return asynq.InfoLevel
	case "warn":
		return asynq.WarnLevel
	case "error":
		return asynq.ErrorLevel
	case "fatal":
		return asynq.FatalLevel
	default:
		return asynq.InfoLevel
	}
}

// GetQueueStats returns statistics about the queues
func GetQueueStats() (map[string]interface{}, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	stats, err := inspector.GetQueueInfo("default")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"active":    stats.Active,
		"pending":   stats.Pending,
		"scheduled": stats.Scheduled,
		"retry":     stats.Retry,
		"archived":  stats.Archived,
		"completed": stats.Completed,
		"processed": stats.Processed,
		"failed":    stats.Failed,
		"paused":    stats.Paused,
		"timestamp": stats.Timestamp,
	}, nil
}

// PauseQueue pauses a specific queue
func PauseQueue(queueName string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	return inspector.PauseQueue(queueName)
}

// UnpauseQueue unpauses a specific queue
func UnpauseQueue(queueName string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	return inspector.UnpauseQueue(queueName)
}

// DeleteQueue deletes all tasks in a specific queue
func DeleteQueue(queueName string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer inspector.Close()

	return inspector.DeleteQueue(queueName, false)
}
