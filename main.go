package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Helper function to encode JSON with error handling
func encodeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// Lightweight replicas of types and behavior from Sandman's internal mock

type ResourceGroup struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Location string            `json:"location"`
	Tags     map[string]string `json:"tags"`
}

type MockVM struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ResourceGroup     string            `json:"resourceGroup"`
	Location          string            `json:"location"`
	VMSize            string            `json:"vmSize"`
	OSType            string            `json:"osType"`
	ProvisioningState string            `json:"provisioningState"`
	PowerState        string            `json:"powerState"`
	Status            string            `json:"status"`
	LastUpdated       time.Time         `json:"lastUpdated"`
	Tags              map[string]string `json:"tags"`
	Owner             string            `json:"owner"`
	CostCenter        string            `json:"costCenter"`
	Environment       string            `json:"environment"`
}

type MockAzureRole struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
	Scope       string   `json:"scope"`
}

type MockPermission struct {
	Resource      string   `json:"resource"`
	Actions       []string `json:"actions"`
	ResourceGroup string   `json:"resourceGroup"`
}

type MockUser struct {
	ID                string           `json:"id"`
	DisplayName       string           `json:"displayName"`
	UserPrincipalName string           `json:"userPrincipalName"`
	Mail              string           `json:"mail"`
	JobTitle          string           `json:"jobTitle"`
	Department        string           `json:"department"`
	OfficeLocation    string           `json:"officeLocation"`
	UserType          string           `json:"userType"`
	AccountEnabled    bool             `json:"accountEnabled"`
	Roles             []string         `json:"roles"`
	AzureRoles        []MockAzureRole  `json:"azureRoles"`
	Permissions       []MockPermission `json:"permissions"`
	ResourceGroups    []string         `json:"resourceGroups"`
	Subscriptions     []string         `json:"subscriptions"`
}

// ServiceAccount represents an Azure Service Principal / Service Account
type ServiceAccount struct {
	ID               string              `json:"id"`
	ApplicationID    string              `json:"applicationId"` // Client ID
	DisplayName      string              `json:"displayName"`
	Description      string              `json:"description"`
	AccountEnabled   bool                `json:"accountEnabled"`
	CreatedDateTime  time.Time           `json:"createdDateTime"`
	Permissions      []ResourceGroupPerm `json:"permissions"`
	ServicePrincipal bool                `json:"servicePrincipal"`
	GraphPermissions []string            `json:"graphPermissions"` // Microsoft Graph API permissions
}

// ResourceGroupPerm represents permissions for a service account on a resource group
type ResourceGroupPerm struct {
	ResourceGroup string   `json:"resourceGroup"` // Resource group name or "*" for all
	Permissions   []string `json:"permissions"`   // "read", "write", "start", "stop", "restart"
}

// ServiceAccountConfig holds the secret configuration for service accounts
type ServiceAccountConfig struct {
	ServiceAccounts []ServiceAccountSecret `json:"serviceAccounts"`
}

// ServiceAccountSecret holds the secret for a service account
type ServiceAccountSecret struct {
	ApplicationID    string   `json:"applicationId"`
	Secret           string   `json:"secret"`
	DisplayName      string   `json:"displayName,omitempty"`
	Description      string   `json:"description,omitempty"`
	GraphPermissions []string `json:"graphPermissions,omitempty"`
}

type MockEntraIDResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type MockUserInfo struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	JobTitle          string   `json:"job_title"`
	Department        string   `json:"department"`
	OfficeLocation    string   `json:"office_location"`
	Roles             []string `json:"roles"`
	AccountEnabled    bool     `json:"account_enabled"`
	UserPrincipalName string   `json:"user_principal_name"`
}

type Store struct {
	resourceGroups  []*ResourceGroup
	vms             []*MockVM
	users           []*MockUser
	serviceAccounts []*ServiceAccount
	clients         map[string]*RegisteredClient
	codes           map[string]*AuthCode
	config          *ServiceAccountConfig
}

