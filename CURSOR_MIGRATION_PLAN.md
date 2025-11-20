# 🎯 แผนการปรับระบบเป็น Cursor-Based Pagination

วันที่จัดทำ: 2025-11-14
เวอร์ชัน: 1.0
ผู้รับผิดชอบ: Development Team

---

## 📋 สารบัญ

1. [สรุปภาพรวม](#1-สรุปภาพรวม)
2. [การจัดกลุ่ม Endpoints](#2-การจัดกลุ่ม-endpoints)
3. [ลำดับความสำคัญ Priority](#3-ลำดับความสำคัญ-priority)
4. [แผนการทำงานทีละ Phase](#4-แผนการทำงานทีละ-phase)
5. [Checklist แต่ละ Endpoint](#5-checklist-แต่ละ-endpoint)
6. [Timeline รวม](#6-timeline-รวม)
7. [Resource Requirements](#7-resource-requirements)
8. [Risk & Mitigation](#8-risk--mitigation)

---

## 1. สรุปภาพรวม

### 📊 สถิติ Pagination Endpoints

```
Total Pagination Endpoints: 35 endpoints

✅ ใช้ Cursor อยู่แล้ว:        4 endpoints (11%)
🔴 ควรเปลี่ยนเป็น Cursor:    14 endpoints (40%)  ← HIGH PRIORITY
🟡 พิจารณาเปลี่ยน:           10 endpoints (29%)  ← MEDIUM PRIORITY
⚪ ไม่จำเป็นเปลี่ยน:          7 endpoints (20%)   ← KEEP OFFSET
```

### 🎯 เป้าหมาย

1. **User Experience**: ให้ feed ทุก feed ทำงานแบบ Facebook/Instagram (no duplicates, smooth scrolling)
2. **Performance**: เพิ่มความเร็วในการดึงข้อมูล 100-1000 เท่า สำหรับ deep pagination
3. **Scalability**: รองรับการเติบโตของข้อมูลในอนาคต
4. **Consistency**: ใช้ pagination pattern เดียวกันทั้งระบบ

---

## 2. การจัดกลุ่ม Endpoints

### ✅ กลุ่ม A: ใช้ Cursor อยู่แล้ว (4 endpoints)

**สถานะ**: ✅ DONE - ใช้เป็นตัวอย่างสำหรับ endpoints อื่น

| # | Endpoint | Handler | Status |
|---|----------|---------|--------|
| 1 | `GET /conversations` | ConversationHandler.ListConversations | ✅ Cursor |
| 2 | `GET /conversations/:id/messages` | MessageHandler.ListMessages | ✅ Cursor |
| 3 | `GET /conversations/:id/media` | MessageHandler.ListMediaMessages | ✅ Cursor |
| 4 | `GET /conversations/:id/links` | MessageHandler.ListMessagesWithLinks | ✅ Cursor |
| 5 | `GET /conversations/:id/files` | MessageHandler.ListFileMessages | ✅ Cursor |

**Note**: ใช้เป็นแบบอย่าง (reference implementation) สำหรับ endpoints อื่น

---

### 🔴 กลุ่ม B: HIGH PRIORITY - ควรเปลี่ยนเป็น Cursor ทันที (14 endpoints)

**เหตุผล**: เป็น social feed features ที่ผู้ใช้งานบ่อย, มีการเพิ่ม/ลบข้อมูลตลอดเวลา

#### B1: Posts & Feed (6 endpoints) ⭐ **HIGHEST PRIORITY**

| # | Endpoint | Handler | Current | Target | Impact |
|---|----------|---------|---------|--------|--------|
| 1 | `GET /posts` | PostHandler.ListPosts | Offset | Cursor | 🔥 CRITICAL |
| 2 | `GET /posts/feed` | PostHandler.GetFeed | Offset | Cursor | 🔥 CRITICAL |
| 3 | `GET /posts/tag/:tagName` | PostHandler.ListPostsByTag | Offset | Cursor | 🔥 HIGH |
| 4 | `GET /posts/tag/:tagId` | PostHandler.ListPostsByTagID | Offset | Cursor | 🔥 HIGH |
| 5 | `GET /posts/author/:authorId` | PostHandler.ListPostsByAuthor | Offset | Cursor | 🔥 HIGH |
| 6 | `GET /posts/:id/crossposts` | PostHandler.GetCrossposts | Offset | Cursor | 🟡 MEDIUM |

**Business Impact**:
- ❌ ปัญหาปัจจุบัน: Users เห็นโพสต์ซ้ำเมื่อมีคนโพสต์ใหม่
- ✅ หลังแก้: Smooth infinite scroll เหมือน Facebook
- 📈 Expected improvement: 100-1000x faster สำหรับ deep scrolling

#### B2: Comments (3 endpoints)

| # | Endpoint | Handler | Current | Target | Impact |
|---|----------|---------|---------|--------|--------|
| 7 | `GET /comments/post/:postId` | CommentHandler.ListCommentsByPost | Offset | Cursor | 🔥 HIGH |
| 8 | `GET /comments/author/:authorId` | CommentHandler.ListCommentsByAuthor | Offset | Cursor | 🟡 MEDIUM |
| 9 | `GET /comments/:id/replies` | CommentHandler.ListReplies | Offset | Cursor | 🔥 HIGH |

**Business Impact**:
- Comment threads ยาว ๆ จะ scroll ได้เร็วขึ้น
- ไม่เห็น comment ซ้ำเมื่อมีคนแสดงความคิดเห็นใหม่

#### B3: Social Features (5 endpoints)

| # | Endpoint | Handler | Current | Target | Impact |
|---|----------|---------|---------|--------|--------|
| 10 | `GET /follows/user/:userId/followers` | FollowHandler.GetFollowers | Offset | Cursor | 🟡 MEDIUM |
| 11 | `GET /follows/user/:userId/following` | FollowHandler.GetFollowing | Offset | Cursor | 🟡 MEDIUM |
| 12 | `GET /follows/mutual` | FollowHandler.GetMutualFollows | Offset | Cursor | 🟡 MEDIUM |
| 13 | `GET /saved/posts` | SavedPostHandler.GetSavedPosts | Offset | Cursor | 🟡 MEDIUM |
| 14 | `GET /notifications` | NotificationHandler.GetNotifications | Offset | Cursor | 🔥 HIGH |

**Business Impact**:
- Followers/Following lists โหลดเร็วขึ้น
- Notifications แบบ real-time (เหมือน Facebook bell icon)

---

### 🟡 กลุ่ม C: MEDIUM PRIORITY - พิจารณาเปลี่ยน (10 endpoints)

**เหตุผล**: ใช้งานไม่บ่อยมาก, หรือข้อมูลไม่ได้เปลี่ยนแปลงตลอดเวลา

#### C1: Search & Discovery (4 endpoints)

| # | Endpoint | Handler | Current | Consideration | Decision |
|---|----------|---------|---------|---------------|----------|
| 1 | `GET /posts/search` | PostHandler.SearchPosts | Offset | Cursor? | 🟡 Consider |
| 2 | `GET /tags` | TagHandler.ListTags | Offset | Cursor? | 🟡 Consider |
| 3 | `GET /search/history` | SearchHandler.GetSearchHistory | Offset | Cursor? | 🟡 Consider |
| 4 | `GET /notifications/unread` | NotificationHandler.GetUnreadNotifications | Offset | Cursor | 🔥 HIGH |

**Recommendation**:
- ✅ **เปลี่ยนเป็น Cursor**: Search history, Unread notifications
- ⚪ **พิจารณา**: Tags list (ข้อมูล static)

#### C2: User Content (4 endpoints)

| # | Endpoint | Handler | Current | Consideration |
|---|----------|---------|---------|---------------|
| 5 | `GET /votes/user` | VoteHandler.GetUserVotes | Offset | Cursor? |
| 6 | `GET /media/user/:userId` | MediaHandler.GetUserMedia | Offset | Cursor? |
| 7 | `GET /blocks` | BlockHandler.ListBlockedUsers | Offset | Cursor? |
| 8 | `GET /users/search` | UserHandler.ListUsers | Offset | Cursor? |

**Recommendation**:
- ✅ **เปลี่ยนเป็น Cursor**: User votes, Media library
- 🟡 **อาจเปลี่ยน**: Blocked users
- ⚪ **ไม่จำเป็น**: User search (มี filters ซับซ้อน)

#### C3: Legacy Features (2 endpoints)

| # | Endpoint | Handler | Current | Decision |
|---|----------|---------|---------|----------|
| 9 | `GET /tasks` | TaskHandler.ListTasks | Offset | ⚪ KEEP (Legacy) |
| 10 | `GET /files` | FileHandler.ListFiles | Offset | ⚪ KEEP (Legacy) |

---

### ⚪ กลุ่ม D: KEEP OFFSET - ไม่จำเป็นเปลี่ยน (7 endpoints)

**เหตุผล**: Admin tools, internal features, หรือข้อมูล static ที่ต้องการ page numbers

| # | Endpoint | Handler | Reason | Keep Offset |
|---|----------|---------|--------|-------------|
| 1 | `GET /admin/jobs` | JobHandler.ListJobs | Admin tool, need pagination | ✅ |
| 2 | `GET /admin/users` | UserHandler.ListUsers | Admin tool, need filters | ✅ |
| 3 | `GET /tasks/user` | TaskHandler.GetUserTasks | Legacy feature | ✅ |
| 4 | `GET /files/user` | FileHandler.GetUserFiles | Legacy feature | ✅ |
| 5 | RSS Feed | SEOHandler.GetRSSFeed | RSS format requirement | ✅ |
| 6 | Sitemap | SEOHandler (if exists) | SEO requirement | ✅ |

**Recommendation**: เก็บ offset-based ไว้สำหรับ admin tools และ legacy features

---

## 3. ลำดับความสำคัญ (Priority)

### 🔥 Phase 1: CRITICAL - Posts & Feed (2 weeks)

**Target**: 6 endpoints
**Impact**: ผู้ใช้ 100% ได้ประโยชน์
**Complexity**: High (ต้องทำ hot score algorithm)

```
Week 1-2:
✅ GET /posts
✅ GET /posts/feed
✅ GET /posts/tag/:tagName
✅ GET /posts/tag/:tagId
✅ GET /posts/author/:authorId
✅ GET /posts/:id/crossposts
```

**Dependencies**:
- [ ] Database indexes (composite indexes)
- [ ] Cursor utilities for posts
- [ ] Hot score calculation
- [ ] Service layer refactor
- [ ] Frontend update (React infinite scroll)

---

### 🔥 Phase 2: HIGH - Comments & Notifications (1.5 weeks)

**Target**: 5 endpoints
**Impact**: Better comment threads, real-time notifications
**Complexity**: Medium

```
Week 3-4:
✅ GET /comments/post/:postId
✅ GET /comments/:id/replies
✅ GET /comments/author/:authorId
✅ GET /notifications
✅ GET /notifications/unread
```

**Dependencies**:
- [ ] Comment tree structure support
- [ ] Notification ordering by time
- [ ] WebSocket integration

---

### 🟡 Phase 3: MEDIUM - Social Features (1 week)

**Target**: 5 endpoints
**Impact**: Improved social graph browsing
**Complexity**: Low-Medium

```
Week 5:
✅ GET /follows/user/:userId/followers
✅ GET /follows/user/:userId/following
✅ GET /follows/mutual
✅ GET /saved/posts
✅ GET /votes/user
```

---

### 🟡 Phase 4: OPTIONAL - Other Features (1 week)

**Target**: 4 endpoints
**Impact**: Nice to have
**Complexity**: Low

```
Week 6:
🟡 GET /posts/search
🟡 GET /media/user/:userId
🟡 GET /blocks
🟡 GET /search/history
```

---

## 4. แผนการทำงานทีละ Phase

### 📅 Phase 1: Posts & Feed (Week 1-2)

#### Week 1: Foundation & Database

##### Day 1-2: Database Setup
```bash
□ Create composite indexes
  □ idx_posts_feed_new (created_at DESC, id DESC)
  □ idx_posts_feed_top (votes DESC, created_at DESC, id DESC)
  □ idx_posts_feed_hot (created_at DESC, votes DESC, id DESC)
  □ idx_posts_by_tag (tag_id, created_at DESC, id DESC)
  □ idx_posts_by_author (author_id, created_at DESC, id DESC)

□ Test indexes with EXPLAIN ANALYZE
□ Deploy indexes to staging with CONCURRENTLY
□ Monitor index build progress
```

**SQL Script**:
```sql
-- migrations/YYYYMMDD_posts_cursor_indexes.up.sql
BEGIN;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_new
ON posts(is_deleted, status, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_top
ON posts(is_deleted, status, votes DESC, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ... etc
COMMIT;
```

##### Day 3: Cursor Utilities
```bash
□ Create pkg/utils/post_cursor.go
  □ type PostCursor struct
  □ EncodePostCursor(sortValue, createdAt, id)
  □ DecodePostCursor(cursorStr)

□ Write unit tests
  □ TestEncodeDecodePostCursor
  □ TestDecodeCursor_InvalidBase64
  □ TestDecodeCursor_EmptyString
  □ TestPostCursor_WithSortValue
  □ TestPostCursor_WithoutSortValue

□ Test coverage > 90%
```

**Code**:
```go
// pkg/utils/post_cursor.go
type PostCursor struct {
    SortValue *float64   `json:"sort_value,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
    ID        uuid.UUID  `json:"id"`
}

func EncodePostCursor(sortValue *float64, createdAt time.Time, id uuid.UUID) (string, error) {
    cursor := PostCursor{
        SortValue: sortValue,
        CreatedAt: createdAt,
        ID:        id,
    }

    jsonBytes, _ := json.Marshal(cursor)
    return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

func DecodePostCursor(cursorStr string) (*PostCursor, error) {
    if cursorStr == "" {
        return nil, nil
    }

    jsonBytes, err := base64.URLEncoding.DecodeString(cursorStr)
    if err != nil {
        return nil, err
    }

    var cursor PostCursor
    json.Unmarshal(jsonBytes, &cursor)
    return &cursor, nil
}
```

##### Day 4: Update DTOs
```bash
□ Update domain/dto/post.go
  □ Add PostListResponse.NextCursor *string
  □ Add PostListResponse.HasMore bool
  □ Remove PostListResponse.Total (expensive)
  □ Add PostResponse.HotScore *float64 (internal)

□ Update domain/dto/pagination.go (if exists)
  □ Create generic CursorResponse struct
```

**Code**:
```go
// domain/dto/post.go
type PostListResponse struct {
    Posts      []PostResponse `json:"posts"`
    NextCursor *string        `json:"nextCursor,omitempty"`
    HasMore    bool           `json:"hasMore"`
    // Total removed - too expensive to compute
}

type PostResponse struct {
    ID           uuid.UUID       `json:"id"`
    Title        string          `json:"title"`
    Content      string          `json:"content"`
    Author       UserResponse    `json:"author"`
    Votes        int             `json:"votes"`
    CommentCount int             `json:"commentCount"`
    UserVote     *int            `json:"userVote"`
    IsSaved      bool            `json:"isSaved"`
    CreatedAt    time.Time       `json:"createdAt"`

    // Internal use only (for cursor generation)
    HotScore     *float64        `json:"-"`
}
```

##### Day 5: Repository Layer
```bash
□ Update domain/repositories/post_repository.go
  □ Add ListWithCursor() method
  □ Add ListByTagWithCursor() method
  □ Add ListByAuthorWithCursor() method
  □ Add GetFeedWithCursor() method

□ Implement in infrastructure/persistence/post_repository.go
  □ Implement all cursor methods
  □ Add hot score calculation
  □ Handle composite cursors
  □ Write SQL queries with proper indexes
```

**Interface**:
```go
// domain/repositories/post_repository.go
type PostRepository interface {
    // Existing methods...

    // New cursor-based methods
    ListWithCursor(
        ctx context.Context,
        cursor *utils.PostCursor,
        limit int,
        sortBy PostSortBy,
    ) ([]*models.Post, error)

    ListByTagWithCursor(
        ctx context.Context,
        tagName string,
        cursor *utils.PostCursor,
        limit int,
        sortBy PostSortBy,
    ) ([]*models.Post, error)

    ListByAuthorWithCursor(
        ctx context.Context,
        authorID uuid.UUID,
        cursor *utils.PostCursor,
        limit int,
    ) ([]*models.Post, error)

    GetFeedWithCursor(
        ctx context.Context,
        userID uuid.UUID,
        cursor *utils.PostCursor,
        limit int,
        sortBy PostSortBy,
    ) ([]*models.Post, error)
}
```

**Implementation Example**:
```go
// infrastructure/persistence/post_repository.go

func (r *PostRepositoryImpl) ListWithCursor(
    ctx context.Context,
    cursor *utils.PostCursor,
    limit int,
    sortBy repositories.PostSortBy,
) ([]*models.Post, error) {
    query := r.db.WithContext(ctx).
        Where("is_deleted = ?", false).
        Where("status = ?", "published")

    switch sortBy {
    case repositories.SortByNew:
        query = r.applyCursorForNew(query, cursor)
        query = query.Order("created_at DESC, id DESC")

    case repositories.SortByTop:
        query = r.applyCursorForTop(query, cursor)
        query = query.Order("votes DESC, created_at DESC, id DESC")

    case repositories.SortByHot:
        query = r.applyCursorForHot(query, cursor)
        query = query.Select(`
            posts.*,
            votes / POWER(EXTRACT(EPOCH FROM (NOW() - created_at))/3600 + 2, 1.5) as hot_score
        `).
        Where("created_at > ?", time.Now().Add(-7*24*time.Hour)).
        Order("hot_score DESC, created_at DESC, id DESC")
    }

    var posts []*models.Post
    err := query.Limit(limit + 1).Find(&posts).Error

    return posts, err
}

func (r *PostRepositoryImpl) applyCursorForNew(query *gorm.DB, cursor *utils.PostCursor) *gorm.DB {
    if cursor == nil {
        return query
    }

    return query.Where(`
        (created_at < ? OR (created_at = ? AND id < ?))
    `, cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
}

func (r *PostRepositoryImpl) applyCursorForTop(query *gorm.DB, cursor *utils.PostCursor) *gorm.DB {
    if cursor == nil || cursor.SortValue == nil {
        return query
    }

    return query.Where(`
        (votes < ?
         OR (votes = ? AND created_at < ?)
         OR (votes = ? AND created_at = ? AND id < ?))
    `, int(*cursor.SortValue),
        int(*cursor.SortValue), cursor.CreatedAt,
        int(*cursor.SortValue), cursor.CreatedAt, cursor.ID)
}
```

#### Week 2: Service & Handler Layers

##### Day 6-7: Service Layer
```bash
□ Update domain/services/post_service.go interface
  □ Add ListPostsWithCursor() method
  □ Add GetFeedWithCursor() method
  □ Mark old methods as deprecated

□ Implement in application/serviceimpl/post_service_impl.go
  □ Implement cursor methods
  □ Generate nextCursor from last item
  □ Calculate hasMore (fetch limit+1 pattern)
  □ Enrich with user data (votes, saved status)
  □ Handle hot score calculation

□ Write unit tests with mocks
  □ Test first page (no cursor)
  □ Test second page (with cursor)
  □ Test last page (hasMore = false)
  □ Test empty results
  □ Test different sort types
```

**Service Implementation**:
```go
// application/serviceimpl/post_service_impl.go

func (s *PostServiceImpl) ListPostsWithCursor(
    ctx context.Context,
    cursor *utils.PostCursor,
    limit int,
    sortBy repositories.PostSortBy,
    userID *uuid.UUID,
) (*dto.PostListResponse, error) {
    // Validate limit
    if limit <= 0 || limit > 100 {
        limit = 20
    }

    // Fetch limit+1 to check hasMore
    posts, err := s.postRepo.ListWithCursor(ctx, cursor, limit+1, sortBy)
    if err != nil {
        return nil, err
    }

    // Check hasMore
    hasMore := len(posts) > limit
    if hasMore {
        posts = posts[:limit]
    }

    // Convert to DTOs
    postResponses := make([]dto.PostResponse, len(posts))
    for i, post := range posts {
        postResponses[i] = *s.toPostResponse(post, userID)
    }

    // Generate next cursor
    var nextCursor *string
    if hasMore && len(posts) > 0 {
        lastPost := posts[len(posts)-1]

        var sortValue *float64
        if sortBy == repositories.SortByTop {
            val := float64(lastPost.Votes)
            sortValue = &val
        } else if sortBy == repositories.SortByHot {
            // Calculate hot score
            hours := time.Since(lastPost.CreatedAt).Hours()
            score := float64(lastPost.Votes) / math.Pow(hours+2, 1.5)
            sortValue = &score
        }

        encoded, err := utils.EncodePostCursor(sortValue, lastPost.CreatedAt, lastPost.ID)
        if err == nil {
            nextCursor = &encoded
        }
    }

    return &dto.PostListResponse{
        Posts:      postResponses,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}
```

##### Day 8-9: Handler Layer
```bash
□ Update interfaces/api/handlers/post_handler.go
  □ Support both cursor and offset parameters
  □ Add cursor validation
  □ Handle decode errors gracefully
  □ Add deprecation warnings for offset
  □ Update all post endpoints

□ Endpoints to update:
  ✅ ListPosts()
  ✅ GetFeed()
  ✅ ListPostsByTag()
  ✅ ListPostsByTagID()
  ✅ ListPostsByAuthor()
  ✅ GetCrossposts()
```

**Handler Implementation**:
```go
// interfaces/api/handlers/post_handler.go

func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
    // Check if using cursor or offset
    cursorStr := c.Query("cursor")
    offsetStr := c.Query("offset")

    // Prefer cursor over offset
    if cursorStr != "" {
        return h.listPostsWithCursor(c, cursorStr)
    } else if offsetStr != "" {
        // Legacy offset-based (deprecated)
        return h.listPostsWithOffsetDeprecated(c, offsetStr)
    } else {
        // Default to cursor-based (first page)
        return h.listPostsWithCursor(c, "")
    }
}

func (h *PostHandler) listPostsWithCursor(c *fiber.Ctx, cursorStr string) error {
    // Parse cursor
    var cursor *utils.PostCursor
    var err error
    if cursorStr != "" {
        cursor, err = utils.DecodePostCursor(cursorStr)
        if err != nil {
            return utils.ValidationErrorResponse(c, "Invalid cursor")
        }
    }

    // Get parameters
    limitStr := c.Query("limit", "20")
    limit, _ := strconv.Atoi(limitStr)

    sortBy := c.Query("sort", "hot")
    var sortByEnum repositories.PostSortBy
    switch sortBy {
    case "hot":
        sortByEnum = repositories.SortByHot
    case "new":
        sortByEnum = repositories.SortByNew
    case "top":
        sortByEnum = repositories.SortByTop
    default:
        sortByEnum = repositories.SortByHot
    }

    // Get user ID if authenticated
    var userIDPtr *uuid.UUID
    if userID, ok := c.Locals("userID").(uuid.UUID); ok {
        userIDPtr = &userID
    }

    // Fetch posts
    response, err := h.postService.ListPostsWithCursor(
        c.Context(),
        cursor,
        limit,
        sortByEnum,
        userIDPtr,
    )
    if err != nil {
        return utils.ErrorResponse(c, apperrors.ErrInternal.WithInternal(err))
    }

    return utils.SuccessResponse(c, response, "Posts retrieved successfully")
}

func (h *PostHandler) listPostsWithOffsetDeprecated(c *fiber.Ctx, offsetStr string) error {
    // Legacy implementation with deprecation warning
    offset, _ := strconv.Atoi(offsetStr)
    limit, _ := strconv.Atoi(c.Query("limit", "20"))

    // ... existing offset logic ...

    // Add deprecation warning to response
    response := utils.SuccessResponse(c, posts, "Posts retrieved successfully")
    return c.Status(200).JSON(fiber.Map{
        "success": true,
        "message": "Posts retrieved successfully",
        "data":    posts,
        "meta": fiber.Map{
            "deprecated":          true,
            "deprecationMessage":  "Offset-based pagination is deprecated. Please use cursor-based pagination.",
            "migrateToEndpoint":   "/api/v1/posts?sort=hot&limit=20",
            "documentationUrl":    "https://docs.example.com/api/cursor-pagination",
        },
    })
}
```

##### Day 10: Testing
```bash
□ Write integration tests
  □ Test full pagination flow (3+ pages)
  □ Test no duplicates when new posts added
  □ Test all sort orders (hot, new, top)
  □ Test different tags
  □ Test by author
  □ Test feed (following)

□ Write load tests
  □ Benchmark cursor vs offset
  □ Test with 1M posts
  □ Test concurrent requests
  □ Memory profiling

□ Manual testing
  □ Test with Postman/Thunder Client
  □ Test frontend integration
  □ Test edge cases
```

**Integration Test Example**:
```go
// tests/integration/post_cursor_test.go

func TestPostFeed_CursorPagination_NoDuplicates(t *testing.T) {
    app := setupTestApp(t)
    user := createTestUser(t, app.DB)
    token := generateTestToken(user.ID)

    // Create 50 test posts
    for i := 0; i < 50; i++ {
        createTestPost(t, app.DB, user.ID, fmt.Sprintf("Post %d", i))
    }

    // Fetch page 1
    req1 := httptest.NewRequest("GET", "/api/v1/posts?sort=new&limit=20", nil)
    req1.Header.Set("Authorization", "Bearer "+token)
    resp1, _ := app.Test(req1)

    var page1 dto.PostListResponse
    json.NewDecoder(resp1.Body).Decode(&page1)

    assert.Len(t, page1.Posts, 20)
    assert.True(t, page1.HasMore)

    // Add 5 new posts while paginating
    for i := 50; i < 55; i++ {
        createTestPost(t, app.DB, user.ID, fmt.Sprintf("New Post %d", i))
    }

    // Fetch page 2 with cursor
    req2 := httptest.NewRequest("GET", "/api/v1/posts?sort=new&limit=20&cursor="+*page1.NextCursor, nil)
    req2.Header.Set("Authorization", "Bearer "+token)
    resp2, _ := app.Test(req2)

    var page2 dto.PostListResponse
    json.NewDecoder(resp2.Body).Decode(&page2)

    // Verify no duplicates
    page1IDs := getPostIDs(page1.Posts)
    page2IDs := getPostIDs(page2.Posts)

    for _, id := range page2IDs {
        assert.NotContains(t, page1IDs, id, "Found duplicate post ID")
    }
}
```

---

### 📅 Phase 2: Comments & Notifications (Week 3-4)

#### Week 3: Comments

```bash
Day 11-12: Repository Layer
□ Update domain/repositories/comment_repository.go
  □ Add ListByPostWithCursor()
  □ Add ListByAuthorWithCursor()
  □ Add ListRepliesWithCursor()

□ Implement cursor queries
  □ Handle nested comments
  □ Support different sort orders (hot, new, top, old)

Day 13: Service Layer
□ Update domain/services/comment_service.go
□ Implement service methods with cursor
□ Add comment tree support for cursor pagination

Day 14: Handler Layer & Testing
□ Update CommentHandler
  □ ListCommentsByPost
  □ ListCommentsByAuthor
  □ ListReplies

□ Write tests
□ Manual testing
```

**Comment Cursor Structure**:
```go
// Comments use similar cursor to posts
type CommentCursor struct {
    SortValue *float64   `json:"sort_value,omitempty"` // votes, hot_score
    CreatedAt time.Time  `json:"created_at"`
    ID        uuid.UUID  `json:"id"`
}
```

#### Week 4: Notifications

```bash
Day 15-16: Notifications
□ Update NotificationRepository
  □ Add GetNotificationsWithCursor()
  □ Add GetUnreadNotificationsWithCursor()

□ Update NotificationService
  □ Implement cursor pagination
  □ Sort by created_at DESC

□ Update NotificationHandler
  □ GetNotifications -> use cursor
  □ GetUnreadNotifications -> use cursor

Day 17-18: Testing & Optimization
□ Integration tests for comments
□ Integration tests for notifications
□ Performance testing
□ Fix any issues found
```

---

### 📅 Phase 3: Social Features (Week 5)

```bash
Week 5: Followers, Following, Saved Posts, Votes

Day 19-20: Follow System
□ Update FollowRepository
  □ GetFollowersWithCursor()
  □ GetFollowingWithCursor()
  □ GetMutualFollowsWithCursor()

□ Update handlers
□ Write tests

Day 21: Saved Posts & Votes
□ Update SavedPostRepository
  □ GetSavedPostsWithCursor()

□ Update VoteRepository
  □ GetUserVotesWithCursor()

□ Update handlers
□ Write tests

Day 22-23: Integration & Testing
□ Full integration testing
□ Cross-feature testing
□ Performance benchmarks
```

---

### 📅 Phase 4: Optional Features (Week 6)

```bash
Week 6: Search, Media, Blocks

Day 24-25: Search & Media
□ Update SearchRepository (if cursor makes sense)
□ Update MediaRepository
  □ GetUserMediaWithCursor()

Day 26-27: Polish & Documentation
□ Update API documentation
□ Write migration guide for frontend
□ Create code examples
□ Record demo videos

Day 28-30: Buffer & Deployment
□ Fix any remaining issues
□ Prepare for staging deployment
□ Code review
□ Final testing
```

---

## 5. Checklist แต่ละ Endpoint

### 📝 Template Checklist (ใช้สำหรับทุก endpoint)

```markdown
Endpoint: GET /api/v1/{endpoint}

Phase 1: Database
□ Create/verify composite index
□ Test index with EXPLAIN ANALYZE
□ Deploy index to staging
□ Monitor performance

Phase 2: Code Changes
□ Define cursor structure
□ Implement EncodeXXXCursor() if needed
□ Implement DecodeXXXCursor() if needed
□ Update repository interface
□ Implement repository method
□ Update service interface
□ Implement service method
□ Update handler to support cursor
□ Keep offset support with deprecation warning

Phase 3: Testing
□ Unit tests for cursor encode/decode
□ Unit tests for repository
□ Unit tests for service
□ Integration test - first page
□ Integration test - multiple pages
□ Integration test - no duplicates
□ Integration test - edge cases
□ Load test - performance benchmark

Phase 4: Documentation
□ Update Swagger/OpenAPI docs
□ Add code examples
□ Update API changelog
□ Add migration notes

Phase 5: Deployment
□ Deploy to staging
□ Smoke testing
□ Performance monitoring
□ Deploy to production (gradual rollout)
□ Monitor metrics
```

---

## 6. Timeline รวม

### 📊 Gantt Chart Overview

```
Week 1-2:  🔴 Posts & Feed (CRITICAL)
           ████████████████ (Foundation + Implementation)

Week 3-4:  🟡 Comments & Notifications
           ████████████████ (Medium Priority)

Week 5:    🟢 Social Features
           ████████ (Lower Priority)

Week 6:    🔵 Optional + Buffer
           ████████ (Nice to have + Cleanup)

───────────────────────────────────────────────────────
 W1    W2    W3    W4    W5    W6    W7    W8
```

### 📅 Detailed Schedule

| Week | Phase | Endpoints | Status | Deliverables |
|------|-------|-----------|--------|--------------|
| 1 | Foundation | - | 🏗️ Setup | Indexes, Utils, DTOs |
| 2 | Posts & Feed | 6 | 🔴 CRITICAL | Core feed working |
| 3 | Comments | 3 | 🟡 HIGH | Comment pagination |
| 4 | Notifications | 2 | 🟡 HIGH | Real-time notifs |
| 5 | Social | 5 | 🟢 MEDIUM | Follow/Save/Vote |
| 6 | Optional | 4 | 🔵 LOW | Search, Media |
| 7 | Testing | All | ✅ QA | Full testing |
| 8 | Deployment | All | 🚀 Release | Production |

---

## 7. Resource Requirements

### 👥 Team Composition

```
Minimum Team: 2-3 developers

Recommended Team:
- 1x Backend Lead (Senior) - Architecture & code review
- 2x Backend Developers - Implementation
- 1x Frontend Developer - API integration
- 1x QA Engineer - Testing
- 0.5x DevOps - Infrastructure & monitoring
```

### 🛠️ Technical Requirements

```
Development:
- PostgreSQL 14+ (for indexes)
- Go 1.21+
- GORM latest version
- Testing frameworks

Infrastructure:
- Staging environment (mirror production)
- Database migration tools
- Monitoring (Prometheus/Grafana)
- Load testing tools (k6, Apache Bench)

Frontend:
- React Query / SWR (for infinite scroll)
- Updated API client
```

### ⏱️ Time Estimates

```
Total Duration: 6-8 weeks

Breakdown:
- Planning & Design:       1 week (DONE ✅)
- Phase 1 (Critical):      2 weeks
- Phase 2 (High):          2 weeks
- Phase 3 (Medium):        1 week
- Phase 4 (Optional):      1 week
- Testing & Deployment:    1 week
```

### 💰 Cost Estimate (Optional)

```
Development Cost:
- 2 Backend Devs x 8 weeks x 40hrs = 640 hours
- 1 Frontend Dev x 4 weeks x 40hrs = 160 hours
- 1 QA x 2 weeks x 40hrs = 80 hours
Total: ~880 developer hours

Infrastructure:
- Staging environment: $XXX/month
- Monitoring tools: $XXX/month
- Database optimization: one-time cost
```

---

## 8. Risk & Mitigation

### ⚠️ Risk Analysis

#### Risk 1: Performance Degradation
**Probability**: 🟡 Medium
**Impact**: 🔴 High

**Description**: Composite indexes อาจทำให้ write operations ช้าลง

**Mitigation**:
- ✅ Test indexes ใน staging ก่อน
- ✅ Monitor write performance metrics
- ✅ ใช้ partial indexes ลด overhead
- ✅ Review query plans regularly
- ✅ ถ้าจำเป็น ใช้ materialized views

#### Risk 2: Data Inconsistency During Migration
**Probability**: 🟡 Medium
**Impact**: 🟡 Medium

**Description**: Users อาจเห็นข้อมูลไม่สอดคล้องระหว่าง offset และ cursor

**Mitigation**:
- ✅ Support ทั้ง offset และ cursor ชั่วคราว
- ✅ Gradual rollout (10% -> 50% -> 100%)
- ✅ Feature flag สำหรับ toggle
- ✅ Monitor user feedback
- ✅ Rollback plan พร้อม

#### Risk 3: Frontend Breaking Changes
**Probability**: 🟡 Medium
**Impact**: 🔴 High

**Description**: Frontend ที่ใช้ offset อาจ break

**Mitigation**:
- ✅ Maintain backward compatibility
- ✅ Deprecation warnings ใน API response
- ✅ ประสานงานกับ frontend team
- ✅ Provide migration examples
- ✅ Version API if needed (v2)

#### Risk 4: Cursor Encoding Bugs
**Probability**: 🟢 Low
**Impact**: 🟡 Medium

**Description**: Cursor encode/decode อาจมี bugs

**Mitigation**:
- ✅ Comprehensive unit tests (>90% coverage)
- ✅ Integration tests for pagination flow
- ✅ Fuzzing tests for cursor parsing
- ✅ Graceful error handling
- ✅ Logging for debug

#### Risk 5: Database Load Spike
**Probability**: 🟢 Low
**Impact**: 🔴 High

**Description**: Index creation อาจทำให้ database slow

**Mitigation**:
- ✅ ใช้ CREATE INDEX CONCURRENTLY
- ✅ Create indexes นอก peak hours
- ✅ Monitor database metrics
- ✅ Test ใน staging ก่อน
- ✅ มี rollback script พร้อม

#### Risk 6: Timeline Overrun
**Probability**: 🟡 Medium
**Impact**: 🟡 Medium

**Description**: อาจใช้เวลานานกว่าที่คาดไว้

**Mitigation**:
- ✅ Prioritize critical endpoints first
- ✅ Buffer time ใน timeline (Week 6-8)
- ✅ Daily standups to track progress
- ✅ ตัด optional features ถ้าจำเป็น
- ✅ Incremental delivery (phase by phase)

---

## 9. Success Metrics

### 📊 Key Performance Indicators (KPIs)

#### Performance Metrics
```
Target Performance (Cursor vs Offset):

Offset-Based (Current):
- Page 1:    ~1-2ms
- Page 100:  ~100-200ms
- Page 1000: ~1-2s

Cursor-Based (Target):
- Page 1:    ~1-2ms     (same)
- Page 100:  ~1-2ms     (100x faster ✅)
- Page 1000: ~1-2ms     (1000x faster ✅)

Success Criteria:
✅ Cursor pagination <= 5ms for any page
✅ No performance regression for page 1
✅ 95th percentile < 10ms
```

#### User Experience Metrics
```
Before (Offset):
- Duplicate posts reported: ~5-10% of users
- Missing posts reported: ~2-5% of users
- Slow scroll complaints: ~15% of users

After (Cursor):
- Duplicate posts: 0% target
- Missing posts: 0% target
- Slow scroll: <1% target
- Bounce rate: -10% improvement
```

#### Technical Metrics
```
Code Quality:
✅ Test coverage > 80%
✅ Zero critical bugs in production
✅ API response time < 100ms (p95)
✅ Error rate < 0.1%

Database:
✅ Index hit rate > 95%
✅ Query time < 5ms (p95)
✅ Write performance impact < 5%
```

---

## 10. Rollout Strategy

### 🚀 Gradual Rollout Plan

#### Stage 1: Internal Testing (Week 7)
```
Audience: Development team only
Duration: 3-5 days

Actions:
□ Deploy to staging
□ Internal QA testing
□ Fix critical bugs
□ Performance benchmarking
□ Load testing
```

#### Stage 2: Beta Testing (Week 7)
```
Audience: 10% of users (feature flag)
Duration: 3-4 days

Actions:
□ Deploy to production with feature flag
□ Enable for 10% random users
□ Monitor error rates
□ Monitor performance metrics
□ Collect user feedback
□ Fix any issues
```

#### Stage 3: Expand to 50% (Week 8)
```
Audience: 50% of users
Duration: 2-3 days

Actions:
□ Increase feature flag to 50%
□ Continue monitoring
□ Compare metrics offset vs cursor
□ Ensure stability
```

#### Stage 4: Full Rollout (Week 8)
```
Audience: 100% of users
Duration: Ongoing

Actions:
□ Enable cursor for all users
□ Keep offset as fallback (deprecated)
□ Announce migration
□ Update documentation
□ Monitor for 1 week
```

#### Stage 5: Deprecation (Week 10+)
```
After 2-4 weeks of stable cursor usage:

□ Add sunset date for offset endpoints
□ Send deprecation notices to clients
□ Remove offset support (if possible)
□ Clean up legacy code
```

---

## 11. Frontend Migration Guide

### 📱 React Integration Example

**Before (Offset)**:
```typescript
// ❌ Old way - offset pagination
const [page, setPage] = useState(1);
const limit = 20;

const { data } = useQuery(['posts', page], () =>
  axios.get(`/api/v1/posts?offset=${(page-1)*limit}&limit=${limit}`)
);

// User sees duplicates when new posts are added!
```

**After (Cursor)**:
```typescript
// ✅ New way - cursor pagination
import { useInfiniteQuery } from '@tanstack/react-query';

const {
  data,
  fetchNextPage,
  hasNextPage,
  isFetchingNextPage,
} = useInfiniteQuery({
  queryKey: ['posts', 'hot'],
  queryFn: ({ pageParam }) => {
    const params = new URLSearchParams({
      sort: 'hot',
      limit: '20',
    });
    if (pageParam) params.append('cursor', pageParam);

    return axios.get(`/api/v1/posts?${params}`);
  },
  getNextPageParam: (lastPage) =>
    lastPage.data.hasMore ? lastPage.data.nextCursor : undefined,
  initialPageParam: undefined,
});

// Smooth infinite scroll, no duplicates!
```

---

## 12. Monitoring & Alerting

### 📈 Metrics to Monitor

```
API Metrics:
- Endpoint response time (p50, p95, p99)
- Error rate per endpoint
- Request rate per endpoint
- Cursor decode errors

Database Metrics:
- Query execution time
- Index usage statistics
- Connection pool usage
- Slow query log

User Metrics:
- Scroll depth
- Bounce rate
- Time on page
- User complaints
```

### 🚨 Alerts to Setup

```
Critical Alerts:
- Error rate > 1% for cursor endpoints
- Response time p95 > 100ms
- Database connection pool > 90%
- Cursor decode error > 10/min

Warning Alerts:
- Response time p95 > 50ms
- Error rate > 0.5%
- Unusual traffic patterns
- Index not being used
```

---

## 13. Documentation Checklist

### 📚 Documentation Tasks

```
Internal Documentation:
□ Architecture decision record (ADR)
□ Database schema changes
□ Migration guide for developers
□ Testing guide
□ Deployment runbook

External Documentation:
□ API documentation (Swagger/OpenAPI)
□ Migration guide for API clients
□ Code examples
□ FAQ
□ Changelog

Training Materials:
□ Video tutorial
□ Demo application
□ Best practices guide
□ Troubleshooting guide
```

---

## สรุป 🎯

### ✅ Action Items

**Immediate (This Week)**:
1. ✅ Review และ approve แผนนี้
2. ✅ Set up project tracking (Jira/Linear/GitHub Projects)
3. ✅ Assign team members
4. ✅ Schedule kickoff meeting

**Week 1-2 (Critical Path)**:
1. 🔴 Create database indexes
2. 🔴 Implement cursor utilities
3. 🔴 Update Posts & Feed endpoints
4. 🔴 Frontend integration (React)

**Week 3-8 (Execution)**:
1. 🟡 Roll out phase by phase
2. 🟡 Testing continuously
3. 🟡 Monitor metrics
4. 🟡 Adjust based on feedback

### 📊 Expected Outcomes

**Performance**:
- ✅ 100-1000x faster pagination
- ✅ Consistent response times
- ✅ Better database utilization

**User Experience**:
- ✅ No duplicate posts
- ✅ Smooth infinite scroll
- ✅ Real-time feed updates
- ✅ Facebook-like experience

**Technical**:
- ✅ Scalable to millions of posts
- ✅ Modern API design
- ✅ Better code quality
- ✅ Easier to maintain

---

**Document Version**: 1.0
**Last Updated**: 2025-11-14
**Status**: Ready for Review
**Next Review**: Weekly during execution

---

**Questions or Concerns?**
Contact: Development Team Lead
