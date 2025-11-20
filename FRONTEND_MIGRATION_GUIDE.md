# Frontend Migration Guide - Cursor-Based Pagination

## 📋 สรุปภาพรวม

Backend ได้อัพเกรดระบบ pagination จาก **offset-based** เป็น **cursor-based** เพื่อ:
- ✅ Performance ดีขึ้น 500-5000x ในการ scroll ลึกๆ
- ✅ ไม่มีข้อมูลซ้ำหรือหายขณะ scroll
- ✅ เหมาะกับ infinite scroll UI
- ✅ รองรับ real-time updates ได้ดีกว่า

---

## 🎯 Status การ Migrate

### Phase 1 - **พร้อมใช้งานแล้ว (ควร migrate ทันที)**
- ✅ GET `/api/v1/posts` - List all posts (hot/new/top)
- ✅ GET `/api/v1/posts/author/:authorId` - Posts by author
- ✅ GET `/api/v1/posts/tag/:tagName` - Posts by tag
- ✅ GET `/api/v1/posts/feed` - Following feed

### Phase 2 - **ยังไม่พร้อม (อยู่ระหว่างพัฒนา)**
- ⏳ Comments endpoints
- ⏳ Notifications endpoints

### Phase 3 - **ยังไม่พร้อม (อยู่ระหว่างพัฒนา)**
- ⏳ Followers/Following endpoints
- ⏳ Saved posts endpoints

---

## 🔄 การเปลี่ยนแปลง API

### เดิม: Offset-Based Pagination
```typescript
// Request
GET /api/v1/posts?offset=0&limit=20&sort=hot

// Response
{
  "data": {
    "posts": [...],
    "meta": {
      "total": 1000,
      "offset": 0,
      "limit": 20
    }
  }
}
```

### ใหม่: Cursor-Based Pagination
```typescript
// Request (First page)
GET /api/v1/posts?limit=20&sort=hot

// Response
{
  "data": {
    "posts": [...],
    "meta": {
      "nextCursor": "eyJjcmVhdGVkX2F0IjoiMjAyNS0wMS0xNFQxMDowMDowMFoiLCJpZCI6IjEyMzQ1Njc4LTEyMzQtMTIzNC0xMjM0LTEyMzQ1Njc4OTBhYiJ9",
      "hasMore": true,
      "limit": 20
    }
  }
}

// Request (Next page)
GET /api/v1/posts?cursor=eyJjcmVhdGVkX2F0...&limit=20&sort=hot

// Response
{
  "data": {
    "posts": [...],
    "meta": {
      "nextCursor": "eyJjcmVhdGVkX2F0IjoiMjAyNS0wMS0xNFQwOTowMDowMFoiLCJpZCI6Ijk4NzY1NDMyLTEyMzQtMTIzNC0xMjM0LTEyMzQ1Njc4OTBhYiJ9",
      "hasMore": true,
      "limit": 20
    }
  }
}

// Request (Last page)
GET /api/v1/posts?cursor=eyJjcmVhdGVkX2F0...&limit=20&sort=hot

// Response
{
  "data": {
    "posts": [...],
    "meta": {
      "nextCursor": null,
      "hasMore": false,
      "limit": 20
    }
  }
}
```

---

## 📝 Request Parameters

### ใหม่: Cursor-Based (แนะนำ)
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `cursor` | string | No | Cursor token สำหรับ page ถัดไป | `eyJjcmVhdGVkX2F0...` |
| `limit` | number | No | จำนวนรายการต่อหน้า (default: 20, max: 100) | `20` |
| `sort` | string | No | รูปแบบการเรียง: `new`, `top`, `hot` (default: `hot`) | `hot` |
| `tag` | string | No | Filter by tag (for /posts endpoint only) | `javascript` |

### เดิม: Offset-Based (ยังใช้ได้ แต่ deprecated)
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `offset` | number | No | จำนวนรายการที่ข้าม (default: 0) | `0` |
| `limit` | number | No | จำนวนรายการต่อหน้า (default: 20, max: 100) | `20` |
| `sort` | string | No | รูปแบบการเรียง | `hot` |

---

## 📊 Response Format

### CursorPaginationMeta
```typescript
interface CursorPaginationMeta {
  nextCursor: string | null;  // Cursor สำหรับหน้าถัดไป (null ถ้าหมดแล้ว)
  hasMore: boolean;            // มีข้อมูลอีกหรือไม่
  limit: number;               // Limit ที่ใช้ในการ query
}
```

### PostListCursorResponse
```typescript
interface PostListCursorResponse {
  posts: Post[];
  meta: CursorPaginationMeta;
}
```

