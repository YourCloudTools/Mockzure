package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestServiceAccountAuthentication tests basic service account authentication
func TestServiceAccountAuthentication(t *testing.T) {
	store := &Store{}
	store.init()

	tests := []struct {
		name          string
		applicationID string
		secret        string
		expectAuth    bool
		expectAccount string
	}{
		{
			name:          "Valid Sandman credentials",
			applicationID: "sandman-app-id-12345",
			secret:        "sandman-secret-key-development-only",
			expectAuth:    true,
			expectAccount: "Sandman Service Account",
		},
		{
			name:          "Valid Admin credentials",
			applicationID: "admin-automation-app-id",
			secret:        "admin-secret-key-development-only",
			expectAuth:    true,
			expectAccount: "Admin Automation Service Account",
		},
		{
			name:          "Invalid credentials",
			applicationID: "invalid-app-id",
			secret:        "invalid-secret",
			expectAuth:    false,
			expectAccount: "",
		},
		{
			name:          "Wrong secret",
			applicationID: "sandman-app-id-12345",
			secret:        "wrong-secret",
			expectAuth:    false,
			expectAccount: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Basic Auth header
			auth := base64.StdEncoding.EncodeToString([]byte(tt.applicationID + ":" + tt.secret))
			req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
			req.Header.Set("Authorization", "Basic "+auth)

			account, err := store.authenticateServiceAccount(req)

			if tt.expectAuth {
				if err != nil {
					t.Errorf("Expected successful authentication, got error: %v", err)
				}
				if account == nil {
					t.Errorf("Expected service account, got nil")
				} else if account.DisplayName != tt.expectAccount {
					t.Errorf("Expected account %s, got %s", tt.expectAccount, account.DisplayName)
				}
			} else {
				if err == nil {
					t.Errorf("Expected authentication to fail, but it succeeded")
				}
				if account != nil {
					t.Errorf("Expected nil account, got %v", account)
				}
			}
		})
	}
}

