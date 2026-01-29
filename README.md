# Go_Student_Registry🎓

A robust, lightweight RESTful API for student management built entirely with **Go (Golang)** and **SQLite**.

This project demonstrates how to build a production-ready web service using the **Go Standard Library** (specifically Go 1.22+ routing enhancements) without relying on heavy third-party web frameworks like Gin or Echo. It features structured logging, clean architecture, and graceful shutdowns.

## 🚀 Features

* **RESTful CRUD Operations:** Create, Read, Update, and Delete student records.
* **Standard Library Routing:** Utilizes Go 1.22+ `http.ServeMux` with method-based matching (e.g., `POST /api/students`).
* **Structured Logging:** Implements `log/slog` for JSON-formatted, level-based logging.
* **Persistent Storage:** Uses **SQLite** for a self-contained, lightweight database solution.
* **Clean Architecture:** Modular project structure separating Handlers, Storage, and Configuration.
* **Graceful Shutdown:** Handles OS signals (`SIGINT`, `SIGTERM`) to ensure database connections and requests close properly.

## 🛠️ Tech Stack

* **Language:** Go (1.22+)
* **Database:** SQLite 3
* **Router:** `net/http` (Standard Lib)
* **Logging:** `log/slog` (Standard Lib)
* **Config Management:** Custom loader (YAML/Env)

## 📂 Project Structure

```bash
go_stuAPI/
├── cmd/
│   └── stuAPI/        # Main application entry point
├── internal/
│   ├── config/        # Configuration loading logic
│   ├── http/
│   │   └── handlers/  # HTTP handlers (Controllers)
│   └── storage/
│       └── sqlite/    # Database interaction layer
├── main.go            # Entry point
└── go.mod             # Dependencies

```

## 🔌 API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/api/students` | Create a new student |
| `GET` | `/api/students/{id}` | Retrieve a specific student by ID |
| `GET` | `/api/students/` | Retrieve a list of all students |
| `PUT` | `/api/students/{id}` | Update an existing student's details |
| `DELETE` | `/api/students/{id}` | Remove a student from the database |

## ⚙️ Getting Started

### Prerequisites

* [Go](https://go.dev/dl/) installed (version 1.22 or higher).
* Git.

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/Sarthak-D97/go_stuAPI.git
cd go_stuAPI

```


2. **Download dependencies:**
```bash
go mod tidy

```


3. **Configuration:**
Ensure you have a configuration file set up (e.g., `config/local.yaml`) or environment variables set as required by your `internal/config` package.
4. **Run the application:**
```bash
go run main.go

```


*You should see a log message indicating the server has started.*

### Testing the API

You can test the endpoints using `curl` or Postman.

**Example: Create a Student**

```bash
curl -X POST http://localhost:8080/api/students \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "age": 22}'

```

## 🤝 Contributing

Contributions are welcome! Please fork the repository and open a pull request with your changes.

## 📄 License

This project is open-source and available under the [MIT License](https://www.google.com/search?q=LICENSE).

---