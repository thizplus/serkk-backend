package models

// User is DEPRECATED in Backend Service
// User management belongs to Auth Service (gofiber_auth DB)
// Backend Service only has identity data (UsersIdentity)
//
// This is an alias to UsersIdentity for backward compatibility
// Gradually refactor code to use UsersIdentity directly
//
// Social-specific data (bio, karma, followers_count) is in UserProfile model
type User = UsersIdentity