func (s *Store) init() {
	// Initialize resource groups
	s.resourceGroups = []*ResourceGroup{
		{
			ID:       "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-dev",
			Name:     "rg-dev",
			Location: "East US",
			Tags: map[string]string{
				"Environment": "Development",
				"CostCenter":  "IT-001",
			},
		},
		{
			ID:       "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-prod",
			Name:     "rg-prod",
			Location: "West US",
			Tags: map[string]string{
				"Environment": "Production",
				"CostCenter":  "IT-001",
			},
		},
	}

	s.vms = []*MockVM{
		{
			ID:                "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-dev/providers/Microsoft.Compute/virtualMachines/vm-web-01",
			Name:              "vm-web-01",
			ResourceGroup:     "rg-dev",
			Location:          "East US",
			VMSize:            "Standard_B2s",
			OSType:            "linux",
			ProvisioningState: "Succeeded",
			PowerState:        "VM running",
			Status:            "running",
			LastUpdated:       time.Now().Add(-2 * time.Hour),
			Tags: map[string]string{
				"Environment": "Development",
				"Owner":       "john.doe@company.com",
				"CostCenter":  "IT-001",
				"Project":     "WebApp",
				"ManagedBy":   "Sandman",
			},
			Owner:       "john.doe@company.com",
			CostCenter:  "IT-001",
			Environment: "Development",
		},
		{
			ID:                "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-dev/providers/Microsoft.Compute/virtualMachines/vm-api-01",
			Name:              "vm-api-01",
			ResourceGroup:     "rg-dev",
			Location:          "East US",
			VMSize:            "Standard_B2s",
			OSType:            "linux",
			ProvisioningState: "Succeeded",
			PowerState:        "VM deallocated",
			Status:            "stopped",
			LastUpdated:       time.Now().Add(-1 * time.Hour),
			Tags: map[string]string{
				"Environment": "Development",
				"Owner":       "jane.smith@company.com",
				"CostCenter":  "IT-001",
				"Project":     "API",
				"ManagedBy":   "Sandman",
			},
			Owner:       "jane.smith@company.com",
			CostCenter:  "IT-001",
			Environment: "Development",
		},
		{
			ID:                "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-prod/providers/Microsoft.Compute/virtualMachines/vm-web-prod-01",
			Name:              "vm-web-prod-01",
			ResourceGroup:     "rg-prod",
			Location:          "West US",
			VMSize:            "Standard_D2s_v3",
			OSType:            "linux",
			ProvisioningState: "Succeeded",
			PowerState:        "VM running",
			Status:            "running",
			LastUpdated:       time.Now().Add(-30 * time.Minute),
			Tags: map[string]string{
				"Environment": "Production",
				"Owner":       "admin@company.com",
				"CostCenter":  "IT-001",
				"Project":     "WebApp",
				"ManagedBy":   "Sandman",
			},
			Owner:       "admin@company.com",
			CostCenter:  "IT-001",
			Environment: "Production",
		},
	}

	s.users = []*MockUser{
		{
			ID:                "12345678-1234-1234-1234-123456789001",
			DisplayName:       "John Doe",
			UserPrincipalName: "john.doe@company.com",
			Mail:              "john.doe@company.com",
			JobTitle:          "Senior Developer",
			Department:        "Engineering",
			OfficeLocation:    "Seattle",
			UserType:          "Member",
			AccountEnabled:    true,
			Roles:             []string{"Developer", "VM Owner"},
			AzureRoles: []MockAzureRole{{
				ID:          "contributor-001",
				Name:        "Contributor",
				Description: "Can manage all resources except access management",
				Actions:     []string{"*"},
				Scope:       "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-dev",
			}},
			Permissions: []MockPermission{{
				Resource:      "virtualMachines",
				Actions:       []string{"read", "write", "delete"},
				ResourceGroup: "rg-dev",
			}},
			ResourceGroups: []string{"rg-dev"},
			Subscriptions:  []string{"12345678-1234-1234-1234-123456789012"},
		},
		{
			ID:                "12345678-1234-1234-1234-123456789002",
			DisplayName:       "Jane Smith",
			UserPrincipalName: "jane.smith@company.com",
			Mail:              "jane.smith@company.com",
			JobTitle:          "DevOps Engineer",
			Department:        "Engineering",
			OfficeLocation:    "New York",
			UserType:          "Member",
			AccountEnabled:    true,
			Roles:             []string{"DevOps", "VM Owner"},
			AzureRoles: []MockAzureRole{{
				ID:          "contributor-002",
				Name:        "Contributor",
				Description: "Can manage all resources except access management",
				Actions:     []string{"*"},
				Scope:       "/subscriptions/12345678-1234-1234-1234-123456789012",
			}},
			Permissions: []MockPermission{{
				Resource:      "virtualMachines",
				Actions:       []string{"read", "write", "delete"},
				ResourceGroup: "rg-dev",
			}},
			ResourceGroups: []string{"rg-dev", "rg-prod"},
			Subscriptions:  []string{"12345678-1234-1234-1234-123456789012"},
		},
		{
			ID:                "12345678-1234-1234-1234-123456789003",
			DisplayName:       "Admin User",
			UserPrincipalName: "admin@company.com",
			Mail:              "admin@company.com",
			JobTitle:          "System Administrator",
			Department:        "IT",
			OfficeLocation:    "Seattle",
			UserType:          "Member",
			AccountEnabled:    true,
			Roles:             []string{"Global Administrator", "VM Administrator"},
			AzureRoles: []MockAzureRole{{
				ID:          "owner-001",
				Name:        "Owner",
				Description: "Can manage everything including access",
				Actions:     []string{"*"},
				Scope:       "/subscriptions/12345678-1234-1234-1234-123456789012",
			}},
			Permissions: []MockPermission{{
				Resource:      "*",
				Actions:       []string{"*"},
				ResourceGroup: "*",
			}},
			ResourceGroups: []string{"rg-dev", "rg-prod"},
			Subscriptions:  []string{"12345678-1234-1234-1234-123456789012"},
		},
	}

	// Service Accounts (Azure Service Principals)
	s.serviceAccounts = []*ServiceAccount{
		{
			ID:               "sp-12345678-1234-1234-1234-123456789001",
			ApplicationID:    "sandman-app-id-12345",
			DisplayName:      "Sandman Service Account",
			Description:      "Service account for Sandman to manage VMs",
			AccountEnabled:   true,
			CreatedDateTime:  time.Now().Add(-30 * 24 * time.Hour),
			ServicePrincipal: true,
			Permissions: []ResourceGroupPerm{
				{
					ResourceGroup: "rg-dev",
					Permissions:   []string{"read", "start", "stop", "restart"},
				},
				{
					ResourceGroup: "rg-prod",
					Permissions:   []string{"read"},
				},
			},
		},
		{
			ID:               "sp-12345678-1234-1234-1234-123456789002",
			ApplicationID:    "admin-automation-app-id",
			DisplayName:      "Admin Automation Service Account",
			Description:      "Service account for administrative automation",
			AccountEnabled:   true,
			CreatedDateTime:  time.Now().Add(-60 * 24 * time.Hour),
			ServicePrincipal: true,
			Permissions: []ResourceGroupPerm{
				{
					ResourceGroup: "*", // All resource groups
					Permissions:   []string{"read", "write", "start", "stop", "restart", "delete"},
				},
			},
		},
	}

	// Load service account secrets from config file
	s.loadConfig()

	// app registrations and auth codes
	s.clients = make(map[string]*RegisteredClient)
	s.codes = make(map[string]*AuthCode)
}

// loadConfig loads service account secrets from config file
func (s *Store) loadConfig() {
	configPath := "config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		defaultConfig := &ServiceAccountConfig{
			ServiceAccounts: []ServiceAccountSecret{
				{
					ApplicationID: "sandman-app-id-12345",
					Secret:        "sandman-secret-key-development-only",
				},
				{
					ApplicationID: "admin-automation-app-id",
					Secret:        "admin-secret-key-development-only",
				},
			},
		}
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal config: %v", err)
			return
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			log.Printf("Failed to write config file: %v", err)
			return
		}
		s.config = defaultConfig
		log.Printf("Created default config file: %s", configPath)
		return
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		s.config = &ServiceAccountConfig{}
		return
	}

	var config ServiceAccountConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Failed to parse config file: %v", err)
		s.config = &ServiceAccountConfig{}
		return
	}

	s.config = &config
	log.Printf("Loaded %d service account secrets from config", len(config.ServiceAccounts))

	// Update service accounts with Graph permissions from config
	for i, sa := range s.serviceAccounts {
		for _, configSA := range config.ServiceAccounts {
			if sa.ApplicationID == configSA.ApplicationID {
				if len(configSA.GraphPermissions) > 0 {
					s.serviceAccounts[i].GraphPermissions = configSA.GraphPermissions
					log.Printf("Applied Graph permissions to %s: %v", sa.ApplicationID, configSA.GraphPermissions)
				}
			}
		}
	}
}

