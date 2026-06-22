# Project Workflow Backend (Go / Gin / GORM)

[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg?style=flat-square&logo=go)](https://golang.org/)
[![Gin Gonic](https://img.shields.io/badge/Framework-Gin--Gonic-red.svg?style=flat-square)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/ORM-GORM-orange.svg?style=flat-square)](https://gorm.io/)
[![MySQL](https://img.shields.io/badge/Database-MySQL-blue.svg?style=flat-square&logo=mysql)](https://www.mysql.com/)
[![AWS SDK](https://img.shields.io/badge/AWS-SDK--v2-yellow.svg?style=flat-square&logo=amazon-aws)](https://aws.amazon.com/)

This repository contains the backend engine for a multi-stage **Project Approval and Workflow Management System**. The application is written in Go (Golang) using the **Gin Web Framework** and **GORM ORM**, adhering to clean code principles using a Controller-Service-Repository architecture pattern.

---

## 🚀 Key Features

*   **User Management & Authentication**: JWT-based auth (HS256) with role-based access control (RBAC). Supported roles include `RM` (Relationship Manager), `BH` (Branch Head), and `VH` (Vertical Head).
*   **Multi-Stage Project Approval Workflow**: Projects can be submitted, tracked, and approved sequentially based on a dynamic set of workflow approval steps.
*   **API Security & Middleware**:
    *   JWT token authentication middleware.
    *   API Rate Limiter for request control.
    *   Global Logger middleware using **Logrus**.
    *   Production check blockages (e.g. Postman client checkers).
*   **Cloud Integrations**:
    *   **AWS S3** file manager integrations for attachments and logs.
    *   **AWS Secrets Manager** for securely retrieving application keys.
*   **Mailing System**: SMTP transaction support using **Gomail** for automated notifications.
*   **PDF Generation**: PDF formatting engine using `wkhtmltopdf` integrations.

---

## 📁 Repository Structure

The project follows a standard Golang folder layout for enterprise architectures:

```
├── app/               # Application-level dependency containers & controllers init
├── config/            # MySQL and configuration initializations 
├── controller/        # Request handlers, payload validation & HTTP routing controllers
├── database/          # Database connection pooling & connections
├── dto/               # Data Transfer Objects (Request/Response schemas)
├── log/               # Storage for application log files (Git ignored)
├── middleware/        # Gin middlewares (Auth, rate limiters, logging, timeouts)
├── model/             # GORM models/schemas defining database entities
├── repository/        # Direct database interaction layer (Queries)
├── route/             # Endpoint definitions and router engine setup
├── service/           # Core business logic processing layer
├── util/              # Common helper utilities (AWS, SMTP, Encryption, response formatters)
├── main.go            # Entry point of the application
├── go.mod             # Dependency definition
└── .env               # Configuration variables (Git ignored)
```

---

## 🛠️ Prerequisites

To run this project locally, ensure you have the following installed:

1.  **Go** (v1.25.5 or later)
2.  **MySQL Server** (v8.x recommended)
3.  **wkhtmltopdf** (required for PDF generation features)
    *   **Mac**: `brew install wkhtmltopdf`
    *   **Ubuntu/Debian**: `sudo apt-get install wkhtmltopdf`
    *   **Windows**: Download installer from [wkhtmltopdf.org](https://wkhtmltopdf.org/downloads.html)

---

## ⚙️ Configuration & Environment Variables

Create a `.env` file in the root directory. Below is a template containing all required configuration parameters:

```env
# Application Settings
SITE_TITLE="Attendance & Project Workflow Portal"
APP_ENV="local"            # local | uat | production
APP_PORT=8080
ENABLE_RATE_LIMITER=Y      # Y | N
BURST_LAST_SEEAN=3

# Database Settings
DB_DRIVER="mysql"
DB_HOST="localhost"
DB_PORT="3306"
DB_USERNAME="root"
DB_PASSWORD="your_database_password"
DB_DATABASE_NAME="project_workflow"

# SMTP (Mailing) Settings
SMTP_DRIVER="smtp"
SMTP_SERVER="email-smtp.ap-south-1.amazonaws.com"
SMTP_PORT=587
SMTP_USERNAME="YOUR_SMTP_USERNAME"
SMTP_PASSWORD="YOUR_SMTP_PASSWORD"
SMTP_FROM_ADDRESS="noreply@example.com"

# Security & Encryption Keys
SECRET_KEY="your-jwt-hmac-secret-key"
AES_ENCRYPTION_KEY="your-32-character-aes-key-here"
EMAIL_IV="0123456789abcdef"
ENCRYPTION_KEY="your-16-character-key"

# AWS Cloud Settings (S3 / CDN)
AWS_DEFAULT_REGION="ap-south-1"
AWS_BUCKET="your-s3-bucket-name"
AWS_ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY"
AWS_SECRET_ACCESS_KEY="YOUR_AWS_SECRET_ACCESS_KEY"
AWS_BUCKET_SECRET_CODE="raven"
AWS_CDN_URL="https://your-cdn-domain.com"
```

---

## 🗄️ Database Setup (Schema DDL)

Initialize your MySQL database using the following table structures. The schema matches the GORM entity mappings:

```sql
CREATE DATABASE IF NOT EXISTS project_workflow;
USE project_workflow;

-- 1. Users Table
CREATE TABLE `users` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `uuid` VARCHAR(36) NOT NULL UNIQUE,
  `name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL,
  `role_name` VARCHAR(50) NOT NULL, -- e.g. 'RM', 'BH', 'VH'
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` INT,
  `deleted_at` DATETIME DEFAULT NULL,
  INDEX `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. Workflow Steps Table (Defines approval sequence stages)
CREATE TABLE `workflow_steps` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `step_name` VARCHAR(100) NOT NULL,
  `role_name` VARCHAR(50) NOT NULL,   -- The role authorized to execute this step (RM, BH, VH)
  `step_sequence` INT NOT NULL UNIQUE -- Execution order sequence (e.g. 1, 2, 3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed default workflow stages
INSERT INTO `workflow_steps` (`step_name`, `role_name`, `step_sequence`) VALUES 
('Relationship Manager Review', 'RM', 1),
('Branch Head Approval', 'BH', 2),
('Vertical Head Verification', 'VH', 3);

-- 3. Projects Table
CREATE TABLE `projects` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `uuid` VARCHAR(36) NOT NULL UNIQUE,
  `project_name` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `budget` DECIMAL(15,2) NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'Pending', -- 'Pending', 'Approved', 'Rejected'
  `created_by` INT NOT NULL,
  `updated_by` INT DEFAULT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  INDEX `idx_projects_deleted_at` (`deleted_at`),
  FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 4. Project Approvals Tracking Table
CREATE TABLE `project_approvals` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `project_id` INT NOT NULL,
  `step_id` INT NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'Pending', -- 'Pending', 'Approved', 'Rejected'
  `action_by` INT DEFAULT NULL,
  `action_at` DATETIME DEFAULT NULL,
  `remarks` TEXT,
  FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`step_id`) REFERENCES `workflow_steps` (`id`),
  FOREIGN KEY (`action_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 🏃 Running the Application

1.  **Clone the Repository**:
    ```bash
    git clone https://github.com/yourusername/project-workflow-backend.git
    cd project-workflow-backend
    ```

2.  **Download Dependencies**:
    ```bash
    go mod download
    ```

3.  **Prepare local directories**:
    Ensure `/log` and `/signature` folders exist in root directory (since they are git-ignored):
    ```bash
    mkdir -p log signature
    ```

4.  **Run the Server**:
    By default, the server loads `.env`. You can specify a custom env file if needed:
    ```bash
    # Run with default .env
    go run main.go

    # Run with a custom env configuration flag
    go run main.go -env=.env
    ```

    The backend will start and listen on the port defined by `APP_PORT` (default is `8080`).

---

## 🐳 Docker & Production Deployment

The project contains a production-ready `Dockerfile` and `docker-compose.yml` to run the application with all its database dependencies.

### 1. Run Locally with Docker Compose
You can build and start the Go backend along with a MySQL database container in one command:
```bash
docker compose up --build -d
```
*   The Go backend will connect automatically, auto-migrate database tables, and seed default workflow steps.
*   The Go app is exposed on host port `8080` (`http://localhost:8080`).
*   The MySQL database is mapped on host port `3307` to prevent conflicts with local SQL engines.

### 2. Build & Push to AWS ECR (UAT Release)
We have added a helper script `./build_and_push.sh` to automate the local build, ECR login, tagging, and upload process:
```bash
# Push default version (v0.0.6.5)
./build_and_push.sh

# Or push a custom version tag
./build_and_push.sh v0.0.6.6
```

### 3. Server Deployment Commands (UAT)
On your target server, authenticate with ECR, pull the image, and spin up the container:

```bash
# 1. Login to ECR
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin 591051854019.dkr.ecr.ap-south-1.amazonaws.com

# 2. Pull the UAT image
docker pull 591051854019.dkr.ecr.ap-south-1.amazonaws.com/shiv-mitra:v0.0.6.5-attendance_cmrf-uat

# 3. Stop old containers (if applicable)
docker stop <old_container_ids_or_names>

# 4. Spin up the new container with correct port and log volume mappings
docker run --publish 8084:8084 \
  --restart always \
  --name shiv-mitra-v0.0.6.5-attendance_cmrf-uat \
  -d \
  -e APP_PORT=8084 \
  -e LOG_DIR=/app/logs \
  -v /var/logs/container/shiv-mitra:/app/logs \
  591051854019.dkr.ecr.ap-south-1.amazonaws.com/shiv-mitra:v0.0.6.5-attendance_cmrf-uat
```
> [!NOTE]
> *   `-e APP_PORT=8084`: Tells the Go app inside the container to listen on port `8084`, matching host network mapping.
> *   `-e LOG_DIR=/app/logs`: Instructs the Go logger to write logs directly to the mounted directory.
> *   `-v /var/logs/container/shiv-mitra:/app/logs`: Persists container logs on the host.

### 4. Container Management & Orchestration (Cheat Sheet)

In multi-container environments, Docker Compose abstracts raw container manipulation. Instead of copying generated **Container IDs** (e.g. `601fcf9d3fc3`), you can manage services directly using their logical service names (`backend` or `db`) defined in `docker-compose.yml`.

#### 🔍 Service Discovery & Metadata
Use these commands to locate container details, check network ports, and extract IDs:

*   **View Active Container Details:**
    ```bash
    docker ps
    ```
    *Lists all running containers on the system (including ID, Name, Uptime, and Port bindings).*
*   **Extract Compose Container IDs:**
    ```bash
    docker compose ps -q
    ```
    *Returns only the raw hex container IDs of the services in this compose project (highly useful for scripting).*
*   **Inspect Container Configuration JSON:**
    ```bash
    docker inspect <container_id_or_name>
    ```
    *Fetches the raw configuration details, environment variables, mounts, and network IPs.*

#### 🛠️ Service Lifecycle Operations
Control specific services directly using their compose service names without looking up IDs:

*   **Stop a Specific Service:**
    ```bash
    docker compose stop backend
    ```
*   **Start a Specific Service:**
    ```bash
    docker compose start backend
    ```
*   **Restart a Specific Service:**
    ```bash
    docker compose restart backend
    ```

#### 📊 Monitoring & Interactive Debugging
Monitor server logs in real-time or access a shell inside the container virtual environment:

*   **Stream Service Logs (Real-time tailing):**
    ```bash
    docker compose logs -f db
    ```
*   **Stream Last 50 Log Lines:**
    ```bash
    docker compose logs --tail 50 -f backend
    ```
*   **Access Container Shell Environment:**
    ```bash
    docker compose exec backend sh
    ```
    *Opens an interactive shell terminal (`sh`) inside the container sandbox to run inspection commands.*

#### 🧹 Infrastructure Cleanup
Perform cleanup operations to reclaim system disk space:

*   **Stop & Delete Stack Containers/Networks:**
    ```bash
    docker compose down
    ```
*   **Stop & Delete Stack Containers, Networks, and Volumes:**
    ```bash
    docker compose down -v
    ```
    *WARNING: This deletes the MySQL database data volume.*
*   **Prune Dangling Containers and Images:**
    ```bash
    docker system prune -f
    ```

---

## 🔌 API Endpoints Documentation

All requests and responses use the `application/json` format. Headers must supply the `Authorization` header containing the JWT token for all protected endpoints.

### Authentication & Users API

| Method | Endpoint | Auth Required | Description |
| :--- | :--- | :---: | :--- |
| **POST** | `/api/users/create` | ❌ No | Registers a new user. |
| **POST** | `/api/users/login` | ❌ No | Authenticates user credentials and issues a JWT token. |
| **POST** | `/api/users/list` | 🔒 Yes | Retrieves a list of users (Protected). |
| **POST** | `/api/users/details` | 🔒 Yes | Fetches user details by UUID payload. |
| **POST** | `/api/users/update` | 🔒 Yes | Modifies existing user metadata. |
| **POST** | `/api/users/delete` | 🔒 Yes | Soft-deletes a user from the workspace. |

#### Example payloads:

*   **User Registration (`POST /api/users/create`)**:
    ```json
    {
      "name": "Jane Doe",
      "email": "janedoe@example.com",
      "password": "securepassword123",
      "role_name": "BH"
    }
    ```

*   **User Login (`POST /api/users/login`)**:
    ```json
    {
      "email": "janedoe@example.com",
      "password": "securepassword123"
    }
    ```

---

### Projects & Workflow API

| Method | Endpoint | Auth Required | Description |
| :--- | :--- | :---: | :--- |
| **POST** | `/api/projects/create` | 🔒 Yes | Creates a new project in the pipeline. |
| **POST** | `/api/projects/list` | 🔒 Yes | Lists projects (Returns projects corresponding to the user's role). |
| **POST** | `/api/projects/details` | 🔒 Yes | Gets detailed data of a project, including approval log stack. |
| **POST** | `/api/projects/update` | 🔒 Yes | Updates project specifications. |
| **POST** | `/api/projects/approve` | 🔒 Yes | Registers an approval step action (Approve/Reject). |
| **POST** | `/api/projects/approval-update` | 🔒 Yes | Modifies specific step approvals (Manual Override). |

#### Example payloads:

*   **Create Project (`POST /api/projects/create`)**:
    ```json
    {
      "project_name": "Global ERP System",
      "description": "Enterprise resource planning integration software.",
      "budget": 1250000.00
    }
    ```

*   **Approve/Reject Step (`POST /api/projects/approve`)**:
    ```json
    {
      "project_uuid": "c3b07384-d113-4a61-9c3f-42e185c7a6e1",
      "status": "Approved", 
      "remarks": "Budget falls within parameters, moving to Branch Head."
    }
    ```

---

## 🐳 Docker Deployment

The application is dockerized with a multi-stage Go build and containerized MySQL database using Docker Compose.

### Prerequisites
- [Docker](https://www.docker.com/get-started) and Docker Compose installed.
- A configured `.env` file in the root directory (all values will be automatically loaded and forwarded to the backend service).

### Running the Stack
To build the Docker images and run the full backend stack (Go application + MySQL database):

```bash
docker compose up --build -d
```

This will:
1. Build the multi-stage Go container.
2. Spin up a MySQL 8.0 container.
3. Automatically configure the database hostname overrides (`DB_HOST=db`).
4. Mount the local `./log` directory to persist application logs.
5. Expose the API on host port `8080` (or `APP_PORT` from your `.env`).

### Useful Commands

- **Stop Services**:
  ```bash
  docker compose down
  ```
- **Stop Services & Clean Volumes**:
  ```bash
  docker compose down -v
  ```
- **View Container Logs**:
  ```bash
  docker compose logs -f
  ```
- **Access Database inside Container**:
  ```bash
  docker exec -it project-workflow-db mysql -u root -proot@123 project_workflow
  ```

---

## 🔒 Security Practices

1.  **Environment Isolation**: Sensitive credentials (database logins, AWS access tokens, SMTP keys) should remain strictly inside the `.env` file and must never be pushed to your version control repository.
2.  **Password Safety**: User passwords are saved as hashed strings utilizing the `bcrypt` library before writing to the database.
3.  **Token Expiry**: Auth JWT tokens are short-lived (configured for 24-hour expiration) and require cryptographic validation.

---

## 📝 License

This project is licensed under the [MIT License](LICENSE). Feel free to modify and adapt it to your workflow!
