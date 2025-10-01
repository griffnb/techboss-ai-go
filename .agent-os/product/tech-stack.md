# Technical Stack

> Last Updated: 2025-10-01
> Version: 1.0.0

## Application Framework

- **Framework:** Go 1.23 with go-chi router and go-core framework
- **Version:** Go 1.23

## Database

- **Primary Database:** PostgreSQL
- **NoSQL Database:** DynamoDB for specific use cases
- **Migrations:** Custom database migration system

## Frontend

- **Architecture:** Backend API with separate frontend (frontend not in this repository)
- **API Integration:** ai-sdk compatible AI proxy system for frontend communication
- **JavaScript Framework:** To be determined (separate project)

## Cloud Infrastructure

- **Cloud Provider:** AWS
- **SDK:** AWS SDK v2
- **Storage:** Amazon S3 for document storage
- **Queue System:** Amazon SQS for background task processing
- **Deployment:** Docker containerization with AWS deployment

## Authentication & Authorization

- **Auth Provider:** Clerk
- **Session Management:** Clerk-based authentication
- **Role Management:** Custom role-based access control

## AI/ML Integration

- **AI Providers:** OpenAI, Google Gemini, Anthropic Claude, Azure OpenAI
- **Multi-Provider Support:** Unified interface for multiple AI services
- **Document Processing:** Integrated document analysis and storage

## Development Tools

- **Package Management:** Go modules
- **Code Generation:** Custom code generation tools
- **Testing Framework:** Go standard testing with custom helpers
- **Linting:** golangci-lint
- **Containerization:** Docker

## Architecture Patterns

- **Pattern:** Model-View-Controller (MVC)
- **API Design:** RESTful APIs with go-chi routing
- **Database Access:** Custom ORM with go-core framework
- **Background Processing:** SQS-based task queue system
- **Multi-tenancy:** Organization-based isolation

## Monitoring & Operations

- **Logging:** Structured logging
- **Error Handling:** Custom error handling with context
- **Health Checks:** Built-in health check endpoints
- **Admin Interface:** Full CRUD admin dashboard