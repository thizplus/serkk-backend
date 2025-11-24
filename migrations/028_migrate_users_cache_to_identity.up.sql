-- Migration 028: Migrate data from users_cache to users_identity and user_profiles
-- ย้ายข้อมูลจาก users_cache (ถ้ามี) ไปยัง users_identity และ user_profiles
-- Must run AFTER migration 027 (create_user_profiles)

-- 1. ย้าย identity data (id, email, username) ไปยัง users_identity
INSERT INTO users_identity (id, email, username, created_at, updated_at, synced_at)
SELECT
    id,
    email,
    username,
    created_at,
    updated_at,
    synced_at
FROM users_cache
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users_cache')
ON CONFLICT (id) DO NOTHING;

-- 2. ย้าง profile data (display_name, avatar) ไปยัง user_profiles
-- แต่ต้องเช็คว่าตาราง user_profiles มีอยู่หรือยัง
INSERT INTO user_profiles (user_id, display_name, avatar, created_at, updated_at)
SELECT
    id,
    COALESCE(display_name, username) as display_name, -- ถ้าไม่มี display_name ใช้ username
    COALESCE(avatar, '') as avatar,                   -- ถ้าไม่มี avatar ใช้ empty string
    created_at,
    updated_at
FROM users_cache
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users_cache')
ON CONFLICT (user_id) DO UPDATE SET
    display_name = COALESCE(EXCLUDED.display_name, user_profiles.display_name),
    avatar = COALESCE(EXCLUDED.avatar, user_profiles.avatar),
    updated_at = EXCLUDED.updated_at;

-- หมายเหตุ: ไม่ลบตาราง users_cache ในที่นี้
-- ตาราง users_cache จะถูกลบในขั้นตอนถัดไป (migration 029)