---

## 🔌 Endpoints ที่เปลี่ยนแปลง

### 1. GET `/api/v1/posts` - List All Posts

#### เดิม (Offset-Based - Deprecated)
```bash
GET /api/v1/posts?offset=0&limit=20&sort=hot
```

#### ใหม่ (Cursor-Based - Recommended)
```bash
# First page
GET /api/v1/posts?limit=20&sort=hot

# Next page
GET /api/v1/posts?cursor=eyJjcmVhdGVkX2F0...&limit=20&sort=hot
```

**Response:**
```json
{
  "success": true,
  "message": "Posts retrieved successfully",
  "data": {
    "posts": [
      {
        "id": "12345678-1234-1234-1234-123456789abc",
        "title": "Sample Post",
        "content": "Post content...",
        "author": {
          "id": "87654321-1234-1234-1234-123456789abc",
          "username": "john_doe",
          "displayName": "John Doe"
        },
        "votes": 42,
        "commentsCount": 5,
        "createdAt": "2025-01-14T10:00:00Z",
        "isLiked": false,
        "isSaved": false
      }
    ],
    "meta": {
      "nextCursor": "eyJjcmVhdGVkX2F0IjoiMjAyNS0wMS0xNFQwOTowMDowMFoiLCJpZCI6Ijk4NzY1NDMyIn0=",
      "hasMore": true,
      "limit": 20
    }
  }
}
```

---

### 2. GET `/api/v1/posts/author/:authorId` - Posts by Author

#### เดิม (Offset-Based)
```bash
GET /api/v1/posts/author/12345678-1234-1234-1234-123456789abc?offset=0&limit=20
```

#### ใหม่ (Cursor-Based)
```bash
# First page
GET /api/v1/posts/author/12345678-1234-1234-1234-123456789abc?limit=20

# Next page
GET /api/v1/posts/author/12345678-1234-1234-1234-123456789abc?cursor=eyJjcmVhdGVkX2F0...&limit=20
```

**Response Format:** เหมือน endpoint `/posts`

---

### 3. GET `/api/v1/posts/tag/:tagName` - Posts by Tag

#### เดิม (Offset-Based)
```bash
GET /api/v1/posts/tag/javascript?offset=0&limit=20&sort=new
```

#### ใหม่ (Cursor-Based)
```bash
# First page
GET /api/v1/posts/tag/javascript?limit=20&sort=new

# Next page
GET /api/v1/posts/tag/javascript?cursor=eyJjcmVhdGVkX2F0...&limit=20&sort=new
```

**Response Format:** เหมือน endpoint `/posts`

---

### 4. GET `/api/v1/posts/feed` - Following Feed

#### เดิม (Offset-Based)
```bash
GET /api/v1/posts/feed?offset=0&limit=20
```

#### ใหม่ (Cursor-Based)
```bash
# First page
GET /api/v1/posts/feed?limit=20

# Next page
GET /api/v1/posts/feed?cursor=eyJjcmVhdGVkX2F0...&limit=20
```

**Response Format:** เหมือน endpoint `/posts`

**Note:** Endpoint นี้ต้อง authentication (Bearer token)

---

## 💻 Frontend Implementation Examples

### React + TypeScript Example

#### 1. Type Definitions
```typescript
// types/pagination.ts
export interface CursorPaginationMeta {
  nextCursor: string | null;
  hasMore: boolean;
  limit: number;
}

export interface Post {
  id: string;
  title: string;
  content: string;
  author: {
    id: string;
    username: string;
    displayName: string;
  };
  votes: number;
  commentsCount: number;
  createdAt: string;
  isLiked: boolean;
  isSaved: boolean;
}

export interface PostListCursorResponse {
  posts: Post[];
  meta: CursorPaginationMeta;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}
```

