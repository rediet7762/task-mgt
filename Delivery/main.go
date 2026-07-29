// Command task-manager starts the Task Management API HTTP server.
//
// This file is the application's composition root: it is deliberately the
// only place that imports Infrastructure, Repositories, Usecases, and
// Delivery/controllers all at once, wiring concrete implementations into
// the domain interfaces that every other layer depends on. Nothing below
// this layer knows that MongoDB, Gin, or JWT even exist.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	domain "task-manager/Domain"
	"task-manager/Delivery/controllers"
	"task-manager/Delivery/routers"
	infrastructure "task-manager/Infrastructure"
	repositories "task-manager/Repositories"
	usecases "task-manager/Usecases"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	loadEnvFile()

	var (
		taskRepo domain.TaskRepository
		userRepo domain.UserRepository
	)

	client, db, err := connectMongo()
	if err != nil {
		log.Printf("⚠️ MongoDB unavailable, starting with local in-memory storage: %v", err)
		taskRepo = repositories.NewInMemoryTaskRepository()
		userRepo = repositories.NewInMemoryUserRepository()
	} else {
		defer func() {
			if client != nil {
				_ = client.Disconnect(context.Background())
			}
		}()
		log.Println("✅ Successfully connected to MongoDB!")

		collectionName := os.Getenv("MONGODB_COLLECTION")
		if collectionName == "" {
			collectionName = "tasks"
		}
		taskCollection := db.Collection(collectionName)
		userCollection := db.Collection("users")

		// --- Repositories (implement Domain interfaces over MongoDB) ---
		taskRepo = repositories.NewTaskRepository(taskCollection)
		userRepo = repositories.NewUserRepository(userCollection)
	}

	// --- Infrastructure ---
	jwtService := infrastructure.NewJWTService(os.Getenv("JWT_SECRET"))
	passwordService := infrastructure.NewPasswordService()

	// --- Usecases (business logic, depend only on Domain interfaces) ---
	taskUsecase := usecases.NewTaskUsecase(taskRepo)
	userUsecase := usecases.NewUserUsecase(userRepo, passwordService, jwtService)

	// Seed sample data directly through the repository - seeding is a
	// deployment/bootstrap concern, not a use case a client ever invokes.
	if err := taskRepo.SeedSampleTasks(context.Background()); err != nil {
		log.Printf("⚠️ Failed to seed sample tasks: %v", err)
	} else {
		log.Println("✅ Sample tasks are ready.")
	}

	// --- Delivery: controllers ---
	taskController := controllers.NewTaskController(taskUsecase)
	userController := controllers.NewUserController(userUsecase)

	// --- Delivery: HTTP server ---
	r := gin.Default()
	r.Use(corsMiddleware())

	authMW := infrastructure.AuthMiddleware(jwtService)
	adminMW := infrastructure.AdminOnly()
	routers.SetupRouter(r, taskController, userController, authMW, adminMW)

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📚 API Documentation available at /docs")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// connectMongo reads MONGODB_URI/MONGODB_DB from the environment, connects,
// and pings to confirm the connection is live before returning.
func connectMongo() (*mongo.Client, *mongo.Database, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, nil, fmt.Errorf("MONGODB_URI is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ping mongodb: %w", err)
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "task_manager"
	}

	return client, client.Database(dbName), nil
}

// corsMiddleware allows the API to be called from browser-based frontends
// during development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// loadEnvFile loads simple KEY=VALUE pairs from a .env file in the working
// directory, without overriding any variable already set in the real
// environment. Missing the file is not an error - it just means
// configuration is expected to come from the environment directly (e.g. in
// production).
func loadEnvFile() {
	path := filepath.Join(".", ".env")
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
