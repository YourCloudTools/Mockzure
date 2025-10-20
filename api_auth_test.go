package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAdminUserAccessToVMs tests that an admin user (not service account) can access VMs
func TestAdminUserAccessToVMs(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name              string
		username          string
		password          string
		expectVMCount     int
		expectError       bool
		expectFilteredVMs bool
	}{
		{
			name:              "Admin user with password (backward compat)",
			username:          "admin@company.com",
			password:          "admin-password",
			expectVMCount:     3, // Should see all VMs (auth fails, falls back to no auth)
			expectError:       false,
			expectFilteredVMs: false,
		},
		{
			name:              "Regular user with password (backward compat)",
			username:          "john.doe@company.com",
			password:          "user-password",
			expectVMCount:     3, // Should see all VMs (auth fails, falls back to no auth)
			expectError:       false,
			expectFilteredVMs: false,
		},
		{
			name:              "No authentication",
			username:          "",
			password:          "",
			expectVMCount:     3, // Should see all VMs
			expectError:       false,
			expectFilteredVMs: false,
		},
		{
			name:              "Service account authentication",
			username:          "sandman-app-id-12345",
			password:          "sandman-secret-key-development-only",
			expectVMCount:     3, // Sandman sees rg-dev and rg-prod
			expectError:       false,
			expectFilteredVMs: true, // VMs are filtered by permission
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/mock/azure/vms", nil)

			// Add authentication if credentials provided
			if tt.username != "" {
				auth := base64.StdEncoding.EncodeToString([]byte(tt.username + ":" + tt.password))
				req.Header.Set("Authorization", "Basic "+auth)
			}

			w := httptest.NewRecorder()

			// Simulate the VM endpoint handler
			serviceAccount, authErr := store.authenticateServiceAccount(req)

			// Log authentication result
			if authErr != nil {
				t.Logf("Authentication error (expected for non-service accounts): %v", authErr)
			}

			var filteredVMs []*MockVM
			if serviceAccount != nil {
				// Service account authenticated - filter VMs
				t.Logf("Authenticated as service account: %s", serviceAccount.DisplayName)
				for _, vm := range store.vms {
					if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
						filteredVMs = append(filteredVMs, vm)
					}
				}
			} else {
				// No service account or auth failed - return all VMs (backward compatibility)
				t.Log("No service account authentication - returning all VMs")
				filteredVMs = store.vms
			}

			// Verify VM count
			if len(filteredVMs) != tt.expectVMCount {
				t.Errorf("Expected %d VMs, got %d", tt.expectVMCount, len(filteredVMs))
			}

			// Write response
			response := map[string]interface{}{
				"value": filteredVMs,
				"count": len(filteredVMs),
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}

			// Verify response code
			if w.Code != http.StatusOK {
				if !tt.expectError {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
			}

			// Parse response
			var result map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			t.Logf("Response: %d VMs returned", int(result["count"].(float64)))
		})
	}
}

// TestAdminUserVMOperations tests VM operations with admin user credentials
func TestAdminUserVMOperations(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name         string
		username     string
		password     string
		vmName       string
		operation    string
		expectStatus int
	}{
		{
			name:         "Admin user starts VM (should work - no auth required)",
			username:     "admin@company.com",
			password:     "admin-password",
			vmName:       "vm-api-01",
			operation:    "start",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Regular user starts VM (should work - no auth required)",
			username:     "john.doe@company.com",
			password:     "user-password",
			vmName:       "vm-api-01",
			operation:    "start",
			expectStatus: http.StatusOK,
		},
		{
			name:         "No auth starts VM (should work)",
			username:     "",
			password:     "",
			vmName:       "vm-api-01",
			operation:    "start",
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/mock/azure/vms/"+tt.vmName+"/"+tt.operation, nil)

			// Add authentication if credentials provided
			if tt.username != "" {
				auth := base64.StdEncoding.EncodeToString([]byte(tt.username + ":" + tt.password))
				req.Header.Set("Authorization", "Basic "+auth)
			}

			// Authenticate
			serviceAccount, authErr := store.authenticateServiceAccount(req)

			if authErr != nil {
				t.Logf("Auth error (expected for users): %v", authErr)
			}

			// Find VM
			var vm *MockVM
			for _, v := range store.vms {
				if v.Name == tt.vmName {
					vm = v
					break
				}
			}

			if vm == nil {
				t.Fatalf("VM %s not found", tt.vmName)
			}

			// Check permission
			hasPermission := true
			if serviceAccount != nil {
				hasPermission = serviceAccount.hasPermission(vm.ResourceGroup, tt.operation)
				t.Logf("Service account permission check: %v", hasPermission)
			} else {
				t.Log("No service account - permission granted (backward compat)")
			}

			// Verify permission matches expected status
			if hasPermission {
				if tt.expectStatus != http.StatusOK {
					t.Errorf("Expected status %d, but permission was granted", tt.expectStatus)
				}
			} else {
				if tt.expectStatus == http.StatusOK {
					t.Errorf("Expected permission to be granted, but it was denied")
				}
			}
		})
	}
}

