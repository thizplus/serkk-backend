@echo off
REM Auto-Post Setup Script for Windows
REM ช่วยตั้งค่า Auto-Post System แบบอัตโนมัติ

setlocal enabledelayedexpansion

echo ========================================
echo 🚀 Auto-Post Setup Script
echo ========================================
echo.

REM Check if .env exists
if not exist ".env" (
    echo ❌ Error: .env file not found
    echo Please create .env file first
    exit /b 1
)

echo Step 1: Checking configuration...
echo.

REM Check CSV file
set CSV_FILE=%1
if "%CSV_FILE%"=="" set CSV_FILE=suekk_720_posts.csv

if not exist "%CSV_FILE%" (
    echo ❌ Error: CSV file not found: %CSV_FILE%
    echo Usage: setup_auto_post.bat [csv_file]
    echo Example: setup_auto_post.bat suekk_720_posts.csv
    exit /b 1
)

echo ✅ Found CSV file: %CSV_FILE%

REM Check BOT_USER_ID in .env
findstr /C:"AUTO_POST_BOT_USER_ID=" .env >nul
if errorlevel 1 (
    echo ❌ Error: AUTO_POST_BOT_USER_ID not found in .env
    echo Please add to .env:
    echo AUTO_POST_BOT_USER_ID=your-uuid-here
    exit /b 1
)

REM Extract BOT_USER_ID
for /f "tokens=2 delims==" %%a in ('findstr /C:"AUTO_POST_BOT_USER_ID=" .env') do set BOT_USER_ID=%%a

if "%BOT_USER_ID%"=="" (
    echo ❌ Error: AUTO_POST_BOT_USER_ID is empty in .env
    exit /b 1
)

if "%BOT_USER_ID%"=="your-bot-user-uuid-here" (
    echo ❌ Error: AUTO_POST_BOT_USER_ID not configured in .env
    echo Please set your bot user ID in .env:
    echo AUTO_POST_BOT_USER_ID=your-uuid-here
    exit /b 1
)

echo ✅ Bot User ID configured: %BOT_USER_ID%

REM Check OpenAI API Key
findstr /C:"OPENAI_API_KEY=" .env >nul
if errorlevel 1 (
    echo ⚠️  Warning: OPENAI_API_KEY not found in .env
) else (
    for /f "tokens=2 delims==" %%a in ('findstr /C:"OPENAI_API_KEY=" .env') do set OPENAI_KEY=%%a
    if "!OPENAI_KEY!"=="" (
        echo ⚠️  Warning: OPENAI_API_KEY is empty
    ) else if "!OPENAI_KEY!"=="sk-your-openai-api-key-here" (
        echo ⚠️  Warning: OPENAI_API_KEY not configured
    ) else (
        echo ✅ OpenAI API Key configured
    )
)

echo.
echo ========================================
echo Step 2: Importing topics from CSV...
echo ========================================
echo.

REM Run import script
go run scripts\import_csv_to_db.go "%CSV_FILE%"

if errorlevel 1 (
    echo.
    echo ========================================
    echo ❌ Import failed!
    echo ========================================
    echo Please check the error messages above
    exit /b 1
)

echo.
echo ========================================
echo ✅ Import completed successfully!
echo ========================================
echo.
echo 📊 Summary:
echo   - CSV file imported
echo   - Settings created in database
echo   - Topics ready to use
echo.
echo ========================================
echo Next Steps:
echo ========================================
echo.
echo 1. Review imported settings:
echo    psql -U postgres -d gofiber_template -c "SELECT id, tone, jsonb_array_length(topics) as topics_count FROM auto_post_settings;"
echo.
echo 2. Test manual trigger (optional):
echo    curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/trigger ^
echo      -H "Authorization: Bearer YOUR_JWT_TOKEN"
echo.
echo 3. Enable settings:
echo    # Via SQL (enable all):
echo    psql -U postgres -d gofiber_template -c "UPDATE auto_post_settings SET is_enabled = true WHERE bot_user_id = '%BOT_USER_ID%';"
echo.
echo 4. Restart server to activate scheduler:
echo    # Docker:
echo    docker-compose restart app
echo.
echo    # Direct:
echo    bin\api.exe
echo.
echo ========================================
echo 🎉 Setup complete!
echo ========================================

endlocal
