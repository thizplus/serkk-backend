# Scripts

## Initial User Sync

### Purpose
Syncs all users from Auth Service database (`gofiber_auth`) to Backend Service database (`gofiber_social.users_cache`)

### Prerequisites
1. Both databases must be accessible
2. `users_cache` table must exist in backend database (run migration first)
3. Environment variables configured

### Environment Variables

```bash
# Auth Service Database
AUTH_DB_HOST=localhost
AUTH_DB_PORT=5432
AUTH_DB_NAME=gofiber_auth
AUTH_DB_USER=postgres
AUTH_DB_PASSWORD=your_password

# Backend Service Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gofiber_social
DB_USER=postgres
DB_PASSWORD=your_password
```

### Usage

```bash
# Run the sync script
go run scripts/sync_users_initial.go
```

### What it does
1. Connects to both databases
2. Fetches all users from `gofiber_auth.users`
3. Upserts them into `gofiber_social.users_cache`
4. Reports sync statistics

### Expected Output

```
🔄 Starting initial user sync...
📍 Auth DB: localhost:5432/gofiber_auth
📍 Backend DB: localhost:5432/gofiber_social
✅ Connected to Auth DB
✅ Connected to Backend DB
📊 Found 150 users in Auth DB
📊 Existing users in cache: 0
⏳ Synced 100/150 users...

============================================================
✅ Sync completed!
📊 Total users: 150
✅ Synced: 150
❌ Failed: 0
============================================================

📊 Final count in users_cache: 150
🎉 All users synced successfully!
```

### Notes
- Script uses UPSERT (INSERT ... ON CONFLICT) to handle both new and existing users
- Safe to run multiple times - it won't create duplicates
- Failed syncs are logged with details

---

## Query Foreign Key Constraints

### Purpose
Query all Foreign Key constraints that reference the `users` table. Used before dropping FK constraints in migration 024.

### Usage

```bash
# Connect to your database
psql -U postgres -d gofiber_social

# Run the query
\i scripts/query_fk_constraints.sql
```

Or using any PostgreSQL client, execute:

```bash
psql -U postgres -d gofiber_social -f scripts/query_fk_constraints.sql
```

### Expected Output

```
    Table     |    Column     |     Constraint Name      | Referenced Table | Referenced Column
--------------+---------------+--------------------------+------------------+------------------
blocks        | blocker_id    | fk_blocks_blocker        | users            | id
blocks        | blocked_id    | fk_blocks_blocked        | users            | id
comments      | author_id     | fk_comments_author       | users            | id
follows       | follower_id   | fk_follows_follower      | users            | id
follows       | following_id  | fk_follows_following     | users            | id
messages      | sender_id     | fk_messages_sender       | users            | id
notifications | user_id       | fk_notifications_user    | users            | id
notifications | sender_id     | fk_notifications_sender  | users            | id
posts         | author_id     | fk_posts_author          | users            | id
saved_posts   | user_id       | fk_saved_posts_user      | users            | id
votes         | user_id       | fk_votes_user            | users            | id
```

### Why Drop FK Constraints?

With Auth Service using a separate database (`gofiber_auth`):
- PostgreSQL doesn't support cross-database foreign keys
- These FK constraints will FAIL once `users` table is removed from `gofiber_social`
- We use GORM virtual relations instead
- Application-level referential integrity replaces database-level

### Related Migration

See `migrations/024_drop_users_fk_constraints.up.sql` for the migration that drops these constraints.
