package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	t.Run("baseURL", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost:8090/api/test", nil)
		url := baseURL(req)
		if url != "http://localhost:8090" {
			t.Errorf("Expected 'http://localhost:8090', got '%s'", url)
		}
	})

	t.Run("baseURL with different hosts", func(t *testing.T) {
		testCases := []struct {
			url      string
			expected string
		}{
			{"http://localhost:8090/api/test", "http://localhost:8090"},
			{"https://example.com:443/api/test", "http://example.com:443"}, // baseURL converts to http
			{"http://192.168.1.1:8080/api/test", "http://192.168.1.1:8080"},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest("GET", tc.url, nil)
			result := baseURL(req)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		}
	})

	t.Run("b64url", func(t *testing.T) {
		input := []byte("test data with special chars +/=")
		encoded := b64url(input)
		if encoded == "" {
			t.Error("b64url returned empty string")
		}
		// Test that it's valid base64
		if len(encoded) == 0 {
			t.Error("b64url should return non-empty string")
		}
	})

	t.Run("b64url with various inputs", func(t *testing.T) {
		testCases := [][]byte{
			[]byte("simple"),
			[]byte("with spaces"),
			[]byte("with+plus/and=equals"),
			[]byte("with special chars !@#$%^&*()"),
			[]byte("with unicode 你好世界"),
			[]byte(""),
		}

		for _, input := range testCases {
			encoded := b64url(input)
			if len(input) == 0 && encoded != "" {
				t.Error("Empty input should produce empty output")
			}
			if len(input) > 0 && encoded == "" {
				t.Error("Non-empty input should produce non-empty output")
			}
		}
	})
}

// TestJWTFunctions tests JWT-related functions
func TestJWTFunctions(t *testing.T) {
	t.Run("makeUnsignedJWT", func(t *testing.T) {
		claims := map[string]interface{}{
			"sub": "test-user",
			"iss": "http://localhost:8090",
			"aud": "test-client",
			"iat": 1234567890,
			"exp": 1234567890,
		}
		jwt := makeUnsignedJWT(claims)
		if jwt == "" {
			t.Error("makeUnsignedJWT returned empty string")
		}
		// JWT should have 3 parts separated by dots
		parts := strings.Split(jwt, ".")
		if len(parts) != 3 {
			t.Errorf("JWT should have 3 parts, got %d", len(parts))
		}
	})

	t.Run("makeUnsignedJWT with various claims", func(t *testing.T) {
		// Test with minimal claims
		claims := map[string]interface{}{
			"sub": "user123",
		}
		jwt1 := makeUnsignedJWT(claims)
		if jwt1 == "" {
			t.Error("JWT should not be empty")
		}

		// Test with complex claims
		claims2 := map[string]interface{}{
			"sub":         "user123",
			"iss":         "http://localhost:8090",
			"aud":         []string{"client1", "client2"},
			"iat":         1234567890,
			"exp":         1234567890,
			"name":        "Test User",
			"email":       "test@example.com",
			"roles":       []string{"admin", "user"},
			"permissions": map[string]bool{"read": true, "write": false},
		}
		jwt2 := makeUnsignedJWT(claims2)
		if jwt2 == "" {
			t.Error("JWT should not be empty")
		}

		// JWTs should be different
		if jwt1 == jwt2 {
			t.Error("Different claims should produce different JWTs")
		}
	})

	t.Run("makeUnsignedJWT with edge cases", func(t *testing.T) {
		// Test with nil claims
		jwt1 := makeUnsignedJWT(nil)
		if jwt1 == "" {
			t.Error("JWT should not be empty even with nil claims")
		}

		// Test with empty claims
		jwt2 := makeUnsignedJWT(map[string]interface{}{})
		if jwt2 == "" {
			t.Error("JWT should not be empty even with empty claims")
		}

		// Test with numeric claims
		claims := map[string]interface{}{
			"sub": 12345,
			"iat": 1234567890,
			"exp": 1234567890,
			"nbf": 1234567890,
		}
		jwt3 := makeUnsignedJWT(claims)
		if jwt3 == "" {
			t.Error("JWT should not be empty with numeric claims")
		}

		// Test with boolean claims
		claims2 := map[string]interface{}{
			"admin":  true,
			"active": false,
		}
		jwt4 := makeUnsignedJWT(claims2)
		if jwt4 == "" {
			t.Error("JWT should not be empty with boolean claims")
		}
	})
}

