package mappers

import (
	"fmt"
	"log"
	"strings"
)

// MapGraphResponse maps store data to Microsoft Graph API response format
func MapGraphResponse(operationID, pathPattern, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	if store == nil {
		return nil, fmt.Errorf("store is nil")
	}

	pathLower := strings.ToLower(pathPattern)

	// Users operations
	if strings.Contains(pathLower, "/users") {
		return mapUsersResponse(operationID, method, params, store)
	}

	// Service Principals operations
	if strings.Contains(pathLower, "/serviceprincipals") || strings.Contains(pathLower, "/servicePrincipals") {
		return mapServicePrincipalsResponse(operationID, method, params, store)
	}

	// Default: return empty response
	return map[string]interface{}{"value": []interface{}{}}, nil
}

// mapUsersResponse handles Microsoft Graph users operations
func mapUsersResponse(operationID, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	if store == nil {
		return nil, fmt.Errorf("store is nil in mapUsersResponse")
	}

	users := store.GetUsers()
	if users == nil {
		log.Printf("Warning: GetUsers() returned nil, using empty slice")
		users = []interface{}{} // Ensure users is never nil
	}

	log.Printf("mapUsersResponse: processing %d users", len(users))

	// Graph API uses {user-id} as parameter name in specs
	userID := params["user-id"]
	if userID == "" {
		userID = params["id"] // Fallback for compatibility
	}

	switch method {
	case "GET":
		if userID != "" {
			// Get specific user
			for _, user := range users {
				if userMap, ok := user.(map[string]interface{}); ok {
					if id, ok := userMap["id"].(string); ok && id == userID {
						return convertUserToGraphFormat(userMap), nil
					}
					// Also check userPrincipalName
					if upn, ok := userMap["userPrincipalName"].(string); ok && upn == userID {
						return convertUserToGraphFormat(userMap), nil
					}
				}
			}
			return nil, fmt.Errorf("user not found: %s", userID)
		}

		// List users with pagination support
		graphUsers := []interface{}{}
		for i, user := range users {
			if user == nil {
				log.Printf("Warning: user at index %d is nil, skipping", i)
				continue
			}
			userMap, ok := user.(map[string]interface{})
			if !ok {
				log.Printf("Warning: user at index %d is not a map[string]interface{}, got %T, skipping", i, user)
				continue
			}
			if userMap == nil {
				log.Printf("Warning: user map at index %d is nil, skipping", i)
				continue
			}
			converted := convertUserToGraphFormat(userMap)
			if converted != nil {
				graphUsers = append(graphUsers, converted)
			}
		}

		log.Printf("mapUsersResponse: converted %d users to Graph format", len(graphUsers))

		response := map[string]interface{}{
			"value": graphUsers,
		}

		// Handle Graph API pagination
		// Support $top, $skip, $count parameters
		allUsers := graphUsers
		if top, ok := params["$top"]; ok && top != "" {
			// Limit results (simplified - would need proper parsing)
			// For now, return all
		}

		// Add @odata.context for Graph API compliance
		response["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#users"

		// Add @odata.nextLink if there are more results (simplified)
		// In a real implementation, this would be based on $top and $skip
		if len(allUsers) > 0 {
			// For now, don't add nextLink (all results returned)
		}

		return response, nil

	case "POST":
		// Create user
		// This would typically parse the request body
		// For now, return a success response
		return map[string]interface{}{
			"id":                userID,
			"userPrincipalName": params["userPrincipalName"],
			"displayName":       params["displayName"],
		}, nil

	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// convertUserToGraphFormat converts a user from internal format to Graph API format
func convertUserToGraphFormat(user map[string]interface{}) map[string]interface{} {
	if user == nil {
		log.Printf("Warning: convertUserToGraphFormat called with nil user map")
		return map[string]interface{}{}
	}

	graphUser := make(map[string]interface{})

	// Safely extract fields, handling nil values
	if id, ok := user["id"]; ok {
		graphUser["id"] = id
	}
	if displayName, ok := user["displayName"]; ok {
		graphUser["displayName"] = displayName
	}
	if userPrincipalName, ok := user["userPrincipalName"]; ok {
		graphUser["userPrincipalName"] = userPrincipalName
	}
	if mail, ok := user["mail"]; ok {
		graphUser["mail"] = mail
	}
	if jobTitle, ok := user["jobTitle"]; ok {
		graphUser["jobTitle"] = jobTitle
	}
	if department, ok := user["department"]; ok {
		graphUser["department"] = department
	}
	if officeLocation, ok := user["officeLocation"]; ok {
		graphUser["officeLocation"] = officeLocation
	}
	if userType, ok := user["userType"]; ok {
		graphUser["userType"] = userType
	}
	if accountEnabled, ok := user["accountEnabled"]; ok {
		graphUser["accountEnabled"] = accountEnabled
	}

	// Add @odata.context if needed
	// graphUser["@odata.context"] = "https://graph.microsoft.com/v1.0/$metadata#users/$entity"

	return graphUser
}

// mapServicePrincipalsResponse handles Microsoft Graph service principals operations
func mapServicePrincipalsResponse(operationID, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	serviceAccounts := store.GetServiceAccounts()
	// Graph API uses {servicePrincipal-id} as parameter name in specs
	spID := params["servicePrincipal-id"]
	if spID == "" {
		spID = params["id"] // Fallback for compatibility
	}

	switch method {
	case "GET":
		if spID != "" {
			// Get specific service principal
			for _, sa := range serviceAccounts {
				if saMap, ok := sa.(map[string]interface{}); ok {
					if id, ok := saMap["id"].(string); ok && id == spID {
						return convertServiceAccountToGraphFormat(saMap), nil
					}
					if appID, ok := saMap["applicationId"].(string); ok && appID == spID {
						return convertServiceAccountToGraphFormat(saMap), nil
					}
				}
			}
			return nil, fmt.Errorf("service principal not found: %s", spID)
		}

		// List service principals
		graphSPs := []interface{}{}
		for _, sa := range serviceAccounts {
			if saMap, ok := sa.(map[string]interface{}); ok {
				graphSPs = append(graphSPs, convertServiceAccountToGraphFormat(saMap))
			}
		}

		return map[string]interface{}{
			"value": graphSPs,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// convertServiceAccountToGraphFormat converts a service account to Graph API format
func convertServiceAccountToGraphFormat(sa map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":                   sa["id"],
		"appId":                sa["applicationId"],
		"displayName":          sa["displayName"],
		"description":          sa["description"],
		"accountEnabled":       sa["accountEnabled"],
		"servicePrincipalType": "Application",
	}
}
