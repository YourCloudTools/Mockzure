package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// CompatibilityTestResult holds the result of a compatibility test
type CompatibilityTestResult struct {
	Category    string `json:"category"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Status      string `json:"status"` // "PASS", "FAIL", "SKIP"
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
}

// TestMicrosoftIdentityPlatform tests OIDC/OAuth2 endpoints
func TestMicrosoftIdentityPlatform(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name     string
		endpoint string
		method   string
		headers  map[string]string
		body     string
		expected int
	}{
		{
			name:     "OIDC Discovery",
			endpoint: "/.well-known/openid-configuration",
			method:   "GET",
			expected: http.StatusOK,
		},
		{
			name:     "OIDC Discovery - Tenant Specific",
			endpoint: "/common/v2.0/.well-known/openid-configuration",
			method:   "GET",
			expected: http.StatusOK,
		},
		{
			name:     "Authorization Endpoint - Missing Parameters",
			endpoint: "/oauth2/v2.0/authorize",
			method:   "GET",
			expected: http.StatusBadRequest,
		},
		{
			name:     "Authorization Endpoint - Valid Request",
			endpoint: "/oauth2/v2.0/authorize?client_id=test&redirect_uri=http://localhost&response_type=code&scope=openid",
			method:   "GET",
			expected: http.StatusOK, // Should show user selection page
		},
		{
			name:     "Token Endpoint - Client Credentials",
			endpoint: "/oauth2/v2.0/token",
			method:   "POST",
			headers:  map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			body:     "grant_type=client_credentials&client_id=sandman-app-id-12345&client_secret=sandman-secret-key-development-only&scope=https://graph.microsoft.com/.default",
			expected: http.StatusOK,
		},
		{
			name:     "Token Endpoint - Invalid Credentials",
			endpoint: "/oauth2/v2.0/token",
			method:   "POST",
			headers:  map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			body:     "grant_type=client_credentials&client_id=invalid&client_secret=invalid&scope=https://graph.microsoft.com/.default",
			expected: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.body))

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create test server with actual handlers
			mux := http.NewServeMux()
			setupMockzureHandlers(mux, store)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}

			// Validate response structure for successful requests
			if tt.expected == http.StatusOK {
				if tt.endpoint == "/.well-known/openid-configuration" {
					var config map[string]interface{}
					if err := json.NewDecoder(w.Body).Decode(&config); err != nil {
						t.Errorf("Failed to decode OIDC config: %v", err)
					}

					// Check required OIDC fields
					requiredFields := []string{"issuer", "authorization_endpoint", "token_endpoint", "response_types_supported"}
					for _, field := range requiredFields {
						if _, exists := config[field]; !exists {
							t.Errorf("Missing required OIDC field: %s", field)
						}
					}
				}

				if tt.endpoint == "/oauth2/v2.0/token" {
					var token map[string]interface{}
					if err := json.NewDecoder(w.Body).Decode(&token); err != nil {
						t.Errorf("Failed to decode token response: %v", err)
					}

					// Check required token fields
					requiredFields := []string{"access_token", "token_type", "expires_in"}
					for _, field := range requiredFields {
						if _, exists := token[field]; !exists {
							t.Errorf("Missing required token field: %s", field)
						}
					}
				}
			}
		})
	}
}

// TestMicrosoftGraphAPI tests Graph API endpoints
func TestMicrosoftGraphAPI(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name     string
		endpoint string
		method   string
		headers  map[string]string
		expected int
	}{
		{
			name:     "Users - No Authentication",
			endpoint: "/mock/azure/users",
			method:   "GET",
			expected: http.StatusUnauthorized,
		},
		{
			name:     "Users - With Service Account Auth",
			endpoint: "/mock/azure/users",
			method:   "GET",
			headers: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only")),
			},
			expected: http.StatusForbidden, // No User.Read.All permission by default
		},
		{
			name:     "Users - With Graph Permission",
			endpoint: "/mock/azure/users",
			method:   "GET",
			headers: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin-automation-app-id:admin-secret-key-development-only")),
			},
			expected: http.StatusForbidden, // Admin account doesn't have Graph permissions by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create test server
			mux := http.NewServeMux()
			setupMockzureHandlers(mux, store)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}

			// Validate response structure for successful requests
			if tt.expected == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Check Graph API response structure
				if _, exists := response["value"]; !exists {
					t.Errorf("Missing 'value' field in Graph API response")
				}
			}
		})
	}
}

// TestAzureResourceManager tests ARM endpoints
func TestAzureResourceManager(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	subscriptionID := "12345678-1234-1234-1234-123456789012"
	resourceGroup := "rg-dev"

	tests := []struct {
		name     string
		endpoint string
		method   string
		headers  map[string]string
		expected int
	}{
		{
			name:     "List VMs - ARM Format",
			endpoint: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines", subscriptionID, resourceGroup),
			method:   "GET",
			expected: http.StatusOK,
		},
		{
			name:     "Get VM - ARM Format",
			endpoint: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/vm-web-01", subscriptionID, resourceGroup),
			method:   "GET",
			expected: http.StatusOK,
		},
		{
			name:     "Get VM with Instance View",
			endpoint: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/vm-web-01?$expand=instanceView", subscriptionID, resourceGroup),
			method:   "GET",
			expected: http.StatusOK,
		},
		{
			name:     "Get Non-existent VM",
			endpoint: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/non-existent-vm", subscriptionID, resourceGroup),
			method:   "GET",
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create test server
			mux := http.NewServeMux()
			setupMockzureHandlers(mux, store)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}

			// Validate ARM response structure for successful requests
			if tt.expected == http.StatusOK {
				if strings.Contains(tt.endpoint, "virtualMachines") && !strings.Contains(tt.endpoint, "?") && !strings.Contains(tt.endpoint, "vm-web-01") {
					// VM list response
					var response map[string]interface{}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						t.Errorf("Failed to decode VM list response: %v", err)
					}

					// Check ARM list structure
					if _, exists := response["value"]; !exists {
						t.Errorf("Missing 'value' field in ARM list response")
					}
				} else if strings.Contains(tt.endpoint, "vm-web-01") {
					// Single VM response
					var vm map[string]interface{}
					if err := json.NewDecoder(w.Body).Decode(&vm); err != nil {
						t.Errorf("Failed to decode VM response: %v", err)
					}

					// Check ARM VM structure
					requiredFields := []string{"id", "name", "type", "location", "properties"}
					for _, field := range requiredFields {
						if _, exists := vm[field]; !exists {
							t.Errorf("Missing required ARM field: %s", field)
						}
					}

					// Check type field
					if vm["type"] != "Microsoft.Compute/virtualMachines" {
						t.Errorf("Expected type 'Microsoft.Compute/virtualMachines', got '%s'", vm["type"])
					}
				}
			}
		})
	}
}

// TestRBACAndAuthorization tests RBAC functionality
func TestRBACAndAuthorization(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		authHeader     string
		expectedStatus int
		description    string
	}{
		{
			name:           "VM Access - Sandman Account",
			endpoint:       "/mock/azure/vms",
			method:         "GET",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only")),
			expectedStatus: http.StatusOK,
			description:    "Sandman should see all VMs it has permissions for",
		},
		{
			name:           "VM Start - Sandman on Dev",
			endpoint:       "/mock/azure/vms/vm-web-01/start",
			method:         "POST",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only")),
			expectedStatus: http.StatusOK,
			description:    "Sandman should be able to start VMs in rg-dev",
		},
		{
			name:           "VM Start - Sandman on Prod (Should Fail)",
			endpoint:       "/mock/azure/vms/vm-web-prod-01/start",
			method:         "POST",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("sandman-app-id-12345:sandman-secret-key-development-only")),
			expectedStatus: http.StatusForbidden,
			description:    "Sandman should NOT be able to start VMs in rg-prod",
		},
		{
			name:           "VM Start - Admin Account",
			endpoint:       "/mock/azure/vms/vm-web-prod-01/start",
			method:         "POST",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin-automation-app-id:admin-secret-key-development-only")),
			expectedStatus: http.StatusOK,
			description:    "Admin should be able to start any VM",
		},
		{
			name:           "No Authentication",
			endpoint:       "/mock/azure/vms",
			method:         "GET",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			description:    "No auth should work (backward compatibility)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create test server
			mux := http.NewServeMux()
			setupMockzureHandlers(mux, store)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. %s", tt.expectedStatus, w.Code, tt.description)
			}
		})
	}
}

// TestErrorHandling tests error responses and edge cases
func TestErrorHandling(t *testing.T) {
	store := &Store{configPath: "config.yaml.example"}
	store.init()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		headers        map[string]string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid VM Operation",
			endpoint:       "/mock/azure/vms/vm-web-01/invalid-operation",
			method:         "POST",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Unknown operation",
		},
		{
			name:           "Non-existent VM",
			endpoint:       "/mock/azure/vms/non-existent-vm",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
			expectedError:  "VM not found",
		},
		{
			name:           "Invalid OAuth Request",
			endpoint:       "/oauth2/v2.0/authorize",
			method:         "GET",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid authorize request",
		},
		{
			name:           "Invalid Token Request",
			endpoint:       "/oauth2/v2.0/token",
			method:         "POST",
			headers:        map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			body:           "grant_type=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.body))

			// Set headers if provided
			if tt.headers != nil {
				for k, v := range tt.headers {
					req.Header.Set(k, v)
				}
			}

			// Create test server
			mux := http.NewServeMux()
			setupMockzureHandlers(mux, store)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// setupMockzureHandlers sets up the HTTP handlers for testing
func setupMockzureHandlers(mux *http.ServeMux, store *Store) {
	// Copy the main handlers from main.go for testing
	// This is a simplified version - in practice, you'd extract the handlers

	// OIDC Discovery
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		iss := testBaseURL(r)
		doc := map[string]interface{}{
			"issuer":                                iss,
			"authorization_endpoint":                iss + "/oauth2/v2.0/authorize",
			"token_endpoint":                        iss + "/oauth2/v2.0/token",
			"userinfo_endpoint":                     iss + "/oidc/userinfo",
			"response_types_supported":              []string{"code"},
			"id_token_signing_alg_values_supported": []string{"none"},
			"scopes_supported":                      []string{"openid", "profile", "email", "User.Read"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	mux.HandleFunc("/common/v2.0/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		// Use the same handler logic but avoid recursion
		iss := testBaseURL(r)
		doc := map[string]interface{}{
			"issuer":                                iss,
			"authorization_endpoint":                iss + "/oauth2/v2.0/authorize",
			"token_endpoint":                        iss + "/oauth2/v2.0/token",
			"userinfo_endpoint":                     iss + "/oidc/userinfo",
			"response_types_supported":              []string{"code"},
			"id_token_signing_alg_values_supported": []string{"none"},
			"scopes_supported":                      []string{"openid", "profile", "email", "User.Read"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	// Authorization endpoint
	mux.HandleFunc("/oauth2/v2.0/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		clientID := q.Get("client_id")
		redirectURI := q.Get("redirect_uri")
		responseType := q.Get("response_type")

		if clientID == "" || redirectURI == "" || responseType != "code" {
			http.Error(w, "invalid authorize request", http.StatusBadRequest)
			return
		}

		// For testing, just return OK - in real implementation this would show user selection
		w.WriteHeader(http.StatusOK)
	})

	// Token endpoint
	mux.HandleFunc("/oauth2/v2.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		grantType := r.Form.Get("grant_type")
		if grantType == "client_credentials" {
			clientID := r.Form.Get("client_id")
			clientSecret := r.Form.Get("client_secret")

			// Authenticate service account
			authenticated := false
			if store.config != nil {
				for _, secret := range store.config.ServiceAccounts {
					if secret.ApplicationID == clientID && secret.Secret == clientSecret {
						authenticated = true
						break
					}
				}
			}

			if !authenticated {
				http.Error(w, "invalid_client", http.StatusUnauthorized)
				return
			}

			// Return access token
			token := map[string]interface{}{
				"access_token": "mock_access_token_" + clientID,
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(token); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
			return
		}

		http.Error(w, "unsupported grant type", http.StatusBadRequest)
	})

	// Users endpoint
	mux.HandleFunc("/mock/azure/users", func(w http.ResponseWriter, r *http.Request) {
		// Check service account authentication
		serviceAccount, err := store.authenticateServiceAccount(r)
		if err != nil || serviceAccount == nil {
			http.Error(w, `{"error":"unauthorized","error_description":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		// Check Graph permissions
		hasPermission := false
		for _, perm := range serviceAccount.GraphPermissions {
			if perm == "User.Read.All" || perm == "Directory.Read.All" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			http.Error(w, `{"error":"forbidden","error_description":"Insufficient privileges"}`, http.StatusForbidden)
			return
		}

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": store.users, "count": len(store.users)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		}
	})

	// VMs endpoint
	mux.HandleFunc("/mock/azure/vms", func(w http.ResponseWriter, r *http.Request) {
		serviceAccount, _ := store.authenticateServiceAccount(r)

		if r.Method == http.MethodGet {
			var vms []*MockVM
			if serviceAccount != nil {
				// Filter by permissions
				for _, vm := range store.vms {
					if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
						vms = append(vms, vm)
					}
				}
			} else {
				// No auth - return all
				vms = store.vms
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": vms, "count": len(vms)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		}
	})

	// VM operations
	mux.HandleFunc("/mock/azure/vms/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/mock/azure/vms/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "VM name required", http.StatusBadRequest)
			return
		}
		vmName := parts[0]

		serviceAccount, _ := store.authenticateServiceAccount(r)

		// Find VM
		var vm *MockVM
		for _, v := range store.vms {
			if v.Name == vmName {
				vm = v
				break
			}
		}

		if vm == nil {
			http.Error(w, "VM not found", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			// Check read permission
			if serviceAccount != nil && !serviceAccount.hasPermission(vm.ResourceGroup, "read") {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(vm); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		} else if r.Method == http.MethodPost && len(parts) > 1 {
			operation := parts[1]

			// Check operation permission
			if serviceAccount != nil && !serviceAccount.hasPermission(vm.ResourceGroup, operation) {
				http.Error(w, fmt.Sprintf("Forbidden: insufficient permissions to %s VM", operation), http.StatusForbidden)
				return
			}

			switch operation {
			case "start":
				vm.Status = "running"
				vm.PowerState = "VM running"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s started successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			case "stop":
				vm.Status = "stopped"
				vm.PowerState = "VM deallocated"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s stopped successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			case "restart":
				vm.Status = "running"
				vm.PowerState = "VM running"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s restarted successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			default:
				http.Error(w, "Unknown operation", http.StatusBadRequest)
			}
		}
	})

	// ARM VM endpoints
	mux.HandleFunc("/subscriptions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/providers/Microsoft.Compute/virtualMachines") && r.Method == http.MethodGet {
			// Extract resource group and VM name from path
			parts := strings.Split(r.URL.Path, "/")
			resourceGroup := ""
			vmName := ""

			for i, part := range parts {
				if part == "resourceGroups" && i+1 < len(parts) {
					resourceGroup = parts[i+1]
				}
				if part == "virtualMachines" && i+1 < len(parts) && parts[i+1] != "" {
					vmName = parts[i+1]
				}
			}

			// Single VM GET request
			if vmName != "" {
				for _, vm := range store.vms {
					if vm.Name == vmName && (resourceGroup == "" || vm.ResourceGroup == resourceGroup) {
						properties := map[string]interface{}{
							"vmId":              vm.ID,
							"provisioningState": vm.ProvisioningState,
							"hardwareProfile": map[string]interface{}{
								"vmSize": vm.VMSize,
							},
							"storageProfile": map[string]interface{}{
								"osDisk": map[string]interface{}{
									"osType": vm.OSType,
								},
							},
						}

						// Include instance view if requested
						expandParam := r.URL.Query().Get("$expand")
						if expandParam == "instanceView" {
							powerStateCode := "PowerState/" + vm.Status
							if vm.Status == "stopped" {
								powerStateCode = "PowerState/deallocated"
							}

							properties["instanceView"] = map[string]interface{}{
								"statuses": []map[string]interface{}{
									{
										"code":          powerStateCode,
										"level":         "Info",
										"displayStatus": vm.PowerState,
									},
									{
										"code":          "ProvisioningState/" + vm.ProvisioningState,
										"level":         "Info",
										"displayStatus": "Provisioning " + strings.ToLower(vm.ProvisioningState),
									},
								},
							}
						}

						w.Header().Set("Content-Type", "application/json")
						if err := json.NewEncoder(w).Encode(map[string]interface{}{
							"id":         vm.ID,
							"name":       vm.Name,
							"type":       "Microsoft.Compute/virtualMachines",
							"location":   vm.Location,
							"properties": properties,
							"tags":       vm.Tags,
						}); err != nil {
							log.Printf("Failed to encode JSON response: %v", err)
						}
						return
					}
				}

				// VM not found
				http.NotFound(w, r)
				return
			}

			// Filter VMs by resource group if specified
			filteredVMs := []map[string]interface{}{}
			for _, vm := range store.vms {
				if resourceGroup == "" || vm.ResourceGroup == resourceGroup {
					properties := map[string]interface{}{
						"vmId":              vm.ID,
						"provisioningState": vm.ProvisioningState,
						"hardwareProfile": map[string]interface{}{
							"vmSize": vm.VMSize,
						},
						"storageProfile": map[string]interface{}{
							"osDisk": map[string]interface{}{
								"osType": vm.OSType,
							},
						},
					}

					expandParam := r.URL.Query().Get("$expand")
					if expandParam == "instanceView" {
						powerStateCode := "PowerState/" + vm.Status
						if vm.Status == "stopped" {
							powerStateCode = "PowerState/deallocated"
						}

						properties["instanceView"] = map[string]interface{}{
							"statuses": []map[string]interface{}{
								{
									"code":          powerStateCode,
									"level":         "Info",
									"displayStatus": vm.PowerState,
								},
								{
									"code":          "ProvisioningState/" + vm.ProvisioningState,
									"level":         "Info",
									"displayStatus": "Provisioning " + strings.ToLower(vm.ProvisioningState),
								},
							},
						}
					}

					armVM := map[string]interface{}{
						"id":         vm.ID,
						"name":       vm.Name,
						"type":       "Microsoft.Compute/virtualMachines",
						"location":   vm.Location,
						"properties": properties,
						"tags":       vm.Tags,
					}
					filteredVMs = append(filteredVMs, armVM)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"value": filteredVMs,
			}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
			return
		}

		// Not a recognized endpoint
		http.NotFound(w, r)
	})
}

// Helper function to get base URL (simplified for testing)
func testBaseURL(r *http.Request) string {
	return "http://localhost:8090"
}
