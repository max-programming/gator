# 🦎 Gator - RSS Feed Aggregator

A powerful command-line RSS feed aggregator built with Go and PostgreSQL. Gator allows you to follow your favorite RSS feeds, aggregate posts, and browse them from your terminal.

## ✨ Features

- 👤 **User Management**: Register and manage multiple users
- 📰 **RSS Feed Management**: Add, follow, and unfollow RSS feeds
- 🔄 **Feed Aggregation**: Automatically fetch and parse RSS feeds at configurable intervals
- 📖 **Post Browsing**: Browse aggregated posts with customizable limits
- 🗄️ **PostgreSQL Storage**: Robust data persistence with PostgreSQL
- ⚡ **Fast CLI Interface**: Efficient command-line interface for all operations

## 🚀 Quick Start

### Prerequisites

Before running Gator, make sure you have the following installed:

- **Go** (version 1.24.6 or later)
- **PostgreSQL** (version 12 or later)
- **sqlc** (for generating type-safe Go code from SQL)
- **goose** (for database migrations)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/max-programming/gator.git
   cd gator
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Install required tools**

   ```bash
   # Install sqlc
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

   # Install goose
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

4. **Set up PostgreSQL database**

   ```bash
   # Create a new database (replace with your preferred name)
   createdb gator
   ```

5. **Run database migrations**

   ```bash
   # Navigate to the schema directory
   cd sql/schema

   # Run migrations (replace connection string with your database URL)
   goose postgres "postgres://username:password@localhost/gator?sslmode=disable" up

   # Return to project root
   cd ../..
   ```

6. **Generate database code**

   ```bash
   sqlc generate
   ```

7. **Build the application**

   ```bash
   go build -o gator
   ```

8. **Create configuration file**
   Create a `.gatorconfig.json` file in your home directory:
   ```json
   {
     "db_url": "postgres://username:password@localhost/gator?sslmode=disable",
     "current_user_name": ""
   }
   ```

## 📖 Usage

### User Management

**Register a new user:**

```bash
./gator register <username>
```

**Login as an existing user:**

```bash
./gator login <username>
```

**List all users:**

```bash
./gator users
```

**Reset all users (⚠️ destructive):**

```bash
./gator reset
```

### Feed Management

**Add a new RSS feed:**

```bash
./gator addfeed <feed_name> <feed_url>
```

Example:

```bash
./gator addfeed "TechCrunch" "https://techcrunch.com/feed/"
```

**List all feeds:**

```bash
./gator feeds
```

**Follow an existing feed:**

```bash
./gator follow <feed_url>
```

**List feeds you're following:**

```bash
./gator following
```

**Unfollow a feed:**

```bash
./gator unfollow <feed_url>
```

### Feed Aggregation

**Start the feed aggregator:**

```bash
./gator agg <time_interval>
```

Example:

```bash
./gator agg 1m    # Fetch feeds every minute
./gator agg 30s   # Fetch feeds every 30 seconds
./gator agg 1h    # Fetch feeds every hour
```

### Browse Posts

**Browse latest posts:**

```bash
./gator browse [limit]
```

Examples:

```bash
./gator browse     # Shows 2 posts by default
./gator browse 10  # Shows 10 latest posts
```

## 🏗️ Project Structure

```
gator/
├── main.go                 # Main application entry point
├── go.mod                  # Go module dependencies
├── sqlc.yaml              # sqlc configuration
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go
│   └── database/          # Generated database code (sqlc)
│       ├── db.go
│       ├── models.go
│       └── *.sql.go
├── sql/
│   ├── schema/            # Database migrations
│   │   ├── 001_users.sql
│   │   ├── 002_feeds.sql
│   │   ├── 003_feed_follows.sql
│   │   └── 005_posts.sql
│   └── queries/           # SQL queries
│       ├── users.sql
│       ├── feeds.sql
│       ├── feed_follows.sql
│       └── posts.sql
└── README.md
```

## 🛠️ Development

### Database Schema

The application uses the following main tables:

- **users**: Store user information
- **feeds**: Store RSS feed metadata
- **feed_follows**: Track which users follow which feeds
- **posts**: Store individual RSS feed posts

### Adding New Features

1. **Database changes**: Add migrations in `sql/schema/`
2. **Queries**: Add SQL queries in `sql/queries/`
3. **Generate code**: Run `sqlc generate`
4. **Implement handlers**: Add command handlers in `main.go`

### Code Generation

This project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries. After modifying SQL files, run:

```bash
sqlc generate
```

## 🚀 Future Features

We're always looking to improve Gator! Here are some exciting features we'd love to add. **Contributions are welcome!**

If you're interested in implementing any of these features, please open an issue to discuss the approach before starting work:

- [ ] **Enhanced Browse Command**

  - [ ] Add sorting options (by date, title, feed)
  - [ ] Add filtering capabilities (by feed, date range, keywords)
  - [ ] Add pagination for better navigation through large post collections

- [ ] **Performance & Scalability**

  - [ ] Add concurrency to the `agg` command for faster feed fetching
  - [ ] Implement parallel feed processing for better performance

- [ ] **Search & Discovery**

  - [ ] Add a `search` command with fuzzy searching capabilities
  - [ ] Full-text search across post titles and descriptions

- [ ] **User Experience**

  - [ ] Add bookmarking/liking functionality for posts
  - [ ] Build a Terminal User Interface (TUI) for better post viewing
  - [ ] Option to open posts in browser directly from TUI

- [ ] **API & Remote Access**

  - [ ] Create HTTP REST API for remote access
  - [ ] Add authentication and authorization system
  - [ ] Enable multi-user remote access

- [ ] **Service Management**
  - [ ] Build a service manager for background `agg` command
  - [ ] Auto-restart functionality if aggregator crashes
  - [ ] Systemd integration for Linux systems

**Have an idea for a new feature?** We'd love to hear about it! Please open an issue to discuss your ideas.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built following the excellent [Build a Blog Aggregator in Go](https://www.boot.dev/courses/build-blog-aggregator-golang) course from [boot.dev](https://www.boot.dev/)
- Built with [Go](https://golang.org/)
- Database powered by [PostgreSQL](https://www.postgresql.org/)
- SQL code generation by [sqlc](https://sqlc.dev/)
- Database migrations by [goose](https://github.com/pressly/goose)
- RSS feed parsing and date handling by various Go libraries

---

**Happy feed reading! 📰✨**
