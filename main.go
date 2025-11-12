package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourcloudtools/mockzure/internal/routes"
	"github.com/yourcloudtools/mockzure/internal/specs"
	yaml "gopkg.in/yaml.v3"
)

// Helper function to encode JSON with error handling
func encodeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// Lightweight replicas of types and behavior from Sandman's internal mock

type ResourceGroup struct {
	ID       string            `json:"id" yaml:"id"`
	Name     string            `json:"name" yaml:"name"`
	Location string            `json:"location" yaml:"location"`
	Tags     map[string]string `json:"tags" yaml:"tags"`
}

type MockVM struct {
	ID                string            `json:"id" yaml:"id"`
	Name              string            `json:"name" yaml:"name"`
	ResourceGroup     string            `json:"resourceGroup" yaml:"resourceGroup"`
	Location          string            `json:"location" yaml:"location"`
	VMSize            string            `json:"vmSize" yaml:"vmSize"`
	OSType            string            `json:"osType" yaml:"osType"`
	ProvisioningState string            `json:"provisioningState" yaml:"provisioningState"`
	PowerState        string            `json:"powerState" yaml:"powerState"`
	Status            string            `json:"status" yaml:"status"`
	LastUpdated       time.Time         `json:"lastUpdated" yaml:"lastUpdated"`
	Tags              map[string]string `json:"tags" yaml:"tags"`
	Owner             string            `json:"owner" yaml:"owner"`
	CostCenter        string            `json:"costCenter" yaml:"costCenter"`
	Environment       string            `json:"environment" yaml:"environment"`
}

type MockAzureRole struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Actions     []string `json:"actions" yaml:"actions"`
	Scope       string   `json:"scope" yaml:"scope"`
}

type MockPermission struct {
	Resource      string   `json:"resource" yaml:"resource"`
	Actions       []string `json:"actions" yaml:"actions"`
	ResourceGroup string   `json:"resourceGroup" yaml:"resourceGroup"`
}

type MockUser struct {
	ID                string           `json:"id" yaml:"id"`
	DisplayName       string           `json:"displayName" yaml:"displayName"`
	UserPrincipalName string           `json:"userPrincipalName" yaml:"userPrincipalName"`
	Mail              string           `json:"mail" yaml:"mail"`
	JobTitle          string           `json:"jobTitle" yaml:"jobTitle"`
	Department        string           `json:"department" yaml:"department"`
	OfficeLocation    string           `json:"officeLocation" yaml:"officeLocation"`
	UserType          string           `json:"userType" yaml:"userType"`
	AccountEnabled    bool             `json:"accountEnabled" yaml:"accountEnabled"`
	Roles             []string         `json:"roles" yaml:"roles"`
	AzureRoles        []MockAzureRole  `json:"azureRoles" yaml:"azureRoles"`
	Permissions       []MockPermission `json:"permissions" yaml:"permissions"`
	ResourceGroups    []string         `json:"resourceGroups" yaml:"resourceGroups"`
	Subscriptions     []string         `json:"subscriptions" yaml:"subscriptions"`
}

// ServiceAccount represents an Azure Service Principal / Service Account
type ServiceAccount struct {
	ID               string              `json:"id" yaml:"id"`
	ApplicationID    string              `json:"applicationId" yaml:"applicationId"` // Client ID
	DisplayName      string              `json:"displayName" yaml:"displayName"`
	Description      string              `json:"description" yaml:"description"`
	AccountEnabled   bool                `json:"accountEnabled" yaml:"accountEnabled"`
	CreatedDateTime  time.Time           `json:"createdDateTime" yaml:"createdDateTime"`
	Permissions      []ResourceGroupPerm `json:"permissions" yaml:"permissions"`
	ServicePrincipal bool                `json:"servicePrincipal" yaml:"servicePrincipal"`
	GraphPermissions []string            `json:"graphPermissions" yaml:"graphPermissions"` // Microsoft Graph API permissions
}

