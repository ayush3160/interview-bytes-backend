package server

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayush3160/interview-bytes-backend/pkg/handlers"
	"github.com/ayush3160/interview-bytes-backend/pkg/models"
	"github.com/ayush3160/interview-bytes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultPort = "8080"

func flagInit() {
	models.IsDebugLevel = flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()
}

func setupLogger() *zap.Logger {
	logCfg := zap.NewDevelopmentConfig()

	// Customize the encoder config to put the emoji at the beginning.
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if *models.IsDebugLevel {
		logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logCfg.EncoderConfig.EncodeCaller = nil
	}

	logCfg.DisableStacktrace = true
	logger, err := logCfg.Build()
	if err != nil {
		log.Panic("failed to start the logger for the CLI")
		return nil
	}
	return logger
}

func Start() {
	flagInit()
	logger := setupLogger()

	err := godotenv.Load(".env.local")

	if err != nil {
		logger.Error("Error loading .env file", zap.Error(err))
	}

	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", defaultPort)
	}

	port := os.Getenv("PORT")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://172.25.224.1:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Error("Error connecting to MongoDB", zap.Error(err))
		return
	}

	if err = client.Ping(ctx, nil); err != nil {
		logger.Error("Error connecting to MongoDB", zap.Error(err))
		return
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			logger.Error("Error disconnecting from MongoDB", zap.Error(err))
		}
	}()

	database := os.Getenv("MONGO_DB")
	if database == "" {
		database = "interview-bytes"
	}

	mongoDb := client.Database(database)

	userCollection := mongoDb.Collection("users")

	app := fiber.New()

	app.Use(cors.New())

	// Defining routes for user's
	appSvc := handlers.NewUserHandler(logger, userCollection)
	app.Post("/register", appSvc.CreateUser)
	app.Post("/login", appSvc.Login)

	// Defining routes for room's
	roomSvc := handlers.NewRoomHandler(logger)
	app.Post("/create-room", utils.AuthMiddleware, roomSvc.CreateRoom)
	app.Post("/join-room", roomSvc.CanAccessRoom)
	app.Get("ws/:room_id", websocket.New(roomSvc.JoinRoom))

	app.Get("/user", utils.AuthMiddleware, func(c *fiber.Ctx) error {
		username := c.Locals("username").(string)
		return c.Status(fiber.StatusOK).SendString("Hello " + username)
	})

	logger.Info("Server is listening on", zap.String("port", port))

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		log.Println("Shutting down server...")

		// Shutdown the server gracefully
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		log.Println("Server exiting")
	}()

	if err = app.Listen(":" + port); err != nil {
		logger.Error("Error starting the server", zap.Error(err))
	}

}
