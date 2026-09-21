Architecture

This project is a full-stack social-network application organized as a monorepo.

Technology Stack

Frontend: Next.js + TypeScript

Backend: Go

Database: SQLite

Real-time communication: WebSockets

Containerization: Docker Compose

Project Structure
social-network/
├── README.md
├── docker-compose.yml
├── .gitignore
│
├── frontend/                         # Next.js frontend
│   ├── package.json
│   ├── next.config.ts
│   ├── tsconfig.json
│   ├── public/
│   │   ├── images/
│   │   └── icons/
│   │
│   └── src/
│       ├── app/                      # Application routes
│       │   ├── layout.tsx
│       │   ├── page.tsx
│       │   ├── login/
│       │   │   └── page.tsx
│       │   ├── register/
│       │   │   └── page.tsx
│       │   ├── profile/
│       │   │   └── [username]/
│       │   │       └── page.tsx
│       │   ├── groups/
│       │   │   ├── page.tsx
│       │   │   └── [id]/
│       │   │       └── page.tsx
│       │   ├── messages/
│       │   │   └── page.tsx
│       │   └── notifications/
│       │       └── page.tsx
│       │
│       ├── components/               # Reusable UI components
│       │   ├── layout/
│       │   ├── posts/
│       │   ├── profile/
│       │   ├── groups/
│       │   ├── chat/
│       │   └── notifications/
│       │
│       ├── lib/                      # API, authentication, WebSocket
│       ├── hooks/                    # React application hooks
│       ├── types/                    # TypeScript types
│       └── styles/
│           └── globals.css
│
├── backend/                          # Go backend API
│   ├── go.mod
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   │   ├── sqlite.go
│   │   │   └── migrations/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── services/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── websocket/
│   │
│   └── storage/
│       └── social.db
│
└── docs/
    ├── architecture.md
    ├── api.md
    └── database.md

Backend Architecture

The backend follows a layered architecture:

HTTP Request
     │
     ▼
┌─────────────────┐
│    Handlers     │
│ HTTP/API layer  │
└────────┬────────┘
         ▼
┌─────────────────┐
│    Services     │
│  Business Logic │
└────────┬────────┘
         ▼
┌─────────────────┐
│  Repositories   │
│   Data Access   │
└────────┬────────┘
         ▼
┌─────────────────┐
│     SQLite      │
│    Database     │
└─────────────────┘

Real-Time Communication

WebSocket connections are handled separately from the normal HTTP request flow:

Next.js Client
      │
      │ WebSocket
      ▼
┌─────────────────┐
│  WebSocket Hub  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Chat Service   │
└────────┬────────┘
         ▼
┌─────────────────┐
│ Message         │
│ Repository      │
└────────┬────────┘
         ▼
      SQLite


This is primarily used for real-time messaging and can later be extended to support live notifications, typing indicators, online presence, and other real-time features.

Architecture Principles

The project follows these principles:

Separation of concerns — UI, business logic, and data access are separated.

Repository pattern — database access is isolated from business logic.

Service layer — application/business rules are handled by services.

Reusable frontend components — common UI behavior is shared through components and hooks.

API separation — the Next.js frontend communicates with the Go backend through an API.

Real-time communication — WebSockets are isolated in their own backend package.

Database migrations — database schema changes are managed through migration files.

Data Flow

A typical post request follows this flow:

User
 │
 ▼
Next.js CreatePost
 │
 ▼
API Client
 │
 │ HTTP POST
 ▼
Go Post Handler
 │
 ▼
Post Service
 │
 ▼
Post Repository
 │
 ▼
SQLite


A chat message follows this flow:

User A
 │
 ▼
ChatWindow
 │
 │ WebSocket
 ▼
WebSocket Hub
 │
 ▼
Chat Service
 │
 ├──────────────► Message Repository
 │                       │
 │                       ▼
 │                    SQLite
 │
 └──────────────► User B

Security

Authentication and authorization are handled by the backend.

Security responsibilities include:

Password hashing

Authentication/session management

Authorization checks

Input validation

CORS configuration

Request logging

WebSocket authentication

Protection against unauthorized resource access

Database

SQLite is used for the initial implementation.

Database schema changes are stored in:

backend/internal/database/migrations/


Migration files are numbered sequentially:

001_users.sql
002_posts.sql
003_followers.sql
004_groups.sql
005_notifications.sql
006_messages.sql


The SQLite database file should be treated as runtime data and should normally be excluded from Git:

backend/storage/*.db
backend/storage/*.db-shm
backend/storage/*.db-wal

Documentation

Additional technical documentation is located in:

docs/
├── architecture.md
├── api.md
└── database.md


architecture.md — system architecture and design decisions

api.md — API endpoints and request/response formats

database.md — database schema and relationships