// ResourceGroupPerm represents permissions for a service account on a resource group
type ResourceGroupPerm struct {
	ResourceGroup string   `json:"resourceGroup" yaml:"resourceGroup"` // Resource group name or "*" for all
	Permissions   []string `json:"permissions" yaml:"permissions"`     // "read", "write", "start", "stop", "restart"
}

// ServiceAccountConfig holds the secret configuration for service accounts
type ServiceAccountConfig struct {
	ServiceAccounts []ServiceAccountSecret `json:"serviceAccounts" yaml:"serviceAccounts"`
}

// ServiceAccountSecret holds the secret for a service account
type ServiceAccountSecret struct {
	ApplicationID    string   `json:"applicationId" yaml:"applicationId"`
	Secret           string   `json:"secret" yaml:"secret"`
	DisplayName      string   `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description      string   `json:"description,omitempty" yaml:"description,omitempty"`
	GraphPermissions []string `json:"graphPermissions,omitempty" yaml:"graphPermissions,omitempty"`
}

// FullConfig represents the YAML/JSON configuration file schema
type FullConfig struct {
	ResourceGroups  []*ResourceGroup       `json:"resourceGroups" yaml:"resourceGroups"`
	VMs             []*MockVM              `json:"vms" yaml:"vms"`
	Users           []*MockUser            `json:"users" yaml:"users"`
	ServiceAccounts []FullConfigServiceAcc `json:"serviceAccounts" yaml:"serviceAccounts"`
}

// FullConfigServiceAcc is a service account definition including secret as stored in config
type FullConfigServiceAcc struct {
	ID               string              `json:"id,omitempty" yaml:"id,omitempty"`
	ApplicationID    string              `json:"applicationId" yaml:"applicationId"`
	Secret           string              `json:"secret" yaml:"secret"`
	DisplayName      string              `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	AccountEnabled   bool                `json:"accountEnabled,omitempty" yaml:"accountEnabled,omitempty"`
	CreatedDateTime  time.Time           `json:"createdDateTime,omitempty" yaml:"createdDateTime,omitempty"`
	Permissions      []ResourceGroupPerm `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	ServicePrincipal bool                `json:"servicePrincipal,omitempty" yaml:"servicePrincipal,omitempty"`
	GraphPermissions []string            `json:"graphPermissions,omitempty" yaml:"graphPermissions,omitempty"`
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
	configPath      string
}

// GetResourceGroups returns resource groups as interface slice for mappers
func (s *Store) GetResourceGroups() []interface{} {
	result := make([]interface{}, len(s.resourceGroups))
	for i, rg := range s.resourceGroups {
		result[i] = map[string]interface{}{
			"id":       rg.ID,
			"name":     rg.Name,
			"location": rg.Location,
			"tags":     rg.Tags,
		}
	}
	return result
}

// GetVMs returns VMs as interface slice for mappers
func (s *Store) GetVMs() []interface{} {
	result := make([]interface{}, len(s.vms))
	for i, vm := range s.vms {
		result[i] = map[string]interface{}{
			"id":                vm.ID,
			"name":              vm.Name,
			"resourceGroup":     vm.ResourceGroup,
			"location":          vm.Location,
			"vmSize":            vm.VMSize,
			"osType":            vm.OSType,
			"provisioningState": vm.ProvisioningState,
			"powerState":        vm.PowerState,
			"status":            vm.Status,
			"tags":              vm.Tags,
		}
	}
	return result
}

// GetUsers returns users as interface slice for mappers
func (s *Store) GetUsers() []interface{} {
	result := make([]interface{}, len(s.users))
	for i, user := range s.users {
		result[i] = map[string]interface{}{
			"id":                user.ID,
			"displayName":       user.DisplayName,
			"userPrincipalName": user.UserPrincipalName,
			"mail":              user.Mail,
			"jobTitle":          user.JobTitle,
			"department":        user.Department,
			"officeLocation":    user.OfficeLocation,
			"userType":          user.UserType,
			"accountEnabled":    user.AccountEnabled,
			"roles":             user.Roles,
		}
	}
	return result
}

// GetServiceAccounts returns service accounts as interface slice for mappers
func (s *Store) GetServiceAccounts() []interface{} {
	result := make([]interface{}, len(s.serviceAccounts))
	for i, sa := range s.serviceAccounts {
		result[i] = map[string]interface{}{
			"id":               sa.ID,
			"applicationId":    sa.ApplicationID,
			"displayName":      sa.DisplayName,
			"description":      sa.Description,
			"accountEnabled":   sa.AccountEnabled,
			"servicePrincipal": sa.ServicePrincipal,
		}
	}
	return result
}

func (s *Store) init() {
	// Start empty; load only what is defined in config
	s.resourceGroups = []*ResourceGroup{}
	s.vms = []*MockVM{}
	s.users = []*MockUser{}
	s.serviceAccounts = []*ServiceAccount{}

	// Load from config path (must be set)
	if err := s.loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// app registrations and auth codes
	s.clients = make(map[string]*RegisteredClient)
	s.codes = make(map[string]*AuthCode)
}

// loadConfig loads resources and secrets from the configured file
func (s *Store) loadConfig() error {
	if s.configPath == "" {
		return fmt.Errorf("config path not set")
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var fc FullConfig
	ext := strings.ToLower(filepath.Ext(s.configPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &fc); err != nil {
			return fmt.Errorf("parse json: %w", err)
		}
	default:
		// Try YAML then JSON
		if err := yaml.Unmarshal(data, &fc); err != nil {
			if err2 := json.Unmarshal(data, &fc); err2 != nil {
				return fmt.Errorf("unsupported config format: %v / %v", err, err2)
			}
		}
	}

	// Secrets for auth
	s.config = &ServiceAccountConfig{ServiceAccounts: []ServiceAccountSecret{}}

	// Hydrate resources
	if fc.ResourceGroups != nil {
		s.resourceGroups = fc.ResourceGroups
	}
	if fc.VMs != nil {
		s.vms = fc.VMs
	}
	if fc.Users != nil {
		s.users = fc.Users
	}
	if fc.ServiceAccounts != nil {
		for _, csa := range fc.ServiceAccounts {
			// Build store service account (without secret)
			sa := &ServiceAccount{
				ID:               csa.ID,
				ApplicationID:    csa.ApplicationID,
				DisplayName:      csa.DisplayName,
				Description:      csa.Description,
				AccountEnabled:   csa.AccountEnabled || true,
				CreatedDateTime:  csa.CreatedDateTime,
				Permissions:      csa.Permissions,
				ServicePrincipal: csa.ServicePrincipal || true,
				GraphPermissions: csa.GraphPermissions,
			}
			s.serviceAccounts = append(s.serviceAccounts, sa)
			// Add secret to auth config
			s.config.ServiceAccounts = append(s.config.ServiceAccounts, ServiceAccountSecret{
				ApplicationID:    csa.ApplicationID,
				Secret:           csa.Secret,
				DisplayName:      csa.DisplayName,
				Description:      csa.Description,
				GraphPermissions: csa.GraphPermissions,
			})
		}
	}

	log.Printf("Config loaded: %d RGs, %d VMs, %d users, %d service accounts",
		len(s.resourceGroups), len(s.vms), len(s.users), len(s.serviceAccounts))
	return nil
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
	// Parse command line flags
	var showHelp = flag.Bool("help", false, "Show help information")
	var showVersion = flag.Bool("version", false, "Show version information")
	var configPathFlag = flag.String("config", "", "Path to config file (json|yaml). Can also use MOCKZURE_CONFIG env var")
	flag.Parse()

	// Handle help flag
	if *showHelp {
		fmt.Println("Mockzure - Azure API Mock Server")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  mockzure --config /path/to/config.(json|yaml) [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --config   Path to config file (or set MOCKZURE_CONFIG)")
		fmt.Println("  --help     Show this help message")
		fmt.Println("  --version  Show version information")
		fmt.Println("")
		fmt.Println("Description:")
		fmt.Println("  Mockzure is a mock server that provides Azure-compatible APIs")
		fmt.Println("  for testing and development purposes.")
		fmt.Println("")
		fmt.Println("  A configuration file is required and can be JSON or YAML.")
		fmt.Println("  The server will start on port 8090 by default.")
		fmt.Println("")
		fmt.Println("Endpoints:")
		fmt.Println("  GET  /mock/azure/vms           - List virtual machines")
		fmt.Println("  POST /mock/azure/vms           - Create virtual machine")
		fmt.Println("  GET  /mock/azure/users         - List users")
		fmt.Println("  GET  /mock/azure/stats         - Get server statistics")
		fmt.Println("  POST /mock/azure/data/clear    - Clear all mock data")
		fmt.Println("  POST /mock/azure/data/reset    - Reset to default data")
		os.Exit(0)
	}

	// Handle version flag
	if *showVersion {
		fmt.Println("Mockzure v1.0.0")
		fmt.Println("Azure API Mock Server")
		os.Exit(0)
	}

	// Resolve config path
	cfgPath := *configPathFlag
	if cfgPath == "" {
		cfgPath = os.Getenv("MOCKZURE_CONFIG")
	}
	if cfgPath == "" {
		log.Fatal("config path required via --config or MOCKZURE_CONFIG")
	}

	// Check if config path exists and is a file (not a directory)
	info, err := os.Stat(cfgPath)
	if err != nil {
		log.Fatalf("config file not accessible: %v", err)
	}
	if info.IsDir() {
		log.Fatalf("config path is a directory, not a file: %s (hint: use a file like config.yaml or config.json)", cfgPath)
	}

	store := &Store{configPath: cfgPath}
	store.init()

	mux := http.NewServeMux()

	// Load API specifications and generate routes
	specsDir := "mockzure-specs"
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		log.Printf("Warning: specs directory '%s' not found, skipping spec-driven routes", specsDir)
	} else {
		// Initialize spec loader and registry
		loader := specs.NewLoader(specsDir)
		registry := specs.NewRegistry()

		// Load all specs
		if err := loader.LoadAll(registry); err != nil {
			log.Printf("Warning: Failed to load specs: %v", err)
			log.Printf("Continuing without spec-driven routes")
		} else {
			log.Printf("Loaded API specifications successfully")

			// Generate routes from specs
			routeGen := routes.NewRouteGenerator(store)
			generatedRoutes, err := routeGen.GenerateRoutes(registry)
			if err != nil {
				log.Printf("Warning: Failed to generate routes from specs: %v", err)
			} else {
				log.Printf("Generated %d routes from specifications", len(generatedRoutes))

				// Register spec-driven routes
				// All Azure API endpoints are now generated from specs
				// Remaining hardcoded routes are only for mock-specific functionality (portal, stats, data management)
				routes.RegisterRoutes(mux, generatedRoutes)
			}
		}
	}

	// OIDC Discovery endpoints
	// Note: These are kept as hardcoded handlers because they require custom mock logic
	// (dynamic issuer URL, mock-specific endpoints) that isn't in the OIDC spec.
	// The spec defines the endpoint structure, but the implementation is mock-specific.
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
	mux.HandleFunc("/.well-known/openid-configuration", oidcDiscoveryHandler)
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

	// Resource Groups routes are now generated from ARM specs

	// Service Accounts routes are now generated from Graph API specs

	// VM routes are now generated from ARM specs

	// Users routes are now generated from Graph API specs

	// OIDC/OAuth2 endpoints
	// Note: These are kept as hardcoded handlers because they require custom mock logic
	// (user selection page, token generation, auth code management) that isn't in the OIDC spec.
	// The spec defines the endpoint structure, but the implementation is mock-specific.

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
			var email, name, givenName, familyName = "unknown@dev.local", "Unknown User", "Unknown", "User"

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
		var email, name, givenName, familyName = "unknown@dev.local", "Unknown User", "Unknown", "User"
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

	// OIDC userinfo endpoint (mock-specific implementation)
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
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Mockzure failed to start: %v", err)
	}
}