// authenticateServiceAccount validates a service account request
func (s *Store) authenticateServiceAccount(r *http.Request) (*ServiceAccount, error) {
	// Check for service account authentication header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("no authorization header")
	}

	// Support Bearer token (OAuth 2.0)
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")

		// Mock tokens have format: "mock_access_token_{clientID}"
		if strings.HasPrefix(token, "mock_access_token_") {
			clientID := strings.TrimPrefix(token, "mock_access_token_")

			// Find and return the service account
			for _, sa := range s.serviceAccounts {
				if sa.ApplicationID == clientID && sa.AccountEnabled {
					return sa, nil
				}
			}
		}

		return nil, fmt.Errorf("invalid or expired token")
	}

	// Support Basic auth (applicationId:secret)
	if strings.HasPrefix(auth, "Basic ") {
		// Decode Basic auth
		encoded := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("invalid basic auth encoding")
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid basic auth format")
		}

		appID := parts[0]
		secret := parts[1]

		// Validate credentials
		var validSecret string
		for _, sa := range s.config.ServiceAccounts {
			if sa.ApplicationID == appID {
				validSecret = sa.Secret
				break
			}
		}

		if validSecret == "" || validSecret != secret {
			return nil, fmt.Errorf("invalid credentials")
		}

		// Find service account
		for _, sa := range s.serviceAccounts {
			if sa.ApplicationID == appID && sa.AccountEnabled {
				return sa, nil
			}
		}

		return nil, fmt.Errorf("service account not found or disabled")
	}

	return nil, fmt.Errorf("unsupported authentication method")
}

// hasPermission checks if a service account has a specific permission on a resource group
func (sa *ServiceAccount) hasPermission(resourceGroup, permission string) bool {
	for _, perm := range sa.Permissions {
		// Check for wildcard or specific resource group
		if perm.ResourceGroup != "*" && perm.ResourceGroup != resourceGroup {
			continue
		}

		// Check if permission is granted
		for _, p := range perm.Permissions {
			if p == permission || p == "*" {
				return true
			}
		}
	}
	return false
}

// OIDC app registration and code store
type RegisteredClient struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Name         string   `json:"name,omitempty"`
}

type AuthCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	Scope       string
	UserSub     string
	IssuedAt    time.Time
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func b64url(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func makeUnsignedJWT(claims map[string]interface{}) string {
	header := map[string]string{"alg": "none", "typ": "JWT"}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(claims)
	return b64url(hb) + "." + b64url(pb) + "."
}

// renderUserSelectionPage renders an HTML page for selecting a user to log in as
// renderPortalPage renders the main Mockzure portal with tabs
func renderPortalPage(w http.ResponseWriter, store *Store) {
	// Group VMs by resource group
	vmsByRG := make(map[string][]*MockVM)
	for _, vm := range store.vms {
		vmsByRG[vm.ResourceGroup] = append(vmsByRG[vm.ResourceGroup], vm)
	}

	// Stats
	running, stopped := 0, 0
	for _, v := range store.vms {
		if v.Status == "running" {
			running++
		} else {
			stopped++
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
	<title>Mockzure Portal</title>
	<script src="https://cdn.tailwindcss.com"></script>
	<style>
		.tab-content { display: none; }
		.tab-content.active { display: block; }
		.status-indicator { display: inline-block; width: 8px; height: 8px; border-radius: 50%; }
		.status-running { background-color: #10b981; }
		.status-stopped { background-color: #ef4444; }
	</style>
</head>
<body class="bg-gray-50">
	<!-- Header -->
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between items-center py-4">
				<div class="flex items-center">
					<div class="w-8 h-8 bg-gradient-to-br from-purple-500 to-blue-500 rounded-full flex items-center justify-center mr-3">
						<div class="w-3 h-3 bg-white rounded-full"></div>
					</div>
					<h1 class="text-xl font-semibold text-gray-900">Mockzure Portal</h1>
					<span class="ml-3 text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded-full">:8090</span>
				</div>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<!-- Tabs Navigation -->
		<div class="border-b border-gray-200 mb-6">
			<nav class="-mb-px flex space-x-8">
				<button id="resource-groups-tab" class="tab-button border-b-2 border-purple-500 py-2 px-1 text-sm font-medium text-gray-900" onclick="showTab('resource-groups')">
					Resource Groups
				</button>
				<button id="entra-id-tab" class="tab-button border-b-2 border-transparent py-2 px-1 text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300" onclick="showTab('entra-id')">
					Entra ID
				</button>
				<button id="settings-tab" class="tab-button border-b-2 border-transparent py-2 px-1 text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300" onclick="showTab('settings')">
					Settings
				</button>
			</nav>
		</div>

		<!-- Resource Groups Tab -->
		<div id="resource-groups-content" class="tab-content active">`

	// Resource Groups section
	for _, rg := range store.resourceGroups {
		vms := vmsByRG[rg.Name]
		html += fmt.Sprintf(`
			<div class="bg-white rounded-lg shadow mb-6 overflow-hidden">
				<div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
					<div class="flex items-center justify-between">
						<div>
							<h2 class="text-lg font-medium text-gray-900">%s</h2>
							<p class="text-sm text-gray-500">%s · %d VMs</p>
						</div>
						<a href="/mock/azure/resource-groups/%s" class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded hover:bg-blue-200">JSON</a>
					</div>
				</div>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">OS</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">`, rg.Name, rg.Location, len(vms), rg.Name)

		if len(vms) == 0 {
			html += `<tr><td colspan="6" class="px-6 py-12 text-center text-gray-500">No VMs in this resource group</td></tr>`
		} else {
			for _, vm := range vms {
				statusClass := "status-stopped"
				statusText := "Stopped"
				if vm.Status == "running" {
					statusClass = "status-running"
					statusText = "Running"
				}
				html += fmt.Sprintf(`
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap">
								<a href="/mock/azure/vms/%s" class="text-sm font-medium text-blue-600 hover:text-blue-800">%s</a>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span class="inline-flex items-center">
									<span class="status-indicator %s mr-2"></span>
									<span class="text-sm text-gray-900">%s</span>
								</span>
							</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
								<button onclick="performAction('%s', 'start')" class="text-green-600 hover:text-green-900">Start</button>
								<button onclick="performAction('%s', 'stop')" class="text-red-600 hover:text-red-900">Stop</button>
							</td>
						</tr>`, vm.Name, vm.Name, vm.VMSize, vm.OSType, vm.Owner, statusClass, statusText, vm.Name, vm.Name)
			}
		}

		html += `
					</tbody>
				</table>
			</div>
		</div>`
	}

	html += `
		</div>

		<!-- Entra ID Tab -->
		<div id="entra-id-content" class="tab-content">
			<!-- Users Section -->
			<div class="bg-white rounded-lg shadow mb-6 overflow-hidden">
				<div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
					<div class="flex items-center justify-between">
						<h2 class="text-lg font-medium text-gray-900">Users</h2>
						<a href="/mock/azure/users" class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded hover:bg-blue-200">JSON</a>
					</div>
				</div>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">UPN</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Job Title</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Department</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Roles</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">`

	for _, user := range store.users {
		statusBadge := `<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">Active</span>`
		if !user.AccountEnabled {
			statusBadge = `<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">Disabled</span>`
		}
		html += fmt.Sprintf(`
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap">%s</td>
						</tr>`, user.DisplayName, user.UserPrincipalName, user.JobTitle, user.Department, strings.Join(user.Roles, ", "), statusBadge)
	}

	html += `
					</tbody>
				</table>
			</div>
		</div>

		<!-- Service Accounts Section -->
		<div class="bg-white rounded-lg shadow mb-6 overflow-hidden">
			<div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-medium text-gray-900">Service Accounts (Service Principals)</h2>
					<a href="/mock/azure/service-accounts" class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded hover:bg-blue-200">JSON</a>
				</div>
			</div>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Application ID</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Resource Groups</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Permissions</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">`

	for _, sa := range store.serviceAccounts {
		statusBadge := `<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">Active</span>`
		if !sa.AccountEnabled {
			statusBadge = `<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">Disabled</span>`
		}

		// Build resource groups list
		rgList := []string{}
		permList := []string{}
		for _, perm := range sa.Permissions {
			rgList = append(rgList, perm.ResourceGroup)
			permList = append(permList, strings.Join(perm.Permissions, ", "))
		}

		html += fmt.Sprintf(`
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-mono">%s</td>
							<td class="px-6 py-4 text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap">%s</td>
						</tr>`, sa.DisplayName, sa.ApplicationID, sa.Description, strings.Join(rgList, ", "), strings.Join(permList, " | "), statusBadge)
	}

	html += `
					</tbody>
				</table>
			</div>
		</div>

		<!-- App Registrations Section -->
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
				<div class="flex items-center justify-between">
					<h2 class="text-lg font-medium text-gray-900">App Registrations</h2>
					<a href="/mock/azure/apps" class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded hover:bg-blue-200">JSON</a>
				</div>
			</div>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Client ID</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Redirect URIs</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Scopes</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">`

	if len(store.clients) == 0 {
		html += `<tr><td colspan="4" class="px-6 py-12 text-center text-gray-500">No app registrations</td></tr>`
	} else {
		for _, client := range store.clients {
			html += fmt.Sprintf(`
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">%s</td>
							<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 text-sm text-gray-900">%s</td>
							<td class="px-6 py-4 text-sm text-gray-900">%s</td>
						</tr>`, client.Name, client.ClientID, strings.Join(client.RedirectURIs, ", "), strings.Join(client.Scopes, ", "))
		}
	}

	html += fmt.Sprintf(`
					</tbody>
				</table>
			</div>
		</div>
	</div>

	<!-- Settings Tab -->
	<div id="settings-content" class="tab-content">
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="px-6 py-4 bg-gray-50 border-b border-gray-200">
				<h2 class="text-lg font-medium text-gray-900">Mockzure Settings</h2>
			</div>
			<div class="p-6">
				<div class="space-y-6">
					<div>
						<h3 class="text-sm font-medium text-gray-700 mb-4">Statistics</h3>
						<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
							<div class="border border-gray-200 rounded-lg p-4">
								<div class="text-sm text-gray-500">Total VMs</div>
								<div class="text-2xl font-semibold text-gray-900 mt-1">%d</div>
							</div>
							<div class="border border-gray-200 rounded-lg p-4">
								<div class="text-sm text-gray-500">Running VMs</div>
								<div class="text-2xl font-semibold text-green-600 mt-1">%d</div>
							</div>
							<div class="border border-gray-200 rounded-lg p-4">
								<div class="text-sm text-gray-500">Stopped VMs</div>
								<div class="text-2xl font-semibold text-red-600 mt-1">%d</div>
							</div>
						</div>
					</div>

					<div class="border-t border-gray-200 pt-6">
						<h3 class="text-sm font-medium text-gray-700 mb-4">API Endpoints</h3>
						<div class="bg-gray-50 rounded-lg p-4 space-y-2 text-sm font-mono">
							<div><span class="text-green-600">GET</span> <a href="/mock/azure/resource-groups" class="text-blue-600 hover:underline">/mock/azure/resource-groups</a></div>
							<div><span class="text-green-600">GET</span> <a href="/mock/azure/vms" class="text-blue-600 hover:underline">/mock/azure/vms</a></div>
							<div><span class="text-green-600">GET</span> <a href="/mock/azure/users" class="text-blue-600 hover:underline">/mock/azure/users</a></div>
							<div><span class="text-green-600">GET</span> <a href="/mock/azure/apps" class="text-blue-600 hover:underline">/mock/azure/apps</a></div>
							<div><span class="text-green-600">GET</span> <a href="/mock/azure/stats" class="text-blue-600 hover:underline">/mock/azure/stats</a></div>
						</div>
					</div>

					<div class="border-t border-gray-200 pt-6">
						<h3 class="text-sm font-medium text-gray-700 mb-4">Data Management</h3>
						<div class="flex space-x-4">
							<button onclick="resetData()" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
								Reset to Defaults
							</button>
							<button onclick="clearData()" class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors">
								Clear All Data
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</main>

<script>
	function showTab(tabName) {
		// Hide all tab contents
		document.querySelectorAll('.tab-content').forEach(content => {
			content.classList.remove('active');
		});
		
		// Remove active class from all tab buttons
		document.querySelectorAll('.tab-button').forEach(button => {
			button.classList.remove('border-purple-500', 'text-gray-900');
			button.classList.add('border-transparent', 'text-gray-500');
		});
		
		// Show selected tab content
		document.getElementById(tabName + '-content').classList.add('active');
		
		// Add active class to selected tab button
		const activeButton = document.getElementById(tabName + '-tab');
		activeButton.classList.remove('border-transparent', 'text-gray-500');
		activeButton.classList.add('border-purple-500', 'text-gray-900');
	}

	function performAction(vmName, action) {
		fetch('/mock/azure/vms/' + vmName + '/' + action, {
			method: 'POST'
		})
		.then(response => response.json())
		.then(data => {
			alert(data.message || 'Action completed');
			location.reload();
		})
		.catch(error => {
			alert('Failed to perform action: ' + error);
		});
	}

	function resetData() {
		if (confirm('Reset all data to defaults?')) {
			fetch('/mock/azure/data/reset', { method: 'POST' })
			.then(response => response.json())
			.then(data => {
				alert(data.message);
				location.reload();
			})
			.catch(error => alert('Failed to reset data: ' + error));
		}
	}

	function clearData() {
		if (confirm('Clear all data? This cannot be undone.')) {
			fetch('/mock/azure/data/clear', { method: 'POST' })
			.then(response => response.json())
			.then(data => {
				alert(data.message);
				location.reload();
			})
			.catch(error => alert('Failed to clear data: ' + error));
		}
	}
</script>
</body>
</html>`, len(store.vms), running, stopped)

	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Failed to write HTML response: %v", err)
	}
}

func renderUserSelectionPage(w http.ResponseWriter, r *http.Request, clientID, redirectURI, state, responseType, scope string, store *Store) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Mockzure - Select User</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			display: flex;
			justify-content: center;
			align-items: center;
			min-height: 100vh;
			margin: 0;
			padding: 20px;
		}
		.container {
			background: white;
			border-radius: 12px;
			box-shadow: 0 10px 40px rgba(0,0,0,0.2);
			padding: 40px;
			max-width: 500px;
			width: 100%;
		}
		h1 {
			color: #333;
			margin: 0 0 10px 0;
			font-size: 28px;
			text-align: center;
		}
		.subtitle {
			color: #666;
			text-align: center;
			margin-bottom: 30px;
			font-size: 14px;
		}
		.info {
			background: #f7f7f7;
			border-left: 4px solid #667eea;
			padding: 15px;
			margin-bottom: 25px;
			border-radius: 4px;
		}
		.info strong {
			color: #667eea;
		}
		.user-list {
			display: flex;
			flex-direction: column;
			gap: 12px;
		}
		.user-card {
			border: 2px solid #e0e0e0;
			border-radius: 8px;
			padding: 20px;
			cursor: pointer;
			transition: all 0.3s ease;
			background: white;
		}
		.user-card:hover {
			border-color: #667eea;
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
			transform: translateY(-2px);
		}
		.user-name {
			font-size: 18px;
			font-weight: 600;
			color: #333;
			margin-bottom: 5px;
		}
		.user-email {
			color: #666;
			font-size: 14px;
			margin-bottom: 8px;
		}
		.user-role {
			display: inline-block;
			background: #667eea;
			color: white;
			padding: 4px 12px;
			border-radius: 12px;
			font-size: 12px;
			font-weight: 500;
		}
		.footer {
			text-align: center;
			margin-top: 30px;
			color: #999;
			font-size: 12px;
		}
		.mockzure-badge {
			display: inline-block;
			background: #764ba2;
			color: white;
			padding: 2px 8px;
			border-radius: 4px;
			font-weight: bold;
			font-size: 10px;
			margin-left: 8px;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🔐 Select User <span class="mockzure-badge">MOCKZURE</span></h1>
		<div class="subtitle">Choose a user to sign in to the application</div>
		
		<div class="info">
			<strong>Application:</strong> Sandman<br>
			<strong>Client ID:</strong> ` + clientID + `<br>
			<strong>Scope:</strong> ` + scope + `
		</div>
		
		<div class="user-list">`

	// Add each user from the store
	for _, user := range store.users {
		html += fmt.Sprintf(`
			<div class="user-card" onclick="selectUser('%s')">
				<div class="user-name">%s</div>
				<div class="user-email">%s</div>
				<span class="user-role">%s</span>
			</div>`,
			user.ID,
			user.DisplayName,
			user.UserPrincipalName,
			strings.Join(user.Roles, ", "),
		)
	}

	html += `
		</div>
		
		<div class="footer">
			This is a development mock OAuth server for testing purposes only.
		</div>
	</div>
	
	<script>
		function selectUser(userId) {
			const params = new URLSearchParams(window.location.search);
			params.set('user_id', userId);
			window.location.href = '/oauth2/v2.0/authorize?' + params.toString();
		}
	</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Failed to write HTML response: %v", err)
	}
}

func main() {
	store := &Store{}
	store.init()

	mux := http.NewServeMux()

	// OIDC Discovery
	// OIDC Discovery endpoint handler (reusable)
	oidcDiscoveryHandler := func(w http.ResponseWriter, r *http.Request) {
		iss := baseURL(r)
		doc := map[string]interface{}{
			"issuer":                                iss,
			"authorization_endpoint":                iss + "/oauth2/v2.0/authorize",
			"token_endpoint":                        iss + "/oauth2/v2.0/token",
			"userinfo_endpoint":                     iss + "/oidc/userinfo",
			"response_types_supported":              []string{"code"},
			"id_token_signing_alg_values_supported": []string{"none"},
			"scopes_supported":                      []string{"openid", "profile", "email", "User.Read"},
		}
		if err := encodeJSON(w, doc); err != nil {
			log.Printf("Failed to encode OIDC discovery document: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// OIDC Discovery endpoints (both root and tenant-specific)
	mux.HandleFunc("/.well-known/openid-configuration", oidcDiscoveryHandler)
	// Azure SDK uses tenant-specific path: /{tenant-id}/v2.0/.well-known/openid-configuration
	// We'll handle this with a catch-all that checks the pattern
	mux.HandleFunc("/tenant-id/v2.0/.well-known/openid-configuration", oidcDiscoveryHandler)
	mux.HandleFunc("/common/v2.0/.well-known/openid-configuration", oidcDiscoveryHandler)

	// App registration (JSON)
	mux.HandleFunc("/mock/azure/apps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list := []*RegisteredClient{}
			for _, c := range store.clients {
				list = append(list, c)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": list, "count": len(list)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			var c RegisteredClient
			if err := json.NewDecoder(r.Body).Decode(&c); err != nil || c.ClientID == "" {
				http.Error(w, "invalid client payload", http.StatusBadRequest)
				return
			}
			if c.RedirectURIs == nil {
				c.RedirectURIs = []string{}
			}
			if c.Scopes == nil {
				c.Scopes = []string{"openid", "profile", "email"}
			}
			store.clients[c.ClientID] = &c
			w.WriteHeader(http.StatusCreated)
			if err := encodeJSON(w, c); err != nil {
				log.Printf("Failed to encode client response: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Basic web portal at root with tabbed interface
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderPortalPage(w, store)
	})

	// Resource Groups
	mux.HandleFunc("/mock/azure/resource-groups", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": store.resourceGroups, "count": len(store.resourceGroups)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			var rg ResourceGroup
			if err := json.NewDecoder(r.Body).Decode(&rg); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			if rg.Tags == nil {
				rg.Tags = map[string]string{}
			}
			store.resourceGroups = append(store.resourceGroups, &rg)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(rg); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mock/azure/resource-groups/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/mock/azure/resource-groups/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Resource group name required", http.StatusBadRequest)
			return
		}
		rgName := parts[0]

		switch r.Method {
		case http.MethodGet:
			for _, rg := range store.resourceGroups {
				if rg.Name == rgName {
					// Return resource group with its VMs
					vms := []*MockVM{}
					for _, vm := range store.vms {
						if vm.ResourceGroup == rgName {
							vms = append(vms, vm)
						}
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(map[string]interface{}{
						"resourceGroup": rg,
						"vms":           vms,
					}); err != nil {
						log.Printf("Failed to encode JSON response: %v", err)
					}
					return
				}
			}
			http.Error(w, "Resource group not found", http.StatusNotFound)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Service Accounts (Entra ID)
	mux.HandleFunc("/mock/azure/service-accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": store.serviceAccounts, "count": len(store.serviceAccounts)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			var sa ServiceAccount
			if err := json.NewDecoder(r.Body).Decode(&sa); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			if sa.Permissions == nil {
				sa.Permissions = []ResourceGroupPerm{}
			}
			sa.CreatedDateTime = time.Now()
			store.serviceAccounts = append(store.serviceAccounts, &sa)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(sa); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mock/azure/service-accounts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/mock/azure/service-accounts/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Service account ID required", http.StatusBadRequest)
			return
		}
		saID := parts[0]

		switch r.Method {
		case http.MethodGet:
			for _, sa := range store.serviceAccounts {
				if sa.ID == saID || sa.ApplicationID == saID {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(sa); err != nil {
						log.Printf("Failed to encode JSON response: %v", err)
					}
					return
				}
			}
			http.Error(w, "Service account not found", http.StatusNotFound)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// VMs (with service account authentication support)
	// Azure ARM API endpoint for VMs (Azure SDK uses this path)
	// Pattern: /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/virtualMachines
	mux.HandleFunc("/subscriptions/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a VM list or get request
		if strings.Contains(r.URL.Path, "/providers/Microsoft.Compute/virtualMachines") && r.Method == http.MethodGet {
			log.Printf("DEBUG: ARM VM endpoint called - Path: %s, Query: %s", r.URL.Path, r.URL.RawQuery)

			// Extract resource group and VM name from path
			// Path format: /subscriptions/{id}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines[/{vmName}]
			parts := strings.Split(r.URL.Path, "/")
			resourceGroup := ""
			vmName := ""

			for i, part := range parts {
				if part == "resourceGroups" && i+1 < len(parts) {
					resourceGroup = parts[i+1]
				}
				if part == "virtualMachines" && i+1 < len(parts) && parts[i+1] != "" {
					vmName = parts[i+1]
					log.Printf("DEBUG: ARM - Single VM request for: %s", vmName)
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
							log.Printf("DEBUG: ARM - Returning single VM %s with instance view (status: %s)", vmName, vm.Status)
						}

						// Return single VM (not in array)
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
					// Convert to Azure ARM format with proper nested structure
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

					// Include instance view if requested via $expand=instanceView
					expandParam := r.URL.Query().Get("$expand")
					log.Printf("DEBUG: ARM - Expand parameter: '%s'", expandParam)

					if expandParam == "instanceView" || strings.Contains(r.URL.RawQuery, "expand=instanceView") {
						log.Printf("DEBUG: ARM - Including instance view for VM: %s (status: %s)", vm.Name, vm.Status)

						// Map status to proper PowerState codes
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
					} else {
						log.Printf("DEBUG: ARM - NOT including instance view (expand='%s')", expandParam)
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

			// Return in Azure ARM format
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

	mux.HandleFunc("/mock/azure/vms", func(w http.ResponseWriter, r *http.Request) {
		// Try to authenticate as service account
		serviceAccount, _ := store.authenticateServiceAccount(r)

		switch r.Method {
		case http.MethodGet:
			// If authenticated as service account, filter VMs based on permissions
			if serviceAccount != nil {
				filteredVMs := []*MockVM{}
				for _, vm := range store.vms {
					if serviceAccount.hasPermission(vm.ResourceGroup, "read") {
						filteredVMs = append(filteredVMs, vm)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": filteredVMs, "count": len(filteredVMs)}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
				return
			}

			// No authentication or regular user - return all VMs
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": store.vms, "count": len(store.vms)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			var vm MockVM
			if err := json.NewDecoder(r.Body).Decode(&vm); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			if vm.Tags == nil {
				vm.Tags = map[string]string{}
			}
			if vm.PowerState == "" {
				vm.PowerState = "VM deallocated"
			}
			if vm.Status == "" {
				vm.Status = "stopped"
			}
			vm.LastUpdated = time.Now()
			store.vms = append(store.vms, &vm)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(vm); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mock/azure/vms/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/mock/azure/vms/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "VM name required", http.StatusBadRequest)
			return
		}
		vmName := parts[0]

		// Try to authenticate as service account
		serviceAccount, _ := store.authenticateServiceAccount(r)

		// Helper to find VM
		find := func(name string) *MockVM {
			for _, v := range store.vms {
				if v.Name == name {
					return v
				}
			}
			return nil
		}

		// Helper to check permission
		checkPermission := func(vm *MockVM, permission string) bool {
			if serviceAccount == nil {
				return true // No auth required for non-service accounts (backward compat)
			}
			return serviceAccount.hasPermission(vm.ResourceGroup, permission)
		}

		switch r.Method {
		case http.MethodGet:
			vm := find(vmName)
			if vm == nil {
				http.Error(w, "VM not found", http.StatusNotFound)
				return
			}

			// Check read permission
			if !checkPermission(vm, "read") {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			if len(parts) > 1 && parts[1] == "status" {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{
					"status":      vm.Status,
					"powerState":  vm.PowerState,
					"lastUpdated": vm.LastUpdated,
				}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(vm); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			if len(parts) < 2 {
				http.Error(w, "Operation required", http.StatusBadRequest)
				return
			}
			vm := find(vmName)
			if vm == nil {
				http.Error(w, "VM not found", http.StatusNotFound)
				return
			}
			switch parts[1] {
			case "start":
				// Check start permission
				if !checkPermission(vm, "start") {
					http.Error(w, "Forbidden: insufficient permissions to start VM", http.StatusForbidden)
					return
				}
				vm.Status = "running"
				vm.PowerState = "VM running"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s started successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			case "stop":
				// Check stop permission
				if !checkPermission(vm, "stop") {
					http.Error(w, "Forbidden: insufficient permissions to stop VM", http.StatusForbidden)
					return
				}
				vm.Status = "stopped"
				vm.PowerState = "VM deallocated"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s stopped successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			case "restart":
				// Check restart permission
				if !checkPermission(vm, "restart") {
					http.Error(w, "Forbidden: insufficient permissions to restart VM", http.StatusForbidden)
					return
				}
				vm.Status = "running"
				vm.PowerState = "VM running"
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("VM %s restarted successfully", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			case "tags":
				var tags map[string]*string
				if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}
				vm.Tags = map[string]string{}
				for k, v := range tags {
					if v != nil {
						vm.Tags[k] = *v
					}
				}
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("Tags updated for VM %s", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			default:
				http.Error(w, "Unknown operation", http.StatusBadRequest)
			}
		case http.MethodPut:
			if len(parts) > 1 && parts[1] == "tags" {
				vm := find(vmName)
				if vm == nil {
					http.Error(w, "VM not found", http.StatusNotFound)
					return
				}
				var tags map[string]*string
				if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}
				vm.Tags = map[string]string{}
				for k, v := range tags {
					if v != nil {
						vm.Tags[k] = *v
					}
				}
				vm.LastUpdated = time.Now()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": fmt.Sprintf("Tags updated for VM %s", vmName), "status": "success"}); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
			} else {
				http.Error(w, "Operation required", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Users - requires User.Read.All permission
	mux.HandleFunc("/mock/azure/users", func(w http.ResponseWriter, r *http.Request) {
		// Check service account authentication and Graph API permissions
		serviceAccount, err := store.authenticateServiceAccount(r)
		if err != nil || serviceAccount == nil {
			http.Error(w, `{"error":"unauthorized","error_description":"Authentication required to access Microsoft Graph API"}`, http.StatusUnauthorized)
			return
		}

		// Check if service account has User.Read.All permission
		hasPermission := false
		for _, perm := range serviceAccount.GraphPermissions {
			if perm == "User.Read.All" || perm == "Directory.Read.All" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			http.Error(w, `{"error":"forbidden","error_description":"Insufficient privileges to access user directory. Requires User.Read.All or Directory.Read.All permission"}`, http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"value": store.users, "count": len(store.users)}); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		case http.MethodPost:
			var user MockUser
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			if user.UserType == "" {
				user.UserType = "Member"
			}
			if user.Roles == nil {
				user.Roles = []string{"User"}
			}
			store.users = append(store.users, &user)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(user); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/mock/azure/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/mock/azure/users/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		userID := parts[0]

		switch r.Method {
		case http.MethodGet:
			for _, u := range store.users {
				if u.ID == userID {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(u); err != nil {
						log.Printf("Failed to encode JSON response: %v", err)
					}
					return
				}
			}
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Entra ID
	// Legacy alias authorize
	mux.HandleFunc("/mock/azure/entra/authorize", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/oauth2/v2.0/authorize"
		mux.ServeHTTP(w, r)
	})

	// OIDC Authorize (code flow) - Show user selection page
	mux.HandleFunc("/oauth2/v2.0/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		clientID := q.Get("client_id")
		redirectURI := q.Get("redirect_uri")
		state := q.Get("state")
		responseType := q.Get("response_type")
		scope := q.Get("scope")
		selectedUser := q.Get("user_id") // Check if user was selected

		if clientID == "" || redirectURI == "" || responseType != "code" {
			http.Error(w, "invalid authorize request", http.StatusBadRequest)
			return
		}
		if c, ok := store.clients[clientID]; ok {
			// validate redirect
			valid := len(c.RedirectURIs) == 0
			for _, ru := range c.RedirectURIs {
				if ru == redirectURI {
					valid = true
					break
				}
			}
			if !valid {
				http.Error(w, "unauthorized redirect_uri", http.StatusBadRequest)
				return
			}
		}

		// If user hasn't been selected yet, show the user selection page
		if selectedUser == "" {
			renderUserSelectionPage(w, r, clientID, redirectURI, state, responseType, scope, store)
			return
		}

		// User was selected, create auth code and redirect
		code := fmt.Sprintf("code_%d", time.Now().UnixNano())
		store.codes[code] = &AuthCode{
			Code:        code,
			ClientID:    clientID,
			RedirectURI: redirectURI,
			Scope:       scope,
			UserSub:     selectedUser,
			IssuedAt:    time.Now(),
		}
		u, err := url.Parse(redirectURI)
		if err != nil {
			http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
			return
		}
		qq := u.Query()
		qq.Set("code", code)
		if state != "" {
			qq.Set("state", state)
		}
		u.RawQuery = qq.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
	})

	// Legacy alias token
	mux.HandleFunc("/mock/azure/entra/token", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/oauth2/v2.0/token"
		mux.ServeHTTP(w, r)
	})

	// OIDC Token endpoint (form-encoded)
	mux.HandleFunc("/oauth2/v2.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// support x-www-form-urlencoded
		if ct := r.Header.Get("Content-Type"); strings.Contains(ct, "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}

			grantType := r.Form.Get("grant_type")

			// Client Credentials Flow (for Azure SDK / Service Accounts)
			if grantType == "client_credentials" {
				clientID := r.Form.Get("client_id")
				clientSecret := r.Form.Get("client_secret")
				scope := r.Form.Get("scope")

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

				// Return access token for service account
				token := map[string]interface{}{
					"access_token": "mock_access_token_" + clientID,
					"token_type":   "Bearer",
					"expires_in":   3600,
					"scope":        scope,
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(token); err != nil {
					log.Printf("Failed to encode JSON response: %v", err)
				}
				return
			}

			// Authorization Code Flow (for user login)
			code := r.Form.Get("code")
			if code == "" {
				http.Error(w, "code or grant_type required", http.StatusBadRequest)
				return
			}
			ac, ok := store.codes[code]
			if !ok {
				http.Error(w, "invalid code", http.StatusBadRequest)
				return
			}
			delete(store.codes, code)
			// build id_token - look up user from store
			iss := baseURL(r)
			var email, name, givenName, familyName string = "unknown@dev.local", "Unknown User", "Unknown", "User"

			// Find the user in the store
			for _, user := range store.users {
				if user.ID == ac.UserSub {
					email = user.UserPrincipalName
					name = user.DisplayName
					// Parse given/family names from display name
					nameParts := strings.Fields(user.DisplayName)
					if len(nameParts) > 0 {
						givenName = nameParts[0]
					}
					if len(nameParts) > 1 {
						familyName = strings.Join(nameParts[1:], " ")
					}
					break
				}
			}

			claims := map[string]interface{}{
				"iss":         iss,
				"aud":         ac.ClientID,
				"sub":         ac.UserSub,
				"email":       email,
				"name":        name,
				"given_name":  givenName,
				"family_name": familyName,
				"iat":         time.Now().Unix(),
				"exp":         time.Now().Add(1 * time.Hour).Unix(),
			}
			idt := makeUnsignedJWT(claims)
			token := map[string]interface{}{
				"access_token":  "mock_access_token_" + code,
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "mock_refresh_token_" + code,
				"scope":         ac.Scope,
				"id_token":      idt,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(token); err != nil {
				log.Printf("Failed to encode JSON response: %v", err)
			}
			return
		}
		// fallback: JSON body with {code}
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
			http.Error(w, "Authorization code required", http.StatusBadRequest)
			return
		}
		ac, ok := store.codes[req.Code]
		if !ok {
			http.Error(w, "invalid code", http.StatusBadRequest)
			return
		}
		delete(store.codes, req.Code)
		iss := baseURL(r)

		// Look up user from store
		var email, name, givenName, familyName string = "unknown@dev.local", "Unknown User", "Unknown", "User"
		for _, user := range store.users {
			if user.ID == ac.UserSub {
				email = user.UserPrincipalName
				name = user.DisplayName
				nameParts := strings.Fields(user.DisplayName)
				if len(nameParts) > 0 {
					givenName = nameParts[0]
				}
				if len(nameParts) > 1 {
					familyName = strings.Join(nameParts[1:], " ")
				}
				break
			}
		}

		idt := makeUnsignedJWT(map[string]interface{}{"iss": iss, "aud": ac.ClientID, "sub": ac.UserSub, "email": email, "name": name, "given_name": givenName, "family_name": familyName, "iat": time.Now().Unix(), "exp": time.Now().Add(1 * time.Hour).Unix()})
		token := map[string]interface{}{"access_token": "mock_access_token_" + req.Code, "token_type": "Bearer", "expires_in": 3600, "refresh_token": "mock_refresh_token_" + req.Code, "scope": ac.Scope, "id_token": idt}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(token); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	// Legacy alias userinfo
	mux.HandleFunc("/mock/azure/entra/userinfo", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/oidc/userinfo"
		mux.ServeHTTP(w, r)
	})

	// OIDC userinfo
	mux.HandleFunc("/oidc/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		token := parts[1]
		// trivial mapping for demo; return first user or an admin
		var info MockUserInfo
		if strings.Contains(token, "admin") {
			info = MockUserInfo{
				Sub:               "admin-user-12345",
				Name:              "Admin User",
				Email:             "admin@dev.local",
				GivenName:         "Admin",
				FamilyName:        "User",
				JobTitle:          "System Administrator",
				Department:        "IT",
				OfficeLocation:    "Headquarters",
				Roles:             []string{"Global Administrator", "VM Administrator"},
				AccountEnabled:    true,
				UserPrincipalName: "admin@dev.local",
			}
		} else if len(store.users) > 0 {
			u := store.users[0]
			names := strings.Split(u.DisplayName, " ")
			gn, fn := u.DisplayName, ""
			if len(names) > 1 {
				gn, fn = names[0], names[1]
			}
			info = MockUserInfo{
				Sub:               u.ID,
				Name:              u.DisplayName,
				Email:             u.Mail,
				GivenName:         gn,
				FamilyName:        fn,
				JobTitle:          u.JobTitle,
				Department:        u.Department,
				OfficeLocation:    u.OfficeLocation,
				Roles:             u.Roles,
				AccountEnabled:    u.AccountEnabled,
				UserPrincipalName: u.UserPrincipalName,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	// Stats and data management
	mux.HandleFunc("/mock/azure/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		running, stopped := 0, 0
		for _, v := range store.vms {
			if v.Status == "running" {
				running++
			} else {
				stopped++
			}
		}
		stats := map[string]interface{}{
			"total_vms":   len(store.vms),
			"running_vms": running,
			"stopped_vms": stopped,
			"total_users": len(store.users),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	mux.HandleFunc("/mock/azure/data/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		store.vms = []*MockVM{}
		store.users = []*MockUser{}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Mock data cleared successfully", "status": "success"}); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	mux.HandleFunc("/mock/azure/data/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		store.vms = nil
		store.users = nil
		store.init()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Mock data reset to defaults successfully", "status": "success"}); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	})

	addr := ":8090"
	log.Printf("Starting Mockzure on %s", addr)
	srv := &http.Server{Addr: addr, Handler: mux}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Mockzure failed to start: %v", err)
	}
}
