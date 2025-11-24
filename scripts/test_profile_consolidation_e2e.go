package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	backendURL := "http://localhost:8080"
	authURL := "http://localhost:8088"

	timestamp := time.Now().Format("150405") // Use only time for shorter username
	testUsername := "ctest" + timestamp      // Alphanumeric only: ctest150405 (12 chars)
	testEmail := "consolidation_test_" + timestamp + "@example.com"
	testPassword := "TestPassword123!"

	// Step 1: Register new user via Auth Service
	fmt.Println("📝 Step 1: Registering new user via Auth Service (8088)...")
	registerPayload := map[string]string{
		"username":    testUsername,
		"email":       testEmail,
		"password":    testPassword,
		"displayName": "Initial Display Name", // This goes to Auth Service initially
	}

	registerBody, _ := json.Marshal(registerPayload)
	registerResp, err := http.Post(
		authURL+"/api/v1/auth/register",
		"application/json",
		bytes.NewBuffer(registerBody),
	)
	if err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		return
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusCreated && registerResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(registerResp.Body)
		fmt.Printf("❌ Registration failed with status %d: %s\n", registerResp.StatusCode, string(body))
		return
	}

	fmt.Printf("✅ User registered: %s (%s)\n", testUsername, testEmail)

	// Wait for event to propagate
	fmt.Println("⏳ Waiting 2 seconds for event propagation...")
	time.Sleep(2 * time.Second)

	// Step 2: Login via Auth Service to get JWT token
	fmt.Println("\n📝 Step 2: Logging in via Auth Service (8088)...")
	loginPayload := map[string]string{
		"email":    testEmail,
		"password": testPassword,
	}

	loginBody, _ := json.Marshal(loginPayload)
	loginResp, err := http.Post(
		authURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(loginBody),
	)
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		return
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		fmt.Printf("❌ Login failed with status %d: %s\n", loginResp.StatusCode, string(body))
		return
	}

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)

	data, ok := loginResult["data"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ Invalid login response format")
		return
	}

	token, ok := data["token"].(string)
	if !ok {
		fmt.Println("❌ No token in login response")
		return
	}

	fmt.Println("✅ Login successful, token obtained")

	// Step 3: Get current profile (should have default values from event handler)
	fmt.Println("\n📝 Step 3: Getting current profile from Backend (8080)...")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", backendURL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	meResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Get profile failed: %v\n", err)
		return
	}
	defer meResp.Body.Close()

	if meResp.StatusCode == http.StatusOK {
		var meResult map[string]interface{}
		json.NewDecoder(meResp.Body).Decode(&meResult)

		if profileData, ok := meResult["data"].(map[string]interface{}); ok {
			fmt.Println("✅ Current profile (default from event handler):")
			fmt.Printf("   - Username: %v\n", profileData["username"])
			fmt.Printf("   - DisplayName: %v (should be username)\n", profileData["displayName"])
			fmt.Printf("   - Avatar: %v (should be empty)\n", profileData["avatar"])
			fmt.Printf("   - Bio: %v (should be empty)\n", profileData["bio"])
		}
	}

	// Step 4: Update profile with ALL fields via Backend Service ONLY
	fmt.Println("\n📝 Step 4: Updating ALL profile fields via Backend Service (8080) ONLY...")
	updateTimestamp := time.Now().Format("15:04:05")
	updatePayload := map[string]string{
		"displayName": "Updated Display Name " + updateTimestamp,
		"avatar":      "https://example.com/avatar-" + timestamp + ".jpg",
		"bio":         "This is my bio updated at " + updateTimestamp,
		"location":    "Bangkok, Thailand",
		"website":     "https://example.com/" + testUsername,
	}

	fmt.Println("   Updating fields:")
	fmt.Printf("   - displayName: %s ⭐ (NEW: now handled by Backend 8080)\n", updatePayload["displayName"])
	fmt.Printf("   - avatar: %s ⭐ (NEW: now handled by Backend 8080)\n", updatePayload["avatar"])
	fmt.Printf("   - bio: %s\n", updatePayload["bio"])
	fmt.Printf("   - location: %s\n", updatePayload["location"])
	fmt.Printf("   - website: %s\n", updatePayload["website"])

	updateBody, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest(
		"PATCH",
		backendURL+"/api/v1/users/profile",
		bytes.NewBuffer(updateBody),
	)
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := client.Do(updateReq)
	if err != nil {
		fmt.Printf("❌ Update failed: %v\n", err)
		return
	}
	defer updateResp.Body.Close()

	updateRespBody, _ := io.ReadAll(updateResp.Body)

	if updateResp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Update failed with status %d: %s\n", updateResp.StatusCode, string(updateRespBody))
		return
	}

	fmt.Println("✅ Profile update request successful!")

	// Step 5: Verify the update
	fmt.Println("\n📝 Step 5: Verifying updated profile...")
	verifyReq, _ := http.NewRequest("GET", backendURL+"/api/v1/users/me", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+token)

	verifyResp, err := client.Do(verifyReq)
	if err != nil {
		fmt.Printf("❌ Verify failed: %v\n", err)
		return
	}
	defer verifyResp.Body.Close()

	if verifyResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(verifyResp.Body)
		fmt.Printf("❌ Verify failed with status %d: %s\n", verifyResp.StatusCode, string(body))
		return
	}

	var verifyResult map[string]interface{}
	json.NewDecoder(verifyResp.Body).Decode(&verifyResult)

	if profileData, ok := verifyResult["data"].(map[string]interface{}); ok {
		fmt.Println("✅ Updated profile verified:")
		fmt.Printf("   - Username: %v\n", profileData["username"])
		fmt.Printf("   - DisplayName: %v\n", profileData["displayName"])
		fmt.Printf("   - Avatar: %v\n", profileData["avatar"])
		fmt.Printf("   - Bio: %v\n", profileData["bio"])
		fmt.Printf("   - Location: %v\n", profileData["location"])
		fmt.Printf("   - Website: %v\n", profileData["website"])

		// Verify all fields match
		allMatch := true
		checks := []struct {
			name     string
			expected string
			actual   interface{}
		}{
			{"DisplayName", updatePayload["displayName"], profileData["displayName"]},
			{"Avatar", updatePayload["avatar"], profileData["avatar"]},
			{"Bio", updatePayload["bio"], profileData["bio"]},
			{"Location", updatePayload["location"], profileData["location"]},
			{"Website", updatePayload["website"], profileData["website"]},
		}

		fmt.Println("\n🔍 Field Verification:")
		for _, check := range checks {
			if check.actual == check.expected {
				fmt.Printf("   ✅ %s: MATCH\n", check.name)
			} else {
				fmt.Printf("   ❌ %s: MISMATCH (expected: %s, got: %v)\n", check.name, check.expected, check.actual)
				allMatch = false
			}
		}

		if allMatch {
			fmt.Println("\n🎉 SUCCESS! Profile Consolidation Working Perfectly!")
			fmt.Println("=====================================")
			fmt.Println("✅ Backend Service (8080) now handles ALL profile fields:")
			fmt.Println("   • displayName ⭐ (previously Auth Service)")
			fmt.Println("   • avatar ⭐ (previously Auth Service)")
			fmt.Println("   • bio")
			fmt.Println("   • location")
			fmt.Println("   • website")
			fmt.Println("\n✅ Frontend can now call ONLY 8080 for profile updates!")
			fmt.Println("✅ No need to split requests between Auth Service (8088) and Backend (8080)")
		} else {
			fmt.Println("\n❌ FAILED: Some fields did not update correctly")
		}
	}
}