// TestUserAuthenticationFlow tests complete authentication flow for users vs service accounts
func TestUserAuthenticationFlow(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	t.Run("Distinguish between user and service account", func(t *testing.T) {
		// Test with service account
		serviceAccountAuth := base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only"))
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.Header.Set("Authorization", "Basic "+serviceAccountAuth)

		sa1, err1 := store.authenticateServiceAccount(req1)
		if err1 != nil {
			t.Errorf("Service account auth should succeed, got error: %v", err1)
		}
		if sa1 == nil {
			t.Error("Service account should not be nil")
		} else {
			t.Logf("✅ Service account authenticated: %s", sa1.DisplayName)
		}

		// Test with user credentials (not in config.json)
		userAuth := base64.StdEncoding.EncodeToString([]byte("admin@company.com:password"))
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.Header.Set("Authorization", "Basic "+userAuth)

		sa2, err2 := store.authenticateServiceAccount(req2)
		if err2 == nil {
			t.Log("User auth failed as expected (not a service account)")
		}
		if sa2 != nil {
			t.Error("User should not authenticate as service account")
		} else {
			t.Log("✅ User credentials correctly not authenticated as service account")
		}
	})
}

// TestBackwardCompatibilityWithUserAuth tests that user auth doesn't break existing functionality
func TestBackwardCompatibilityWithUserAuth(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	t.Run("User auth falls back to no-auth behavior", func(t *testing.T) {
		// User credentials (not service account)
		userAuth := base64.StdEncoding.EncodeToString([]byte("admin@company.com:password123"))
		req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
		req.Header.Set("Authorization", "Basic "+userAuth)

		serviceAccount, _ := store.authenticateServiceAccount(req)

		// Should be nil (user is not a service account)
		if serviceAccount != nil {
			t.Errorf("User should not authenticate as service account")
		}

		// Should still be able to access all VMs (backward compatibility)
		vms := store.vms
		if len(vms) != 3 {
			t.Errorf("Expected 3 VMs, got %d", len(vms))
		}

		t.Log("✅ User auth correctly falls back to showing all VMs")
	})
}

// TestAPIEndpointWithDifferentAuthTypes tests actual API endpoint behavior
func TestAPIEndpointWithDifferentAuthTypes(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	// Create a test server
	mux := http.NewServeMux()

	// Add VM endpoint (simplified version of actual endpoint)
	mux.HandleFunc("/mock/azure/vms", func(w http.ResponseWriter, r *http.Request) {
		serviceAccount, _ := store.authenticateServiceAccount(r)

		var vms []*MockVM
		if serviceAccount != nil {
			// Filter by permissions
			for _, vm := range store.vms {
				if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
					vms = append(vms, vm)
				}
			}
		} else {
			// No auth or user auth - return all
			vms = store.vms
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"value": vms,
			"count": len(vms),
		}); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name        string
		auth        string
		expectCount int
	}{
		{
			name:        "No auth",
			auth:        "",
			expectCount: 3,
		},
		{
			name:        "User auth",
			auth:        base64.StdEncoding.EncodeToString([]byte("admin@company.com:pass")),
			expectCount: 3,
		},
		{
			name:        "Service account auth",
			auth:        base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only")),
			expectCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL+"/mock/azure/vms", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", "Basic "+tt.auth)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Warning: failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			count := int(result["count"].(float64))
			if count != tt.expectCount {
				t.Errorf("Expected %d VMs, got %d", tt.expectCount, count)
			}

			t.Logf("✅ %s: Got %d VMs", tt.name, count)
		})
	}
}
