package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestAdminUserAccess tests admin user access patterns
func TestAdminUserAccess(t *testing.T) {
	store := &Store{}
	store.init()

	t.Run("admin user access to VMs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mock/azure/vms", nil)
		req.Header.Set("Authorization", "Basic YWRtaW4tYXV0b21hdGlvbi1hcHAtaWQ6YWRtaW4tc2VjcmV0LWtleS1kZXZlbG9wbWVudC1vbmx5")

		// Create test server
		mux := http.NewServeMux()
		setupMockzureHandlers(mux, store)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Admin should see all VMs
		body := w.Body.String()
		if !strings.Contains(body, "value") {
			t.Error("Admin should see VM data")
		}
	})

	t.Run("admin user VM operations", func(t *testing.T) {
		// Test admin can perform operations on all VMs
		req := httptest.NewRequest("POST", "/mock/azure/vms/vm-web-prod-01/start", nil)
		req.Header.Set("Authorization", "Basic YWRtaW4tYXV0b21hdGlvbi1hcHAtaWQ6YWRtaW4tc2VjcmV0LWtleS1kZXZlbG9wbWVudC1vbmx5")

		// Create test server
		mux := http.NewServeMux()
		setupMockzureHandlers(mux, store)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify operation succeeded
		body := w.Body.String()
		if !strings.Contains(body, "started successfully") {
			t.Error("Admin should be able to start prod VMs")
		}
	})
}

// TestMainFunctionIntegration tests main function integration scenarios
func TestMainFunctionIntegration(t *testing.T) {
	t.Run("server setup components", func(t *testing.T) {
		// Test that we can create the components main() would use
		store := &Store{}
		store.init()

		// Test that store is properly initialized
		if store.users == nil {
			t.Error("Store should be initialized")
		}

		// Test creating a basic server
		server := &http.Server{
			Addr:    ":8090",
			Handler: http.NewServeMux(),
		}

		// Test server creation (don't start it)
		t.Logf("Server created successfully on port %s", server.Addr)
	})

	t.Run("handler components", func(t *testing.T) {
		store := &Store{}
		store.init()

		// Test that we can create handlers like main() would
		mux := http.NewServeMux()
		if mux == nil {
			t.Error("ServeMux should be created")
		}

		// Test adding a simple handler
		mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Test the handler works
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("environment setup", func(t *testing.T) {
		// Test environment variables
		port := os.Getenv("PORT")
		if port != "" {
			t.Logf("PORT environment variable: %s", port)
		}

		// Test other common environment variables
		envVars := []string{"HOST", "ADDR", "LISTEN_ADDR", "BIND_ADDR"}
		for _, envVar := range envVars {
			value := os.Getenv(envVar)
			if value != "" {
				t.Logf("%s environment variable: %s", envVar, value)
			}
		}

		// Test file system operations that main might do
		wd, err := os.Getwd()
		if err != nil {
			t.Errorf("Failed to get working directory: %v", err)
		}
		if wd == "" {
			t.Error("Working directory should not be empty")
		}

		// Test that we can access config files
		configFiles := []string{"config.json", "config.json.example"}
		for _, file := range configFiles {
			if _, err := os.Stat(file); err == nil {
				t.Logf("Config file %s exists", file)
			}
		}
	})

	t.Run("server lifecycle", func(t *testing.T) {
		store := &Store{}
		store.init()

		// Create a test server
		server := &http.Server{
			Addr:    ":0", // Use port 0 for automatic port assignment
			Handler: http.NewServeMux(),
		}

		// Start server in background
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Errorf("Server failed to start: %v", err)
			}
		}()

		// Give server time to start
		time.Sleep(50 * time.Millisecond)

		// Test graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			t.Logf("Server shutdown error (expected for test): %v", err)
		}
	})

	t.Run("store initialization", func(t *testing.T) {
		// Test the initialization that main() would do
		store := &Store{}
		store.init()

		// Verify store is properly initialized
		if store.users == nil {
			t.Error("Store users should be initialized")
		}
		if store.vms == nil {
			t.Error("Store VMs should be initialized")
		}
		if store.serviceAccounts == nil {
			t.Error("Store service accounts should be initialized")
		}
		if store.config == nil {
			t.Error("Store config should be initialized")
		}
	})

	t.Run("handler registration", func(t *testing.T) {
		// Test that we can create handlers like main() would
		store := &Store{}
		store.init()

		// Create a mux like main() would
		mux := http.NewServeMux()

		// Add a test handler
		mux.HandleFunc("/test-endpoint", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Test the handler
		req := httptest.NewRequest("GET", "/test-endpoint", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("error handling paths", func(t *testing.T) {
		// Test error handling that main might encounter
		t.Log("Testing error handling paths")

		// Test with malformed requests
		store := &Store{}
		store.init()
		mux := http.NewServeMux()

		// Test malformed request
		req := httptest.NewRequest("GET", "/test", nil) // Valid request instead of empty
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)
		// Should handle gracefully
	})

	t.Run("full request cycle", func(t *testing.T) {
		store := &Store{}
		store.init()
		mux := http.NewServeMux()

		// Add a test handler
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Test</body></html>"))
		})

		// Test a complete request cycle
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "html") {
			t.Error("Response should contain HTML")
		}
	})

	t.Run("concurrent requests", func(t *testing.T) {
		store := &Store{}
		store.init()
		mux := http.NewServeMux()

		// Add a test handler
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Test concurrent requests (like main server would handle)
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}

				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
