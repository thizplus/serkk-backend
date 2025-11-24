-- Rollback: ย้ายข้อมูลกลับจาก users_identity และ user_profiles ไปยัง users_cache
-- หมายเหตุ: migration นี้ไม่สามารถ rollback ได้อย่างสมบูรณ์
-- เพราะข้อมูลบางส่วนอาจหายไปแล้ว (deleted_at, role, is_active)

-- ถ้าต้องการ rollback ควร restore จาก backup แทน
SELECT 'Migration 027 cannot be fully rolled back. Please restore from backup.' as warning;
