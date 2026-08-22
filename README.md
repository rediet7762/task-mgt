# Task Management API

A RESTful Task Management API built with **Go (Golang)** and designed using **Clean Architecture** principles.

The project provides a structured backend for managing tasks while separating business logic, application use cases, data access, and HTTP delivery.

## 🚀 Features

* Create tasks
* Get all tasks
* Get a task by ID
* Update tasks
* Delete tasks
* RESTful API endpoints
* Request validation
* Clean Architecture structure
* Separation of business logic and infrastructure
* Repository pattern
* Use case layer for application logic
* HTTP API built with Go

## 🛠️ Technologies

* **Go (Golang)**
* **Gin** — HTTP web framework
* **REST API**
* **Clean Architecture**
* **Go Modules**
* **Repository Pattern**

## 🏗️ Project Architecture

The project follows a Clean Architecture approach:

```text
task-mgt/
│
├── Delivery/
│   └── HTTP handlers and routes
│
├── Domain/
│   └── Business entities and interfaces
│
├── Infrastructure/
│   └── External and infrastructure implementations
│
├── Repositories/
│   └── Data access and repository implementations
│
├── Usecases/
│   └── Application business logic
│
├── docs/
│   └── Project documentation
│
├── task-mgt/
│   └── Application entry point
│
├── ARCHITECTURE.md
├── go.mod
├── go.sum
└── .gitignore
```

## 📋 Requirements

Before running the project, make sure you have:

* Go 1.20+ installed
* Git installed

Check your Go installation:

```bash
go version
```

## 📥 Installation

Clone the repository:

```bash
git clone https://github.com/rediet7762/task-mgt.git
```

Navigate into the project:

```bash
cd task-mgt
```

Download the project dependencies:

```bash
go mod download
```

Or:

```bash
go mod tidy
```

## ▶️ Running the Application

Start the application using:

```bash
go run .
```

If your entry point is inside a specific directory, run the corresponding Go file, for example:

```bash
go run ./task-mgt
```

The API will start on the configured server port.

## 🔌 API

The application exposes RESTful endpoints for task management.

Typical operations include:

| Method   | Endpoint     | Description       |
| -------- | ------------ | ----------------- |
| `POST`   | `/tasks`     | Create a new task |
| `GET`    | `/tasks`     | Get all tasks     |
| `GET`    | `/tasks/:id` | Get a task by ID  |
| `PUT`    | `/tasks/:id` | Update a task     |
| `DELETE` | `/tasks/:id` | Delete a task     |

> The exact routes may depend on the current route configuration in the project.

## 🧪 Testing the API

You can test the API using tools such as:

* Postman
* Swagger, if configured
* cURL
* Any REST API client

Example:

```bash
curl http://localhost:8080/tasks
```

## 🧩 Clean Architecture

The application separates responsibilities into different layers.

### Domain

Contains the core business entities and interfaces. The domain layer does not depend on external frameworks or infrastructure.

### Usecases

Contains the application's business logic and coordinates operations between the delivery layer and repositories.

### Repositories

Responsible for data access and persistence implementations.

### Infrastructure

Contains infrastructure-related implementations and external dependencies.

### Delivery

Responsible for receiving HTTP requests, validating input, calling use cases, and returning HTTP responses.

This separation makes the application easier to:

* Test
* Maintain
* Extend
* Refactor
* Scale

## 📁 Main Components

The repository currently contains:

* `Delivery`
* `Domain`
* `Infrastructure`
* `Repositories`
* `Usecases`
* `docs`
* `ARCHITECTURE.md`
* `go.mod`
* `go.sum`

## 🔄 Development Workflow

A typical development workflow is:

```text
Client
   ↓
HTTP Request
   ↓
Delivery
   ↓
Usecase
   ↓
Repository
   ↓
Infrastructure / Database
   ↓
Response
```

## 📌 Future Improvements

Possible future improvements include:

* [ ] Authentication and authorization
* [ ] JWT-based authentication
* [ ] Database persistence
* [ ] Swagger/OpenAPI documentation
* [ ] Unit tests
* [ ] Integration tests
* [ ] Docker support
* [ ] CI/CD pipeline
* [ ] Pagination and filtering
* [ ] Task priorities and statuses
* [ ] User-specific task management

## 👩‍💻 Author

**Rediet Banteyirga**

GitHub: [rediet7762](https://github.com/rediet7762)

## 📄 License

This project is available for educational and development purposes.