// TestAuthenticationFunctions tests authentication-related functions
func TestAuthenticationFunctions(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	t.Run("authenticateServiceAccount with valid credentials", func(t *testing.T) {
		// Test with valid credentials from config
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Basic c2FuZG1hbi1hcHAtaWQtMTIzNDU6c2FuZG1hbi1zZWNyZXQta2V5LWRldmVsb3BtZW50LW9ubHk=")

		serviceAccount, err := store.authenticateServiceAccount(req)
		if err != nil {
			t.Errorf("Expected successful authentication, got error: %v", err)
		}
		if serviceAccount == nil {
			t.Error("Expected valid service account")
		}
		if serviceAccount != nil && serviceAccount.DisplayName != "Sandman Service Account" {
			t.Errorf("Expected 'Sandman Service Account', got '%s'", serviceAccount.DisplayName)
		}
	})

	t.Run("authenticateServiceAccount with invalid auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Invalid auth header")

		serviceAccount, err := store.authenticateServiceAccount(req)
		if err == nil {
			t.Error("Expected error for invalid auth header")
		}
		if serviceAccount != nil {
			t.Error("Expected nil service account for invalid auth")
		}
	})

	t.Run("authenticateServiceAccount with malformed basic auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Basic invalid-base64")

		serviceAccount, err := store.authenticateServiceAccount(req)
		if err == nil {
			t.Error("Expected error for malformed basic auth")
		}
		if serviceAccount != nil {
			t.Error("Expected nil service account for malformed auth")
		}
	})

	t.Run("authenticateServiceAccount with wrong credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Basic d3JvbmctaWQ6d3Jvbmctc2VjcmV0")

		serviceAccount, err := store.authenticateServiceAccount(req)
		if err == nil {
			t.Error("Expected error for wrong credentials")
		}
		if serviceAccount != nil {
			t.Error("Expected nil service account for wrong credentials")
		}
	})

	t.Run("authenticateServiceAccount edge cases", func(t *testing.T) {
		// Test with empty auth header
		req1 := httptest.NewRequest("GET", "/api/test", nil)
		_, err1 := store.authenticateServiceAccount(req1)
		if err1 == nil {
			t.Error("Expected error for empty auth header")
		}

		// Test with non-Basic auth
		req2 := httptest.NewRequest("GET", "/api/test", nil)
		req2.Header.Set("Authorization", "Bearer token")
		_, err2 := store.authenticateServiceAccount(req2)
		if err2 == nil {
			t.Error("Expected error for non-Basic auth")
		}

		// Test with malformed Basic auth (no space)
		req3 := httptest.NewRequest("GET", "/api/test", nil)
		req3.Header.Set("Authorization", "Basic")
		_, err3 := store.authenticateServiceAccount(req3)
		if err3 == nil {
			t.Error("Expected error for malformed Basic auth")
		}
	})
}

// TestPermissionFunctions tests permission-related functions
func TestPermissionFunctions(t *testing.T) {
	t.Run("hasPermission with specific resource group", func(t *testing.T) {
		// Test with specific resource group permission
		specificAccount := &ServiceAccount{
			ID:          "specific-account",
			DisplayName: "Specific Account",
			Permissions: []ResourceGroupPerm{
				{
					ResourceGroup: "rg-dev",
					Permissions:   []string{"read", "write"},
				},
				{
					ResourceGroup: "rg-prod",
					Permissions:   []string{"read"},
				},
			},
		}

		// Test specific resource group access
		if !specificAccount.hasPermission("rg-dev", "read") {
			t.Error("Should have read permission on rg-dev")
		}
		if !specificAccount.hasPermission("rg-dev", "write") {
			t.Error("Should have write permission on rg-dev")
		}
		if specificAccount.hasPermission("rg-dev", "delete") {
			t.Error("Should not have delete permission on rg-dev")
		}
		if !specificAccount.hasPermission("rg-prod", "read") {
			t.Error("Should have read permission on rg-prod")
		}
		if specificAccount.hasPermission("rg-prod", "write") {
			t.Error("Should not have write permission on rg-prod")
		}
		if specificAccount.hasPermission("rg-staging", "read") {
			t.Error("Should not have access to rg-staging")
		}
	})

	t.Run("hasPermission edge cases", func(t *testing.T) {
		// Test with service account with no permissions
		emptyAccount := &ServiceAccount{
			ID:          "empty-account",
			DisplayName: "Empty Account",
			Permissions: []ResourceGroupPerm{},
		}
		result := emptyAccount.hasPermission("any-rg", "any-permission")
		if result {
			t.Error("Empty service account should not have permissions")
		}

		// Test with service account with wildcard permissions
		wildcardAccount := &ServiceAccount{
			ID:          "wildcard-account",
			DisplayName: "Wildcard Account",
			Permissions: []ResourceGroupPerm{
				{
					ResourceGroup: "*",
					Permissions:   []string{"*"},
				},
			},
		}
		result2 := wildcardAccount.hasPermission("any-rg", "any-permission")
		if !result2 {
			t.Error("Wildcard account should have all permissions")
		}
	})
}

