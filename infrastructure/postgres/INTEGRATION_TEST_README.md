# Integration Tests

Integration tests ต้องการ PostgreSQL database จริงเพื่อทดสอบ

## Setup Test Database

1. สร้าง test database:
```sql
CREATE DATABASE gofiber_test;
```

2. กำหนด environment variables (ถ้าต้องการ override default):
```bash
export TEST_DB_HOST=localhost
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=gofiber_test
export TEST_DB_PORT=5432
```

3. หรือแก้ไขค่าใน `pkg/config/test_config.go` ตามการตั้งค่า database ของคุณ

## วิธีรัน Integration Tests

รัน integration tests ทั้งหมด:
```bash
go test ./infrastructure/postgres/ -tags=integration -v
```

รันเฉพาะ UserRepository tests:
```bash
go test ./infrastructure/postgres/ -tags=integration -run TestUserRepository -v
```

รันเฉพาะ PostRepository tests:
```bash
go test ./infrastructure/postgres/ -tags=integration -run TestPostRepository -v
```

## Test Coverage

### UserRepository (13 tests)
- Create, GetByID, GetByEmail, GetByUsername
- Update, Delete
- List, Count
- Duplicate email/username handling

### PostRepository (13 tests)
- Create, GetByID, Update, Delete
- List, ListByAuthor
- Count, CountByAuthor
- UpdateVoteCount
- Search

## หมายเหตุ

- Integration tests จะทำการ cleanup database หลังแต่ละ test
- ใช้ `+build integration` tag เพื่อแยก integration tests จาก unit tests
- ไม่ควรรัน integration tests ใน production database
