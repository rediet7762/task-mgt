// Package routers wires HTTP routes to controllers. It knows nothing about
// databases, JWTs, or business rules - it only needs a Gin engine, the
// controllers to dispatch to, and the auth middleware to gate protected
// routes. All of the wiring that used to live here (Mongo connection,
// service construction) now happens in Delivery/main.go, which is the only
// place in the application allowed to know about every layer at once.
package routers

import (
	"task-manager/Delivery/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures every route on the given engine: public auth
// routes, task routes that require any authenticated user, and
// task/promotion routes restricted to admins.
//
// Route map:
//
//	POST   /api/v1/register           public
//	POST   /api/v1/login               public
//	GET    /api/v1/tasks/              authenticated
//	GET    /api/v1/tasks/:id           authenticated
//	POST   /api/v1/tasks/              authenticated + admin
//	PUT    /api/v1/tasks/:id           authenticated + admin
//	DELETE /api/v1/tasks/:id           authenticated + admin
//	POST   /api/v1/promote/:username   authenticated + admin
func SetupRouter(
	r *gin.Engine,
	tc *controllers.TaskController,
	uc *controllers.UserController,
	authMiddleware gin.HandlerFunc,
	adminOnly gin.HandlerFunc,
) {
	v1 := r.Group("/api/v1")

	// Public auth routes - no token required.
	v1.POST("/register", uc.Register)
	v1.POST("/login", uc.Login)

	// Task routes require a valid JWT.
	taskGroup := v1.Group("/tasks")
	taskGroup.Use(authMiddleware)
	{
		// Read access is open to any authenticated user, admin or not.
		taskGroup.GET("/", tc.GetTasks)
		taskGroup.GET("/:id", tc.GetTask)

		// Write access requires the admin role.
		adminTasks := taskGroup.Group("/")
		adminTasks.Use(adminOnly)
		{
			adminTasks.POST("/", tc.CreateTask)
			adminTasks.PUT("/:id", tc.UpdateTask)
			adminTasks.DELETE("/:id", tc.DeleteTask)
		}
	}

	// Promoting a user to admin is itself an admin-only action.
	adminGroup := v1.Group("/")
	adminGroup.Use(authMiddleware, adminOnly)
	{
		adminGroup.POST("/promote/:username", uc.Promote)
	}
}
