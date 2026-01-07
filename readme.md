
# Project Title

A brief description of what this project does and who it's for

# Capsule Runner

Capsule Runner is a Go-based worker service designed to process time-capsule–style entries stored in a Supabase-backed PostgreSQL database. It identifies capsules that are due for release, retrieves associated media from Supabase Storage, and sends personalized emails with attachments to specified recipients.

---

## Features

* Fetches capsules due for release based on `release_time`
* Uses Supabase PostgreSQL as the database backend
* Streams media files from Supabase Storage buckets
* Sends emails with attachments to multiple recipients via SMTP
* Exposes a protected HTTP endpoint to trigger processing
* Handles failures gracefully by marking capsules for retry

---

## Architecture Overview

1. **Capsule Fetching**
   Queries the database for capsules whose release time has passed and are pending delivery.

2. **Media Retrieval**
   Streams media files directly from Supabase Storage using the service role key.

3. **Email Sending**
   Sends personalized emails with optional attachments using SMTP.

4. **State Updates**

   * Marks capsules as sent on success
   * Marks capsules as failed and retryable on error

---

## Installation

### Clone the Repository

```
git clone <repository-url>
cd capsule-runner
```

### Install Dependencies

```
go mod download
```

### Build the Application

```
go build -o capsule-runner
```

---

## Usage

### Configure Environment Variables

Set the required environment variables using a `.env` file or directly in your environment. See the [Configuration](#configuration) section below.

### Run the Service

```
./capsule-runner
```

The service starts an HTTP server on the configured port (default: `8080`).

### Trigger Capsule Processing

Send a POST request to the worker endpoint:

```
POST /run
Authorization: Bearer <WORKER_SECRET>
```

Successful execution returns:

```
ok
```

---

## Configuration

The following environment variables are required:

| Variable                  | Description                                   |
| ------------------------- | --------------------------------------------- |
| DATABASE_URL              | PostgreSQL connection string for Supabase     |
| SUPABASE_SERVICE_ROLE_KEY | Supabase service role key for storage access  |
| WORKER_SECRET             | Secret for authenticating the worker endpoint |
| PORT                      | HTTP server port (optional, default: 8080)    |
| USER_EMAIL                | Sender email address                          |
| SMTP_HOST                 | SMTP server host                              |
| SMTP_PORT                 | SMTP server port                              |
| SMTP_USER                 | SMTP username                                 |
| SMTP_PASS                 | SMTP password                                 |

### Example `.env`

```
DATABASE_URL=postgresql://user:password@host:5432/dbname
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
WORKER_SECRET=your_worker_secret
PORT=8080
USER_EMAIL=example@email.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=smtp_user
SMTP_PASS=smtp_password
```

---

## API

### POST /run

Processes all capsules that are due for release.

**Headers**

```
Authorization: Bearer <WORKER_SECRET>
```

**Responses**

* `200 OK` — Processing completed successfully
* `401 Unauthorized` — Invalid or missing secret

---

## Dependencies

* `github.com/joho/godotenv`
  Loads environment variables from a `.env` file

* `gopkg.in/gomail.v2`
  SMTP email sending with attachment support

* `github.com/jackc/pgx/v5/stdlib`
  PostgreSQL driver for Go

---

## Error Handling

* Errors during capsule processing do not crash the service
* Failed capsules are marked for retry
* All failures are logged via the fmt for observability not persistant storage
