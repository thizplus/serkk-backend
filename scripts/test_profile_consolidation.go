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
	baseURL := "http://localhost:8080"

	// Test user credentials (using an existing user from Auth Service)
	// You should replace these with actual test credentials
	loginPayload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}

	// Step 1: Login to get JWT token
	fmt.Println("📝 Step 1: Logging in...")
	loginBody, _ := json.Marshal(loginPayload)
	loginResp, err := http.Post(
		baseURL+"/api/v1/auth/login",
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

	// Step 2: Get current profile
	fmt.Println("\n📝 Step 2: Getting current profile...")
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
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
			fmt.Println("✅ Current profile:")
			fmt.Printf("   - Username: %v\n", profileData["username"])
			fmt.Printf("   - DisplayName: %v\n", profileData["displayName"])
			fmt.Printf("   - Avatar: %v\n", profileData["avatar"])
			fmt.Printf("   - Bio: %v\n", profileData["bio"])
			fmt.Printf("   - Location: %v\n", profileData["location"])
			fmt.Printf("   - Website: %v\n", profileData["website"])
		}
	}

	// Step 3: Update profile with ALL fields (including displayName and avatar)
	fmt.Println("\n📝 Step 3: Updating profile with ALL fields...")
	timestamp := time.Now().Format("15:04:05")
	updatePayload := map[string]string{
		"displayName": "Test Display Name " + timestamp,
		"avatar":      "https://example.com/avatar-" + timestamp + ".jpg",
		"bio":         "Updated bio at " + timestamp,
		"location":    "Bangkok, Thailand",
		"website":     "https://example.com",
	}

	updateBody, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest(
		"PATCH",
		baseURL+"/api/v1/users/profile",
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

	fmt.Println("✅ Profile update successful!")

	// Step 4: Verify the update
	fmt.Println("\n📝 Step 4: Verifying updated profile...")
	verifyReq, _ := http.NewRequest("GET", baseURL+"/api/v1/users/me", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+token)

	verifyResp, err := client.Do(verifyReq)
	if err != nil {
		fmt.Printf("❌ Verify failed: %v\n", err)
		return
	}
	defer verifyResp.Body.Close()

	if verifyResp.StatusCode == http.StatusOK {
		var verifyResult map[string]interface{}
		json.NewDecoder(verifyResp.Body).Decode(&verifyResult)

		if profileData, ok := verifyResult["data"].(map[string]interface{}); ok {
			fmt.Println("✅ Updated profile verified:")
			fmt.Printf("   - Username: %v\n", profileData["username"])
			fmt.Printf("   - DisplayName: %v (expected: %s)\n", profileData["displayName"], updatePayload["displayName"])
			fmt.Printf("   - Avatar: %v (expected: %s)\n", profileData["avatar"], updatePayload["avatar"])
			fmt.Printf("   - Bio: %v (expected: %s)\n", profileData["bio"], updatePayload["bio"])
			fmt.Printf("   - Location: %v (expected: %s)\n", profileData["location"], updatePayload["location"])
			fmt.Printf("   - Website: %v (expected: %s)\n", profileData["website"], updatePayload["website"])

			// Verify all fields match
			allMatch := true
			if profileData["displayName"] != updatePayload["displayName"] {
				fmt.Println("   ❌ DisplayName mismatch!")
				allMatch = false
			}
			if profileData["avatar"] != updatePayload["avatar"] {
				fmt.Println("   ❌ Avatar mismatch!")
				allMatch = false
			}
			if profileData["bio"] != updatePayload["bio"] {
				fmt.Println("   ❌ Bio mismatch!")
				allMatch = false
			}
			if profileData["location"] != updatePayload["location"] {
				fmt.Println("   ❌ Location mismatch!")
				allMatch = false
			}
			if profileData["website"] != updatePayload["website"] {
				fmt.Println("   ❌ Website mismatch!")
				allMatch = false
			}

			if allMatch {
				fmt.Println("\n🎉 SUCCESS! All fields updated correctly!")
				fmt.Println("✅ Profile consolidation working - Backend (8080) now handles ALL profile fields")
			}
		}
	}
}