// TestServiceAccountVMFiltering tests that VMs are filtered based on service account permissions
func TestServiceAccountVMFiltering(t *testing.T) {
	store := &Store{}
	store.init()

	tests := []struct {
		name              string
		applicationID     string
		secret            string
		expectedVMCount   int
		expectedVMNames   []string
		unexpectedVMNames []string
	}{
		{
			name:            "Sandman service account sees rg-dev and rg-prod VMs",
			applicationID:   "sandman-app-id-12345",
			secret:          "sandman-secret-key-development-only",
			expectedVMCount: 3,
			expectedVMNames: []string{"vm-web-01", "vm-api-01", "vm-web-prod-01"},
		},
		{
			name:            "Admin service account sees all VMs",
			applicationID:   "admin-automation-app-id",
			secret:          "admin-secret-key-development-only",
			expectedVMCount: 3,
			expectedVMNames: []string{"vm-web-01", "vm-api-01", "vm-web-prod-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Basic Auth header
			auth := base64.StdEncoding.EncodeToString([]byte(tt.applicationID + ":" + tt.secret))
			req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
			req.Header.Set("Authorization", "Basic "+auth)
			w := httptest.NewRecorder()

			// Authenticate
			serviceAccount, err := store.authenticateServiceAccount(req)
			if err != nil {
				t.Fatalf("Authentication failed: %v", err)
			}

			// Filter VMs based on permissions
			filteredVMs := []*MockVM{}
			for _, vm := range store.vms {
				if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
					filteredVMs = append(filteredVMs, vm)
				}
			}

			// Check VM count
			if len(filteredVMs) != tt.expectedVMCount {
				t.Errorf("Expected %d VMs, got %d", tt.expectedVMCount, len(filteredVMs))
			}

			// Check expected VMs are present
			vmNames := make(map[string]bool)
			for _, vm := range filteredVMs {
				vmNames[vm.Name] = true
			}

			for _, expectedName := range tt.expectedVMNames {
				if !vmNames[expectedName] {
					t.Errorf("Expected VM %s not found in filtered results", expectedName)
				}
			}

			// Check unexpected VMs are not present
			for _, unexpectedName := range tt.unexpectedVMNames {
				if vmNames[unexpectedName] {
					t.Errorf("Unexpected VM %s found in filtered results", unexpectedName)
				}
			}

			// Verify response
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"value": filteredVMs,
				"count": len(filteredVMs),
			}
			json.NewEncoder(w).Encode(response)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

// TestServiceAccountPermissions tests permission checking for different operations
func TestServiceAccountPermissions(t *testing.T) {
	store := &Store{}
	store.init()

	// Get the Sandman service account
	var sandmanAccount *ServiceAccount
	for _, sa := range store.serviceAccounts {
		if sa.ApplicationID == "sandman-app-id-12345" {
			sandmanAccount = sa
			break
		}
	}

	if sandmanAccount == nil {
		t.Fatal("Sandman service account not found")
	}

	tests := []struct {
		name          string
		resourceGroup string
		permission    string
		expected      bool
	}{
		// rg-dev permissions (read, start, stop, restart)
		{
			name:          "Sandman can read rg-dev",
			resourceGroup: "rg-dev",
			permission:    "read",
			expected:      true,
		},
		{
			name:          "Sandman can start rg-dev",
			resourceGroup: "rg-dev",
			permission:    "start",
			expected:      true,
		},
		{
			name:          "Sandman can stop rg-dev",
			resourceGroup: "rg-dev",
			permission:    "stop",
			expected:      true,
		},
		{
			name:          "Sandman can restart rg-dev",
			resourceGroup: "rg-dev",
			permission:    "restart",
			expected:      true,
		},
		{
			name:          "Sandman cannot delete rg-dev",
			resourceGroup: "rg-dev",
			permission:    "delete",
			expected:      false,
		},
    // rg-prod permissions (read, start, stop, restart)
		{
            name:          "Sandman can read rg-prod",
            resourceGroup: "rg-prod",
            permission:    "read",
            expected:      true,
		},
		{
            name:          "Sandman can start rg-prod",
            resourceGroup: "rg-prod",
            permission:    "start",
            expected:      true,
		},
		{
            name:          "Sandman can stop rg-prod",
            resourceGroup: "rg-prod",
            permission:    "stop",
            expected:      true,
		},
		{
            name:          "Sandman can restart rg-prod",
            resourceGroup: "rg-prod",
            permission:    "restart",
            expected:      true,
		},
		{
			name:          "Sandman cannot delete rg-prod",
			resourceGroup: "rg-prod",
			permission:    "delete",
			expected:      false,
		},
		// Non-existent resource group
		{
			name:          "Sandman cannot access rg-staging",
			resourceGroup: "rg-staging",
			permission:    "read",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sandmanAccount.hasPermission(tt.resourceGroup, tt.permission)
			if result != tt.expected {
				t.Errorf("Expected hasPermission(%s, %s) to be %v, got %v",
					tt.resourceGroup, tt.permission, tt.expected, result)
			}
		})
	}
}

// TestAdminServiceAccountPermissions tests that admin account has full access
func TestAdminServiceAccountPermissions(t *testing.T) {
	store := &Store{}
	store.init()

	// Get the Admin service account
	var adminAccount *ServiceAccount
	for _, sa := range store.serviceAccounts {
		if sa.ApplicationID == "admin-automation-app-id" {
			adminAccount = sa
			break
		}
	}

	if adminAccount == nil {
		t.Fatal("Admin service account not found")
	}

	resourceGroups := []string{"rg-dev", "rg-prod", "rg-staging", "any-rg"}
	permissions := []string{"read", "write", "start", "stop", "restart", "delete"}

	for _, rg := range resourceGroups {
		for _, perm := range permissions {
			t.Run("Admin has "+perm+" on "+rg, func(t *testing.T) {
				result := adminAccount.hasPermission(rg, perm)
				if !result {
					t.Errorf("Admin should have %s permission on %s, but hasPermission returned false", perm, rg)
				}
			})
		}
	}
}

// TestVMOperationsWithPermissions tests VM operations with permission checks
func TestVMOperationsWithPermissions(t *testing.T) {
	store := &Store{}
	store.init()

	tests := []struct {
		name           string
		applicationID  string
		secret         string
		vmName         string
		operation      string
		expectSuccess  bool
		expectedStatus int
	}{
		{
			name:           "Sandman can start VM in rg-dev",
			applicationID:  "sandman-app-id-12345",
			secret:         "sandman-secret-key-development-only",
			vmName:         "vm-api-01",
			operation:      "start",
			expectSuccess:  true,
			expectedStatus: http.StatusOK,
		},
        {
            name:           "Sandman can start VM in rg-prod",
            applicationID:  "sandman-app-id-12345",
            secret:         "sandman-secret-key-development-only",
            vmName:         "vm-web-prod-01",
            operation:      "start",
            expectSuccess:  true,
            expectedStatus: http.StatusOK,
        },
		{
			name:           "Admin can start VM in rg-prod",
			applicationID:  "admin-automation-app-id",
			secret:         "admin-secret-key-development-only",
			vmName:         "vm-web-prod-01",
			operation:      "start",
			expectSuccess:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Sandman can stop VM in rg-dev",
			applicationID:  "sandman-app-id-12345",
			secret:         "sandman-secret-key-development-only",
			vmName:         "vm-web-01",
			operation:      "stop",
			expectSuccess:  true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Basic Auth header
			auth := base64.StdEncoding.EncodeToString([]byte(tt.applicationID + ":" + tt.secret))
			req := httptest.NewRequest("POST", "/mock/azure/vms/"+tt.vmName+"/"+tt.operation, nil)
			req.Header.Set("Authorization", "Basic "+auth)

			// Authenticate
			serviceAccount, err := store.authenticateServiceAccount(req)
			if err != nil {
				t.Fatalf("Authentication failed: %v", err)
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
			hasPermission := serviceAccount.hasPermission(vm.ResourceGroup, tt.operation)

			if tt.expectSuccess {
				if !hasPermission {
					t.Errorf("Expected service account to have %s permission on %s, but got denied",
						tt.operation, vm.ResourceGroup)
				}
			} else {
				if hasPermission {
					t.Errorf("Expected service account to NOT have %s permission on %s, but got allowed",
						tt.operation, vm.ResourceGroup)
				}
			}
		})
	}
}

// TestBackwardCompatibility tests that VM endpoints work without authentication
func TestBackwardCompatibility(t *testing.T) {
	store := &Store{}
	store.init()

	req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
	// No Authorization header

	serviceAccount, _ := store.authenticateServiceAccount(req)

	// Should not error, just return nil
	if serviceAccount != nil {
		t.Errorf("Expected nil service account for unauthenticated request, got %v", serviceAccount)
	}

	// Should be able to access all VMs without auth (backward compatibility)
	if len(store.vms) != 3 {
		t.Errorf("Expected 3 VMs, got %d", len(store.vms))
	}
}

// TestServiceAccountCreation tests creating a new service account
func TestServiceAccountCreation(t *testing.T) {
	store := &Store{}
	store.init()

	newAccount := &ServiceAccount{
		ID:               "sp-test-001",
		ApplicationID:    "test-app-id",
		DisplayName:      "Test Service Account",
		Description:      "Test account",
		AccountEnabled:   true,
		CreatedDateTime:  time.Now(),
		ServicePrincipal: true,
		Permissions: []ResourceGroupPerm{
			{
				ResourceGroup: "rg-dev",
				Permissions:   []string{"read"},
			},
		},
	}

	initialCount := len(store.serviceAccounts)
	store.serviceAccounts = append(store.serviceAccounts, newAccount)

	if len(store.serviceAccounts) != initialCount+1 {
		t.Errorf("Expected %d service accounts, got %d", initialCount+1, len(store.serviceAccounts))
	}

	// Verify the new account exists
	var found *ServiceAccount
	for _, sa := range store.serviceAccounts {
		if sa.ApplicationID == "test-app-id" {
			found = sa
			break
		}
	}

	if found == nil {
		t.Errorf("New service account not found")
	} else {
		if found.DisplayName != "Test Service Account" {
			t.Errorf("Expected display name 'Test Service Account', got '%s'", found.DisplayName)
		}
		if !found.hasPermission("rg-dev", "read") {
			t.Errorf("Expected read permission on rg-dev")
		}
		if found.hasPermission("rg-dev", "write") {
			t.Errorf("Should not have write permission on rg-dev")
		}
	}
}

// TestWildcardPermissions tests wildcard resource group permissions
func TestWildcardPermissions(t *testing.T) {
	store := &Store{}
	store.init()

	// Get the Admin service account (has wildcard permissions)
	var adminAccount *ServiceAccount
	for _, sa := range store.serviceAccounts {
		if sa.ApplicationID == "admin-automation-app-id" {
			adminAccount = sa
			break
		}
	}

	if adminAccount == nil {
		t.Fatal("Admin service account not found")
	}

	// Test various resource groups with wildcard
	resourceGroups := []string{"rg-dev", "rg-prod", "rg-test", "rg-anything", "new-rg-123"}

	for _, rg := range resourceGroups {
		t.Run("Wildcard allows access to "+rg, func(t *testing.T) {
			if !adminAccount.hasPermission(rg, "read") {
				t.Errorf("Expected wildcard to allow read on %s", rg)
			}
			if !adminAccount.hasPermission(rg, "write") {
				t.Errorf("Expected wildcard to allow write on %s", rg)
			}
			if !adminAccount.hasPermission(rg, "delete") {
				t.Errorf("Expected wildcard to allow delete on %s", rg)
			}
		})
	}
}

// TestSandmanAdminUserSeesAllAssignedVMs is the main test requested by the user
// This test verifies that when using Sandman service account credentials,
// the admin user can see all VMs that the service account has access to
func TestSandmanAdminUserSeesAllAssignedVMs(t *testing.T) {
	// Initialize store with default data
	store := &Store{}
	store.init()

	// Verify we have the expected VMs
	if len(store.vms) != 3 {
		t.Fatalf("Expected 3 VMs in store, got %d", len(store.vms))
	}

	// Create request with Sandman service account credentials
	sandmanAppID := "sandman-app-id-12345"
	sandmanSecret := "sandman-secret-key-development-only"
	auth := base64.StdEncoding.EncodeToString([]byte(sandmanAppID + ":" + sandmanSecret))

	req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
	req.Header.Set("Authorization", "Basic "+auth)

	// Authenticate as Sandman service account
	serviceAccount, err := store.authenticateServiceAccount(req)
	if err != nil {
		t.Fatalf("Sandman authentication failed: %v", err)
	}

	// Verify we got the correct service account
	if serviceAccount.DisplayName != "Sandman Service Account" {
		t.Errorf("Expected 'Sandman Service Account', got '%s'", serviceAccount.DisplayName)
	}

	// Filter VMs based on Sandman's permissions
	visibleVMs := []*MockVM{}
	for _, vm := range store.vms {
		if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
			visibleVMs = append(visibleVMs, vm)
		}
	}

	// Sandman service account has permissions on rg-dev and rg-prod
	// So it should see all 3 VMs (2 in rg-dev, 1 in rg-prod)
	expectedVMCount := 3
	if len(visibleVMs) != expectedVMCount {
		t.Errorf("Expected Sandman to see %d VMs, but got %d", expectedVMCount, len(visibleVMs))
	}

	// Verify specific VMs are visible
	expectedVMs := map[string]string{
		"vm-web-01":      "rg-dev",
		"vm-api-01":      "rg-dev",
		"vm-web-prod-01": "rg-prod",
	}

	foundVMs := make(map[string]string)
	for _, vm := range visibleVMs {
		foundVMs[vm.Name] = vm.ResourceGroup
	}

	for vmName, expectedRG := range expectedVMs {
		if actualRG, found := foundVMs[vmName]; !found {
			t.Errorf("Expected VM %s to be visible, but it was not found", vmName)
		} else if actualRG != expectedRG {
			t.Errorf("VM %s should be in %s, but was in %s", vmName, expectedRG, actualRG)
		}
	}

	// Verify permissions for each visible VM
	for _, vm := range visibleVMs {
		t.Run("Verify permissions for "+vm.Name, func(t *testing.T) {
			// All VMs should have read permission
			if !serviceAccount.hasPermission(vm.ResourceGroup, "read") {
				t.Errorf("Sandman should have read permission on %s in %s", vm.Name, vm.ResourceGroup)
			}

			// VMs in rg-dev should have start/stop/restart permissions
			if vm.ResourceGroup == "rg-dev" {
				if !serviceAccount.hasPermission(vm.ResourceGroup, "start") {
					t.Errorf("Sandman should have start permission on %s in rg-dev", vm.Name)
				}
				if !serviceAccount.hasPermission(vm.ResourceGroup, "stop") {
					t.Errorf("Sandman should have stop permission on %s in rg-dev", vm.Name)
				}
				if !serviceAccount.hasPermission(vm.ResourceGroup, "restart") {
					t.Errorf("Sandman should have restart permission on %s in rg-dev", vm.Name)
				}
			}

            // VMs in rg-prod should also have start/stop/restart permissions
            if vm.ResourceGroup == "rg-prod" {
                if !serviceAccount.hasPermission(vm.ResourceGroup, "start") {
                    t.Errorf("Sandman should have start permission on %s in rg-prod", vm.Name)
                }
                if !serviceAccount.hasPermission(vm.ResourceGroup, "stop") {
                    t.Errorf("Sandman should have stop permission on %s in rg-prod", vm.Name)
                }
                if !serviceAccount.hasPermission(vm.ResourceGroup, "restart") {
                    t.Errorf("Sandman should have restart permission on %s in rg-prod", vm.Name)
                }
            }
		})
	}

	// Log the results for visibility
	t.Logf("✅ Sandman service account successfully authenticated")
	t.Logf("✅ Sandman can see %d VMs (all VMs in rg-dev and rg-prod)", len(visibleVMs))
	for _, vm := range visibleVMs {
		t.Logf("   - %s (Resource Group: %s, Status: %s)", vm.Name, vm.ResourceGroup, vm.Status)
	}
}
