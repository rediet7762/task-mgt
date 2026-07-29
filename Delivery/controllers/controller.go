package controllers

import (
	"errors"
	"net/http"

	"task-manager/Domain"

	"github.com/gin-gonic/gin"
)

// TaskController handles HTTP requests for tasks. It depends only on the
// domain.TaskUsecase interface, so it never touches persistence or business
// rules directly - it just translates HTTP <-> use case calls.
type TaskController struct {
	taskUsecase domain.TaskUsecase
}

// NewTaskController creates a new instance of TaskController.
func NewTaskController(taskUsecase domain.TaskUsecase) *TaskController {
	return &TaskController{taskUsecase: taskUsecase}
}

// GetTasks handles GET /tasks - Get all tasks
// @Summary Get all tasks
// @Description Returns a list of all tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /tasks [get]
func (tc *TaskController) GetTasks(c *gin.Context) {
	status := c.Query("status")

	tasks, err := tc.taskUsecase.GetTasks(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tasks,
		"count": len(tasks),
	})
}

// GetTask handles GET /tasks/:id - Get a specific task
// @Summary Get task by ID
// @Description Returns a single task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /tasks/{id} [get]
func (tc *TaskController) GetTask(c *gin.Context) {
	id := c.Param("id")

	task, err := tc.taskUsecase.GetTask(c.Request.Context(), id)
	if errors.Is(err, domain.ErrTaskNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrInvalidID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": task,
	})
}

// CreateTask handles POST /tasks - Create a new task
// @Summary Create a new task
// @Description Creates a new task with the provided details
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body domain.CreateTaskRequest true "Task details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /tasks [post]
func (tc *TaskController) CreateTask(c *gin.Context) {
	var req domain.CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	task, err := tc.taskUsecase.CreateTask(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    task,
		"message": "Task created successfully",
	})
}

// UpdateTask handles PUT /tasks/:id - Update a specific task
// @Summary Update a task
// @Description Updates an existing task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body domain.UpdateTaskRequest true "Updated task details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /tasks/{id} [put]
func (tc *TaskController) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	task, err := tc.taskUsecase.UpdateTask(c.Request.Context(), id, req)
	if errors.Is(err, domain.ErrTaskNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrInvalidID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    task,
		"message": "Task updated successfully",
	})
}

// DeleteTask handles DELETE /tasks/:id - Delete a specific task
// @Summary Delete a task
// @Description Deletes a task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /tasks/{id} [delete]
func (tc *TaskController) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	err := tc.taskUsecase.DeleteTask(c.Request.Context(), id)
	if errors.Is(err, domain.ErrTaskNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, domain.ErrInvalidID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// UserController handles HTTP requests for authentication and user
// management. It depends only on the domain.UserUsecase interface.
type UserController struct {
	userUsecase domain.UserUsecase
}

// NewUserController creates a new instance of UserController.
func NewUserController(userUsecase domain.UserUsecase) *UserController {
	return &UserController{userUsecase: userUsecase}
}

// Register handles POST /register - Create a new user account
// @Summary Register a new user
// @Description Creates a new user account. The very first user registered becomes an admin.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body domain.RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /register [post]
func (uc *UserController) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	user, err := uc.userUsecase.Register(c.Request.Context(), req)
	if errors.Is(err, domain.ErrUsernameTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    user,
		"message": "User registered successfully",
	})
}

// Login handles POST /login - Authenticate and receive a JWT
// @Summary Log in
// @Description Authenticates a user and returns a JWT for use on protected endpoints.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body domain.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func (uc *UserController) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	token, err := uc.userUsecase.Login(c.Request.Context(), req)
	if errors.Is(err, domain.ErrInvalidCreds) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{"token": token},
		"message": "Login successful",
	})
}

// Promote handles POST /promote/:username - Grant admin rights to a user
// @Summary Promote a user to admin
// @Description Grants the admin role to the given username. Admin-only.
// @Tags auth
// @Accept json
// @Produce json
// @Param username path string true "Username to promote"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /promote/{username} [post]
func (uc *UserController) Promote(c *gin.Context) {
	username := c.Param("username")

	err := uc.userUsecase.Promote(c.Request.Context(), username)
	if errors.Is(err, domain.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": username + " has been promoted to admin",
	})
}