#### 2. API Service
```typescript
// services/postService.ts
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export const postService = {
  // Cursor-based pagination (recommended)
  getPostsCursor: async (
    cursor?: string,
    limit: number = 20,
    sort: 'hot' | 'new' | 'top' = 'hot'
  ): Promise<PostListCursorResponse> => {
    const params = new URLSearchParams();
    if (cursor) params.append('cursor', cursor);
    params.append('limit', limit.toString());
    params.append('sort', sort);

    const response = await axios.get<ApiResponse<PostListCursorResponse>>(
      `${API_BASE_URL}/posts?${params}`
    );
    return response.data.data;
  },

  getPostsByAuthorCursor: async (
    authorId: string,
    cursor?: string,
    limit: number = 20
  ): Promise<PostListCursorResponse> => {
    const params = new URLSearchParams();
    if (cursor) params.append('cursor', cursor);
    params.append('limit', limit.toString());

    const response = await axios.get<ApiResponse<PostListCursorResponse>>(
      `${API_BASE_URL}/posts/author/${authorId}?${params}`
    );
    return response.data.data;
  },

  getPostsByTagCursor: async (
    tagName: string,
    cursor?: string,
    limit: number = 20,
    sort: 'hot' | 'new' | 'top' = 'new'
  ): Promise<PostListCursorResponse> => {
    const params = new URLSearchParams();
    if (cursor) params.append('cursor', cursor);
    params.append('limit', limit.toString());
    params.append('sort', sort);

    const response = await axios.get<ApiResponse<PostListCursorResponse>>(
      `${API_BASE_URL}/posts/tag/${tagName}?${params}`
    );
    return response.data.data;
  },

  getFeedCursor: async (
    cursor?: string,
    limit: number = 20,
    token: string
  ): Promise<PostListCursorResponse> => {
    const params = new URLSearchParams();
    if (cursor) params.append('cursor', cursor);
    params.append('limit', limit.toString());

    const response = await axios.get<ApiResponse<PostListCursorResponse>>(
      `${API_BASE_URL}/posts/feed?${params}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    return response.data.data;
  },
};
```

#### 3. React Hook - Infinite Scroll
```typescript
// hooks/usePostsCursor.ts
import { useState, useEffect } from 'react';
import { postService } from '../services/postService';
import { Post } from '../types/pagination';

export const usePostsCursor = (
  sort: 'hot' | 'new' | 'top' = 'hot',
  limit: number = 20
) => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load first page
  const loadInitial = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await postService.getPostsCursor(undefined, limit, sort);
      setPosts(response.posts);
      setNextCursor(response.meta.nextCursor);
      setHasMore(response.meta.hasMore);
    } catch (err) {
      setError('Failed to load posts');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Load next page
  const loadMore = async () => {
    if (!hasMore || loading || !nextCursor) return;

    try {
      setLoading(true);
      setError(null);
      const response = await postService.getPostsCursor(nextCursor, limit, sort);
      setPosts((prev) => [...prev, ...response.posts]);
      setNextCursor(response.meta.nextCursor);
      setHasMore(response.meta.hasMore);
    } catch (err) {
      setError('Failed to load more posts');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Reset when sort changes
  useEffect(() => {
    setPosts([]);
    setNextCursor(null);
    setHasMore(true);
    loadInitial();
  }, [sort]);

  return {
    posts,
    hasMore,
    loading,
    error,
    loadMore,
    refresh: loadInitial,
  };
};
```

#### 4. React Component - Infinite Scroll
```typescript
// components/PostList.tsx
import React, { useEffect, useRef } from 'react';
import { usePostsCursor } from '../hooks/usePostsCursor';
import PostCard from './PostCard';

interface PostListProps {
  sort?: 'hot' | 'new' | 'top';
}

const PostList: React.FC<PostListProps> = ({ sort = 'hot' }) => {
  const { posts, hasMore, loading, error, loadMore } = usePostsCursor(sort, 20);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Intersection Observer for infinite scroll
  useEffect(() => {
    if (loading) return;

    if (observerRef.current) {
      observerRef.current.disconnect();
    }

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore) {
          loadMore();
        }
      },
      { threshold: 0.1 }
    );

    if (loadMoreRef.current) {
      observerRef.current.observe(loadMoreRef.current);
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [loading, hasMore, loadMore]);

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="post-list">
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}

      {/* Loading indicator */}
      {loading && (
        <div className="loading">Loading more posts...</div>
      )}

      {/* Infinite scroll trigger */}
      {hasMore && !loading && (
        <div ref={loadMoreRef} className="load-more-trigger" />
      )}

      {/* End of list message */}
      {!hasMore && posts.length > 0 && (
        <div className="end-message">No more posts</div>
      )}
    </div>
  );
};

export default PostList;
```

---

### React Query Example

```typescript
// hooks/useInfinitePosts.ts
import { useInfiniteQuery } from '@tanstack/react-query';
import { postService } from '../services/postService';

export const useInfinitePosts = (sort: 'hot' | 'new' | 'top' = 'hot') => {
  return useInfiniteQuery({
    queryKey: ['posts', sort],
    queryFn: ({ pageParam }) =>
      postService.getPostsCursor(pageParam, 20, sort),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.meta.hasMore ? lastPage.meta.nextCursor : undefined,
  });
};