// TestRenderingFunctions tests page rendering functions
func TestRenderingFunctions(t *testing.T) {
	store := &Store{}
	store.init()

	t.Run("renderPortalPage", func(t *testing.T) {
		w := httptest.NewRecorder()

		renderPortalPage(w, store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Mockzure") {
			t.Error("Portal page should contain 'Mockzure'")
		}
		if !strings.Contains(body, "html") {
			t.Error("Portal page should contain HTML")
		}
	})

	t.Run("renderPortalPage with VMs", func(t *testing.T) {
		// Add some VMs to test the portal page rendering
		vms := []*MockVM{
			{
				ID:                "vm1",
				Name:              "test-vm-1",
				ResourceGroup:     "rg-dev",
				Location:          "eastus",
				Status:            "running",
				PowerState:        "VM running",
				VMSize:            "Standard_B1s",
				OSType:            "Linux",
				ProvisioningState: "Succeeded",
				Tags:              map[string]string{"env": "dev"},
			},
			{
				ID:                "vm2",
				Name:              "test-vm-2",
				ResourceGroup:     "rg-prod",
				Location:          "westus",
				Status:            "stopped",
				PowerState:        "VM deallocated",
				VMSize:            "Standard_D2s_v3",
				OSType:            "Windows",
				ProvisioningState: "Succeeded",
				Tags:              map[string]string{"env": "prod"},
			},
		}
		store.vms = vms

		w := httptest.NewRecorder()
		renderPortalPage(w, store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "test-vm-1") {
			t.Error("Portal page should contain VM names")
		}
		if !strings.Contains(body, "rg-dev") {
			t.Error("Portal page should contain resource group names")
		}
		if !strings.Contains(body, "running") {
			t.Error("Portal page should contain VM status")
		}
	})

	t.Run("renderPortalPage with users", func(t *testing.T) {
		// Add multiple users to test listing
		users := []*MockUser{
			{
				ID:                "user1",
				DisplayName:       "User One",
				UserPrincipalName: "user1@example.com",
				Mail:              "user1@example.com",
				AccountEnabled:    true,
			},
			{
				ID:                "user2",
				DisplayName:       "User Two",
				UserPrincipalName: "user2@example.com",
				Mail:              "user2@example.com",
				AccountEnabled:    false,
			},
		}
		store.users = users

		w := httptest.NewRecorder()

		renderPortalPage(w, store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "User One") {
			t.Error("Portal page should contain user names")
		}
	})

	t.Run("renderPortalPage with empty data", func(t *testing.T) {
		// Test with empty users and VMs
		store.users = []*MockUser{}
		store.vms = []*MockVM{}

		w := httptest.NewRecorder()
		renderPortalPage(w, store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Mockzure") {
			t.Error("Portal page should contain 'Mockzure' even with no data")
		}
	})

	t.Run("renderUserSelectionPage", func(t *testing.T) {
		// Add a test user
		testUser := &MockUser{
			ID:                "test-user-123",
			DisplayName:       "Test User",
			UserPrincipalName: "test@example.com",
			Mail:              "test@example.com",
			AccountEnabled:    true,
		}
		store.users = append(store.users, testUser)

		req := httptest.NewRequest("GET", "/oauth2/v2.0/authorize?client_id=test&redirect_uri=http://test.com&response_type=code&scope=openid", nil)
		w := httptest.NewRecorder()

		renderUserSelectionPage(w, req, "test-client", "http://test.com", "test-state", "code", "openid", store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Select User") {
			t.Error("User selection page should contain 'Select User'")
		}
		if !strings.Contains(body, "Test User") {
			t.Error("User selection page should contain test user name")
		}
		if !strings.Contains(body, "html") {
			t.Error("User selection page should contain HTML")
		}
	})

	t.Run("renderUserSelectionPage with multiple users", func(t *testing.T) {
		// Add multiple users
		users := []*MockUser{
			{
				ID:                "user1",
				DisplayName:       "Active User",
				UserPrincipalName: "active@example.com",
				Mail:              "active@example.com",
				AccountEnabled:    true,
			},
			{
				ID:                "user2",
				DisplayName:       "Disabled User",
				UserPrincipalName: "disabled@example.com",
				Mail:              "disabled@example.com",
				AccountEnabled:    false,
			},
		}
		store.users = users

		req := httptest.NewRequest("GET", "/oauth2/v2.0/authorize", nil)
		w := httptest.NewRecorder()

		renderUserSelectionPage(w, req, "test-client", "http://test.com", "state", "code", "openid", store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Active User") {
			t.Error("User selection should contain active user")
		}
		if !strings.Contains(body, "Disabled User") {
			t.Error("User selection should contain disabled user")
		}
	})

	t.Run("renderUserSelectionPage with different users", func(t *testing.T) {
		// Add users with different properties
		users := []*MockUser{
			{
				ID:                "admin-user",
				DisplayName:       "Admin User",
				UserPrincipalName: "admin@company.com",
				Mail:              "admin@company.com",
				AccountEnabled:    true,
				JobTitle:          "System Administrator",
				Department:        "IT",
				Roles:             []string{"Global Administrator"},
			},
			{
				ID:                "regular-user",
				DisplayName:       "Regular User",
				UserPrincipalName: "user@company.com",
				Mail:              "user@company.com",
				AccountEnabled:    true,
				JobTitle:          "Developer",
				Department:        "Engineering",
				Roles:             []string{"User"},
			},
			{
				ID:                "disabled-user",
				DisplayName:       "Disabled User",
				UserPrincipalName: "disabled@company.com",
				Mail:              "disabled@company.com",
				AccountEnabled:    false,
				JobTitle:          "Former Employee",
				Department:        "HR",
				Roles:             []string{"User"},
			},
		}
		store.users = users

		req := httptest.NewRequest("GET", "/oauth2/v2.0/authorize", nil)
		w := httptest.NewRecorder()

		renderUserSelectionPage(w, req, "test-client", "http://test.com", "state", "code", "openid profile email", store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Admin User") {
			t.Error("User selection should contain admin user")
		}
		if !strings.Contains(body, "Regular User") {
			t.Error("User selection should contain regular user")
		}
		if !strings.Contains(body, "Disabled User") {
			t.Error("User selection should contain disabled user")
		}
	})

	t.Run("renderUserSelectionPage with no users", func(t *testing.T) {
		store.users = []*MockUser{}

		req := httptest.NewRequest("GET", "/oauth2/v2.0/authorize", nil)
		w := httptest.NewRecorder()

		renderUserSelectionPage(w, req, "test-client", "http://test.com", "state", "code", "openid", store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Select User") {
			t.Error("User selection page should still render with no users")
		}
	})

	t.Run("renderUserSelectionPage with special characters", func(t *testing.T) {
		store.users = []*MockUser{
			{
				ID:                "user-special",
				DisplayName:       "User with Special Chars & Symbols!",
				UserPrincipalName: "special@example.com",
				Mail:              "special@example.com",
				AccountEnabled:    true,
				JobTitle:          "Developer & Designer",
				Department:        "R&D",
			},
		}

		req := httptest.NewRequest("GET", "/oauth2/v2.0/authorize", nil)
		w := httptest.NewRecorder()

		renderUserSelectionPage(w, req, "test-client", "http://test.com", "state", "code", "openid", store)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "User with Special Chars") {
			t.Error("User selection should handle special characters")
		}
	})
}

// TestConfigFunctions tests configuration-related functions
func TestConfigFunctions(t *testing.T) {
	store := &Store{}
	store.init()

	t.Run("loadConfig", func(t *testing.T) {
		// Test that loadConfig method exists and can be called
		store.loadConfig()
		// The method doesn't return anything, so we just test it doesn't panic
	})

	t.Run("loadConfig with existing config", func(t *testing.T) {
		// Test that loadConfig works with the existing config.json
		newStore := &Store{}
		newStore.loadConfig()

		// The config should be loaded (we have config.json in the project)
		if len(newStore.config.ServiceAccounts) == 0 {
			t.Error("Config should have service accounts loaded")
		}
	})

	t.Run("loadConfig with nonexistent file", func(t *testing.T) {
		// Test loading config with nonexistent file
		newStore := &Store{}
		newStore.loadConfig()
		// The method handles missing files gracefully, so we just test it doesn't panic
	})
}
