-- Query all Foreign Key constraints in gofiber_social database
-- Run this to identify all FK constraints before dropping them

SELECT
    tc.table_name AS "Table",
    kcu.column_name AS "Column",
    tc.constraint_name AS "Constraint Name",
    ccu.table_name AS "Referenced Table",
    ccu.column_name AS "Referenced Column"
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE
    tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
    AND ccu.table_name = 'users'  -- Find all FKs that reference users table
ORDER BY
    tc.table_name, kcu.column_name;

-- Expected results (from our analysis):
-- posts.author_id -> users.id
-- comments.author_id -> users.id
-- follows.follower_id -> users.id
-- follows.following_id -> users.id
-- notifications.user_id -> users.id
-- notifications.sender_id -> users.id
-- votes.user_id -> users.id
-- messages.sender_id -> users.id
-- saved_posts.user_id -> users.id
-- blocks.blocker_id -> users.id
-- blocks.blocked_id -> users.id