// Usage in component
const PostList = () => {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfinitePosts('hot');

  const posts = data?.pages.flatMap((page) => page.posts) ?? [];

  return (
    <div>
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}

      {hasNextPage && (
        <button
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? 'Loading...' : 'Load More'}
        </button>
      )}
    </div>
  );
};
```

---

### Vue 3 + Composition API Example

```typescript
// composables/usePostsCursor.ts
import { ref, watch } from 'vue';
import { postService } from '../services/postService';
import type { Post } from '../types/pagination';

export const usePostsCursor = (
  sort: Ref<'hot' | 'new' | 'top'>,
  limit: number = 20
) => {
  const posts = ref<Post[]>([]);
  const nextCursor = ref<string | null>(null);
  const hasMore = ref(true);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const loadInitial = async () => {
    try {
      loading.value = true;
      error.value = null;
      const response = await postService.getPostsCursor(undefined, limit, sort.value);
      posts.value = response.posts;
      nextCursor.value = response.meta.nextCursor;
      hasMore.value = response.meta.hasMore;
    } catch (err) {
      error.value = 'Failed to load posts';
      console.error(err);
    } finally {
      loading.value = false;
    }
  };

  const loadMore = async () => {
    if (!hasMore.value || loading.value || !nextCursor.value) return;

    try {
      loading.value = true;
      error.value = null;
      const response = await postService.getPostsCursor(
        nextCursor.value,
        limit,
        sort.value
      );
      posts.value.push(...response.posts);
      nextCursor.value = response.meta.nextCursor;
      hasMore.value = response.meta.hasMore;
    } catch (err) {
      error.value = 'Failed to load more posts';
      console.error(err);
    } finally {
      loading.value = false;
    }
  };

  watch(sort, () => {
    posts.value = [];
    nextCursor.value = null;
    hasMore.value = true;
    loadInitial();
  });

  return {
    posts,
    hasMore,
    loading,
    error,
    loadMore,
    refresh: loadInitial,
  };
};
```

---

## 🔧 Migration Strategy

### แนวทางการ Migrate (แนะนำ)

#### Step 1: เพิ่ม Feature Flag
```typescript
// config.ts
export const FEATURES = {
  USE_CURSOR_PAGINATION: true, // Toggle ระหว่างเดิมกับใหม่
};
```

#### Step 2: รองรับทั้ง 2 แบบ
```typescript
// services/postService.ts
export const postService = {
  getPosts: async (params: GetPostsParams) => {
    if (FEATURES.USE_CURSOR_PAGINATION) {
      return getPostsCursor(params.cursor, params.limit, params.sort);
    } else {
      return getPostsOffset(params.offset, params.limit, params.sort);
    }
  },
};
```

#### Step 3: ทดสอบ Cursor-based
```typescript
// เปิด feature flag
FEATURES.USE_CURSOR_PAGINATION = true;

// ทดสอบ
// 1. Scroll ลงไปหลายๆ หน้า → ต้องไม่มีข้อมูลซ้ำ
// 2. Refresh กลางคัน → ข้อมูลต้องไม่หาย
// 3. เปลี่ยน sort → ต้อง reset cursor ใหม่
// 4. Performance → ควรโหลดเร็วขึ้น
```

#### Step 4: ลบ Offset-based (ถ้าพร้อม)
```typescript
// ลบ code เก่าออก เมื่อ cursor-based ใช้งานได้ดีแล้ว
```

---

## ⚠️ สิ่งที่ต้องระวัง

### 1. Cursor ไม่สามารถใช้ข้ามกัน
```typescript
// ❌ ผิด - ใช้ cursor จาก "hot" กับ "new"
const hotResponse = await getPostsCursor(undefined, 20, 'hot');
const newResponse = await getPostsCursor(hotResponse.meta.nextCursor, 20, 'new');
// จะได้ผลลัพธ์ผิดพลาด!

// ✅ ถูก - Reset cursor เมื่อเปลี่ยน sort
useEffect(() => {
  setPosts([]);
  setCursor(null);
  loadInitial();
}, [sort]);
```

### 2. Cursor เปลี่ยนตลอด (ไม่ควรเก็บถาวร)
```typescript
// ❌ ผิด - เก็บ cursor ใน localStorage
localStorage.setItem('lastCursor', cursor);

// ✅ ถูก - เก็บแค่ใน state ชั่วคราว
const [cursor, setCursor] = useState<string | null>(null);
```

### 3. ไม่สามารถ jump ไปหน้าไหนก็ได้
```typescript
// ❌ ไม่มี - ไม่สามารถทำแบบนี้ได้
<Pagination currentPage={5} totalPages={100} />

