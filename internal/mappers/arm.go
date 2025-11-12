package mappers

import (
	"fmt"
	"strings"
)

// MapARMResponse maps store data to ARM API response format
func MapARMResponse(operationID, pathPattern, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	// Handle different ARM operations based on operation ID and path pattern
	pathLower := strings.ToLower(pathPattern)

	// Resource Groups operations
	if strings.Contains(pathLower, "resourcegroups") {
		return mapResourceGroupsResponse(operationID, method, params, store)
	}

	// Virtual Machines operations
	if strings.Contains(pathLower, "virtualmachines") || strings.Contains(pathLower, "virtualmachines") {
		return mapVirtualMachinesResponse(operationID, method, params, store)
	}

	// Operations list
	if strings.Contains(pathLower, "/operations") {
		return mapOperationsResponse(operationID, method, params)
	}

	// Default: return empty response
	return map[string]interface{}{"value": []interface{}{}}, nil
}

// mapResourceGroupsResponse handles resource group operations
func mapResourceGroupsResponse(operationID, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	resourceGroups := store.GetResourceGroups()

	switch method {
	case "GET":
		// List resource groups or get specific one
		rgName := params["resourceGroupName"]
		if rgName != "" {
			// Get specific resource group
			for _, rg := range resourceGroups {
				// Type assert to get name (this will need to be adapted based on actual types)
				if rgMap, ok := rg.(map[string]interface{}); ok {
					if name, ok := rgMap["name"].(string); ok && name == rgName {
						return rg, nil
					}
				}
			}
			return nil, fmt.Errorf("resource group not found: %s", rgName)
		}

		// List all resource groups
		return map[string]interface{}{
			"value": resourceGroups,
		}, nil

	case "POST", "PUT":
		// Create or update resource group
		// This would typically parse the request body
		// For now, return a success response
		return map[string]interface{}{
			"id":       fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", params["subscriptionId"], params["resourceGroupName"]),
			"name":     params["resourceGroupName"],
			"location": params["location"],
		}, nil

	case "DELETE":
		// Delete resource group
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// mapVirtualMachinesResponse handles virtual machine operations
func mapVirtualMachinesResponse(operationID, method string, params map[string]string, store StoreInterface) (interface{}, error) {
	vms := store.GetVMs()
	vmName := params["vmName"]
	resourceGroup := params["resourceGroupName"]

	switch method {
	case "GET":
		if vmName != "" {
			// Get specific VM
			for _, vm := range vms {
				if vmMap, ok := vm.(map[string]interface{}); ok {
					if name, ok := vmMap["name"].(string); ok && name == vmName {
						// Check resource group match if specified
						if resourceGroup != "" {
							if rg, ok := vmMap["resourceGroup"].(string); ok && rg != resourceGroup {
								continue
							}
						}
						return convertVMToARMFormat(vmMap), nil
					}
				}
			}
			return nil, fmt.Errorf("virtual machine not found: %s", vmName)
		}

		// List VMs
		filteredVMs := []interface{}{}
		for _, vm := range vms {
			if vmMap, ok := vm.(map[string]interface{}); ok {
				if resourceGroup != "" {
					if rg, ok := vmMap["resourceGroup"].(string); ok && rg == resourceGroup {
						filteredVMs = append(filteredVMs, convertVMToARMFormat(vmMap))
					}
				} else {
					filteredVMs = append(filteredVMs, convertVMToARMFormat(vmMap))
				}
			}
		}

		return map[string]interface{}{
			"value": filteredVMs,
		}, nil

	case "POST":
		// VM actions (start, stop, restart) or create
		if strings.Contains(operationID, "Start") || strings.Contains(operationID, "start") {
			return map[string]interface{}{
				"status": "Succeeded",
			}, nil
		}
		if strings.Contains(operationID, "Deallocate") || strings.Contains(operationID, "stop") {
			return map[string]interface{}{
				"status": "Succeeded",
			}, nil
		}
		if strings.Contains(operationID, "Restart") || strings.Contains(operationID, "restart") {
			return map[string]interface{}{
				"status": "Succeeded",
			}, nil
		}
		// Create VM
		return map[string]interface{}{
			"id":       fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", params["subscriptionId"], resourceGroup, vmName),
			"name":     vmName,
			"type":     "Microsoft.Compute/virtualMachines",
			"location": params["location"],
		}, nil

	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// convertVMToARMFormat converts a VM from internal format to ARM API format
func convertVMToARMFormat(vm map[string]interface{}) map[string]interface{} {
	armVM := map[string]interface{}{
		"id":       vm["id"],
		"name":     vm["name"],
		"type":     "Microsoft.Compute/virtualMachines",
		"location": vm["location"],
		"tags":     vm["tags"],
	}

	// Build properties object
	properties := map[string]interface{}{
		"vmId":              vm["id"],
		"provisioningState": vm["provisioningState"],
		"hardwareProfile": map[string]interface{}{
			"vmSize": vm["vmSize"],
		},
		"storageProfile": map[string]interface{}{
			"osDisk": map[string]interface{}{
				"osType": vm["osType"],
			},
		},
	}

	// Add instance view if status is available
	if status, ok := vm["status"].(string); ok {
		powerStateCode := "PowerState/" + status
		if status == "stopped" {
			powerStateCode = "PowerState/deallocated"
		}

		properties["instanceView"] = map[string]interface{}{
			"statuses": []map[string]interface{}{
				{
					"code":          powerStateCode,
					"level":         "Info",
					"displayStatus": vm["powerState"],
				},
				{
					"code":          "ProvisioningState/" + fmt.Sprintf("%v", vm["provisioningState"]),
					"level":         "Info",
					"displayStatus": "Provisioning " + strings.ToLower(fmt.Sprintf("%v", vm["provisioningState"])),
				},
			},
		}
	}

	armVM["properties"] = properties
	return armVM
}

// mapOperationsResponse handles operations list
func mapOperationsResponse(operationID, method string, params map[string]string) (interface{}, error) {
	// Return list of available operations
	return map[string]interface{}{
		"value": []map[string]interface{}{
			{
				"name": "Microsoft.Resources/ResourceGroups/read",
			},
			{
				"name": "Microsoft.Resources/ResourceGroups/write",
			},
			{
				"name": "Microsoft.Compute/virtualMachines/read",
			},
			{
				"name": "Microsoft.Compute/virtualMachines/write",
			},
		},
	}, nil
}

// MapARMOperationStatus handles ARM Long Running Operation (LRO) status checks
func MapARMOperationStatus(operationID string, params map[string]string) (interface{}, error) {
	// Return operation status for LRO pattern
	// In Azure, operations return an operation ID that can be polled
	return map[string]interface{}{
		"status": "Succeeded",
		"id":     params["operationId"],
	}, nil
}

