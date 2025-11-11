package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get R2 configuration from environment
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	// Validate required variables
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		log.Fatal("❌ Missing required R2 environment variables. Please check your .env file.")
	}

	log.Println("📋 R2 Configuration:")
	log.Printf("   Account ID: %s", accountID)
	log.Printf("   Bucket Name: %s", bucketName)
	log.Printf("   Access Key ID: %s...", accessKeyID[:10])

	// Create R2 endpoint URL
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	log.Printf("   Endpoint: %s", endpoint)

	// Create AWS config with R2 endpoint
	cfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		),
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	log.Println("\n🔧 Configuring CORS for R2 bucket...")

	// Define CORS configuration
	corsConfig := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucketName),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: []types.CORSRule{
				{
					AllowedOrigins: []string{
						"http://localhost:3000",   // Development frontend
						"http://localhost:8080",   // Development backend
						"https://voobize.com",     // Production frontend
						"https://www.voobize.com", // Production frontend (www)
						"https://api.voobize.com", // Production backend
						"*",                       // Allow all origins (can be restricted later)
					},
					AllowedMethods: []string{
						"GET",
						"PUT",
						"POST",
						"DELETE",
						"HEAD",
					},
					AllowedHeaders: []string{
						"*",
					},
					ExposeHeaders: []string{
						"ETag",
						"Content-Length",
						"Content-Type",
					},
					MaxAgeSeconds: aws.Int32(3600), // Cache preflight for 1 hour
				},
			},
		},
	}

	// Apply CORS configuration
	ctx := context.Background()
	_, err := client.PutBucketCors(ctx, corsConfig)
	if err != nil {
		log.Fatalf("❌ Failed to set CORS configuration: %v", err)
	}

	log.Println("✅ CORS configured successfully!")

	// Verify CORS configuration
	log.Println("\n🔍 Verifying CORS configuration...")
	getCorsOutput, err := client.GetBucketCors(ctx, &s3.GetBucketCorsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Printf("⚠️  Warning: Failed to verify CORS: %v", err)
		return
	}

	log.Println("✅ CORS verification successful!")
	log.Println("\n📋 Current CORS Rules:")
	for i, rule := range getCorsOutput.CORSRules {
		log.Printf("\n   Rule #%d:", i+1)
		log.Printf("   - Allowed Origins: %v", rule.AllowedOrigins)
		log.Printf("   - Allowed Methods: %v", rule.AllowedMethods)
		log.Printf("   - Allowed Headers: %v", rule.AllowedHeaders)
		log.Printf("   - Expose Headers: %v", rule.ExposeHeaders)
		if rule.MaxAgeSeconds != nil {
			log.Printf("   - Max Age: %d seconds", *rule.MaxAgeSeconds)
		}
	}

	log.Println("\n🎉 R2 Bucket is ready for presigned uploads!")
}
