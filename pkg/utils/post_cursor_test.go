package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodePostCursor(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	sortValue := 19.5

	// Test encoding with sort value
	encoded, err := EncodePostCursor(&sortValue, now, id)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Test decoding
	decoded, err := DecodePostCursor(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.NotNil(t, decoded.SortValue)
	assert.Equal(t, sortValue, *decoded.SortValue)
	assert.Equal(t, now.Unix(), decoded.CreatedAt.Unix()) // Compare Unix timestamps
	assert.Equal(t, id, decoded.ID)
}

func TestEncodeDecodePostCursor_WithoutSortValue(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	// Test encoding without sort value (for "new" sorting)
	encoded, err := EncodePostCursor(nil, now, id)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Test decoding
	decoded, err := DecodePostCursor(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Nil(t, decoded.SortValue)
	assert.Equal(t, now.Unix(), decoded.CreatedAt.Unix())
	assert.Equal(t, id, decoded.ID)
}

func TestEncodePostCursorSimple(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	// Test simple encoding
	encoded, err := EncodePostCursorSimple(now, id)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode and verify
	decoded, err := DecodePostCursor(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Nil(t, decoded.SortValue)
	assert.Equal(t, now.Unix(), decoded.CreatedAt.Unix())
	assert.Equal(t, id, decoded.ID)
}

func TestDecodePostCursor_EmptyString(t *testing.T) {
	// Empty cursor should return nil (first page)
	decoded, err := DecodePostCursor("")
	assert.NoError(t, err)
	assert.Nil(t, decoded)
}

func TestDecodePostCursor_InvalidBase64(t *testing.T) {
	// Invalid base64 should return error
	decoded, err := DecodePostCursor("invalid-base64!!!")
	assert.Error(t, err)
	assert.Nil(t, decoded)
}

func TestDecodePostCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	invalidJSON := "aGVsbG8gd29ybGQ=" // "hello world" in base64
	decoded, err := DecodePostCursor(invalidJSON)
	assert.Error(t, err)
	assert.Nil(t, decoded)
}

func TestPostCursor_RoundTrip(t *testing.T) {
	testCases := []struct {
		name      string
		sortValue *float64
		createdAt time.Time
		id        uuid.UUID
	}{
		{
			name:      "with sort value (top sorting)",
			sortValue: floatPtr(100.5),
			createdAt: time.Now(),
			id:        uuid.New(),
		},
		{
			name:      "without sort value (new sorting)",
			sortValue: nil,
			createdAt: time.Now(),
			id:        uuid.New(),
		},
		{
			name:      "with zero sort value",
			sortValue: floatPtr(0.0),
			createdAt: time.Now(),
			id:        uuid.New(),
		},
		{
			name:      "with negative sort value",
			sortValue: floatPtr(-5.5),
			createdAt: time.Now(),
			id:        uuid.New(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encode
			encoded, err := EncodePostCursor(tc.sortValue, tc.createdAt, tc.id)
			assert.NoError(t, err)
			assert.NotEmpty(t, encoded)

			// Decode
			decoded, err := DecodePostCursor(encoded)
			assert.NoError(t, err)
			assert.NotNil(t, decoded)

			// Verify
			if tc.sortValue != nil {
				assert.NotNil(t, decoded.SortValue)
				assert.Equal(t, *tc.sortValue, *decoded.SortValue)
			} else {
				assert.Nil(t, decoded.SortValue)
			}
			assert.Equal(t, tc.createdAt.Unix(), decoded.CreatedAt.Unix())
			assert.Equal(t, tc.id, decoded.ID)
		})
	}
}

// Helper function
func floatPtr(f float64) *float64 {
	return &f
}
