package manager

import (
	"testing"
)

func TestRedisStateManagement_GetID(t *testing.T) {

	redis, err := NewRedisStateManagement("localhost:6379")
	if err != nil {
		t.Fatalf("Failed to create Redis state management: %v", err)
	}

	name := "testing-1"
	expectedId := "12345"

	err = redis.StoreID(name, expectedId)
	if err != nil {
		t.Errorf("GetID failed: %v", err)
	}
	id := redis.GetID(name)

	if id != expectedId {
		t.Errorf("GetID failed: want %s got %s", expectedId, id)
	}
}

func TestRedisStateManagement_GetState(t *testing.T) {
	// Initialize Redis client with test instance
	redis, err := NewRedisStateManagement("localhost:6379")
	if err != nil {
		t.Fatalf("Failed to create Redis state management: %v", err)
	}

	// Test cases
	tests := []struct {
		name     string
		windowID string
		setState WindowState
		expected WindowState
	}{
		{
			name:     "Visible State",
			windowID: "test-window-1",
			setState: Visible,
			expected: Visible,
		},
		{
			name:     "NotVisible State",
			windowID: "test-window-2",
			setState: NotVisible,
			expected: NotVisible,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store the state
			err := redis.StoreID(tt.name, tt.windowID)
			if err != nil {
				t.Errorf("SetState failed: %v", err)
			}
			err = redis.SetState(tt.name, tt.setState)
			if err != nil {
				t.Errorf("SetState failed: %v", err)
			}

			// Get the state
			got := redis.GetState(tt.windowID)
			if got != tt.expected {
				t.Errorf("GetState() = %v, want %v", got, tt.expected)
			}
		})
	}

	// Clean up
	// if err := redis.ResetAll(); err != nil {
	// 	t.Errorf("Failed to clean up test data: %v", err)
	// }
}