// ✅ แนะนำ - ใช้ infinite scroll แทน
<InfiniteScroll loadMore={loadMore} hasMore={hasMore} />
```

### 4. Backend ยัง Support Offset-based (Backward Compatible)
```typescript
// ✅ ยังใช้ได้ถ้า frontend ยังไม่พร้อม
GET /api/v1/posts?offset=20&limit=20&sort=hot

// แต่จะมี warning log ใน backend
// "⚠️  Using deprecated offset-based pagination"
```

---

## 🎨 UI/UX Recommendations

### 1. Infinite Scroll (แนะนำ)
```typescript
// Best for: News feed, social media, timeline
// Pros: ราบรื่น, ไม่ต้องคิด pagination
// Cons: ยากต่อการกลับไปหาข้อมูลเดิม

<InfiniteScrollContainer
  dataLength={posts.length}
  next={loadMore}
  hasMore={hasMore}
  loader={<Spinner />}
  endMessage={<p>No more posts</p>}
>
  {posts.map(post => <PostCard key={post.id} post={post} />)}
</InfiniteScrollContainer>
```

### 2. Load More Button
```typescript
// Best for: Search results, product listing
// Pros: User control, predictable
// Cons: ต้องกดปุ่ม

{posts.map(post => <PostCard key={post.id} post={post} />)}
{hasMore && (
  <button onClick={loadMore} disabled={loading}>
    {loading ? 'Loading...' : 'Load More'}
  </button>
)}
```

### 3. Hybrid Approach
```typescript
// Infinite scroll + Load more fallback
// Auto-load หน้าแรก 2-3 หน้า แล้วให้กดปุ่ม

const [autoLoadCount, setAutoLoadCount] = useState(0);
const MAX_AUTO_LOAD = 2;

const handleScroll = () => {
  if (autoLoadCount < MAX_AUTO_LOAD) {
    loadMore();
    setAutoLoadCount(prev => prev + 1);
  }
};
```

---

## 🧪 Testing Checklist

### Frontend Testing
- [ ] Load first page → แสดงข้อมูลถูกต้อง
- [ ] Load more → ข้อมูลไม่ซ้ำกัน
- [ ] Scroll ลึกๆ (10+ pages) → Performance ดี
- [ ] เปลี่ยน sort/filter → Reset cursor และโหลดใหม่
- [ ] Refresh หน้า → กลับไปเริ่มต้นได้
- [ ] Network error → แสดง error message
- [ ] Empty state → แสดงข้อความที่เหมาะสม
- [ ] End of list → แสดงว่าหมดข้อมูลแล้ว

### Integration Testing
- [ ] ทดสอบกับ Backend จริง
- [ ] ทดสอบกับข้อมูลจริง (large dataset)
- [ ] ทดสอบใน production-like environment

---

## 📞 Support & Questions

### ถ้าพบปัญหา:
1. ตรวจสอบ Network tab ว่า request/response ถูกต้องหรือไม่
2. ตรวจสอบว่า `cursor` ถูก encode/decode ถูกต้อง
3. ตรวจสอบ Backend logs สำหรับ errors

### ถ้าต้องการความช่วยเหลือ:
- Backend Team: ตรวจสอบ Backend logs
- Frontend Team: ตรวจสอบ Network requests และ State management

---

## 📚 Additional Resources

### Documentation
- `CURSOR_DEPLOYMENT_GUIDE.md` - Backend deployment guide
- `CURSOR_MIGRATION_COMPLETE_SUMMARY.md` - Complete technical details
- API Swagger Docs: `http://localhost:8080/swagger/index.html`

### Example Code
- ตัวอย่าง React: ดูในเอกสารนี้
- ตัวอย่าง Vue: ดูในเอกสารนี้
- ตัวอย่าง React Query: ดูในเอกสารนี้

---

## 🎯 Quick Start Checklist

- [ ] อ่านเอกสารนี้ทั้งหมด
- [ ] Update type definitions
- [ ] เพิ่ม cursor support ใน API service
- [ ] สร้าง custom hook สำหรับ cursor pagination
- [ ] Update UI components ให้รองรับ infinite scroll
- [ ] ทดสอบใน development environment
- [ ] ทดสอบใน staging environment
- [ ] Deploy to production
- [ ] Monitor metrics และ user feedback

---

**Last Updated:** 2025-01-14
**API Version:** v1.0.0
**Status:** Phase 1 Ready for Frontend Integration

**ติดต่อ Backend Team:** หากพบปัญหาหรือมีคำถามเพิ่มเติม
