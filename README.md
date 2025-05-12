# Forum

A modern web forum built with Go, featuring user authentication, post management, and real-time interactions.

## Features

- User Authentication (Register/Login)
- Post Creation and Management
- Category-based Post Organization
- Comment System
- Like/Dislike Functionality
- Session Management
- Responsive Design

## Tech Stack

- Go 1.21
- SQLite3
- HTML/CSS/JavaScript
- Docker Support

## Getting Started

### Prerequisites

- Go 1.21 or higher
- SQLite3
- Git

### Installation

1. Clone the repository
```bash
git clone <your-repository-url>
```

2. Navigate to the project directory
```bash
cd forum
```

3. Install dependencies
```bash
go mod download
```

4. Run the application
```bash
go run main.go
```
The server will start at **http://localhost:5000**

### Docker Support
To run the application using Docker:

```bash
docker build -t forum .
docker run -p 5000:8080 forum
```

## Project Structure
- /handlers - HTTP request handlers
- /models - Database models and operations
- /database - Database initialization and migrations
- /utils - Utility functions
- /templates - HTML templates
- /static - Static assets (CSS, JavaScript)

## License
This project is licensed under the MIT License - see the LICENSE file for details.