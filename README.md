# 💸 Tricount-Inspired Expense Sharing App

A simple open-source web application to fairly share group expenses with friends — inspired by [Tricount](https://tricount.com/).  
Built in **Go** and **Postgres** with a modern **React frontend**.  

## ✨ Features

- 👥 **Groups of friends**  
  - Create lists for trips, events, households  
  - Invite your friends with an URL

- 💵 **Track expenses**  
  - Add expenses with amount, payer, participants  
  - Automatically split costs fairly  

- 📊 **Balance calculation**  
  - Always see who owes what  
  - Keep history of all transactions  

- 🌐 **Open source & extensible**  
  - Clean schema & codebase  
  - Easy to fork and customize  

## 🛠️ Tech Stack

- **Backend**: Go, sqlc, Postgres, goose migrations  
- **Frontend**: React (TypeScript)  
- **Auth**: JWT (HS256), refresh tokens in cookies  
- **Infra**: Docker Compose for local dev  

## 🚀 Getting Started

### Prerequisites
- Go ≥ 1.22  
- Node.js ≥ 20  
- Docker + Docker Compose  
- Postgres 15  

### Backend setup
```bash
git clone https://github.com/yourname/expense-share.git
cd expense-share

# run database + migrations
docker compose up -d db
go run cmd/migrate/main.go up

# start API
go run cmd/api/main.go
```

### Frontend setup
```bash
cd web
npm install
npm run dev
```

Visit [http://localhost:3000](http://localhost:3000).  

## 🧪 Testing
```bash
go test ./...
npm test
```

## 📂 Project Structure
```text
.
├── cmd/           # entrypoints (api, migrate)
├── internal/      # Go backend logic
│   ├── db/        # sqlc generated queries
│   └── handlers/  # http handlers for requests
├── migrations/    # goose migrations
└── README.md
```

## Features that I want to implement
- [ ] Log out
- [ ] Frontend
- [ ] Categories

## 📜 License
MIT — free to use, modify, and share.  
