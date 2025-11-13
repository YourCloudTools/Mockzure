package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code, response size, and body
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	body         *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Capture the body
	rw.body.Write(b)
	
	// Write to the actual response writer
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// maskSensitiveHeader masks sensitive header values
func maskSensitiveHeader(name, value string) string {
	nameLower := strings.ToLower(name)
	
	// Check if header name contains sensitive keywords
	if strings.Contains(nameLower, "authorization") {
		// Show scheme but mask the actual token
		if strings.HasPrefix(value, "Bearer ") {
			return "Bearer ***"
		}
		if strings.HasPrefix(value, "Basic ") {
			return "Basic ***"
		}
		return "***"
	}
	
	if strings.Contains(nameLower, "secret") ||
		strings.Contains(nameLower, "password") ||
		strings.Contains(nameLower, "token") ||
		strings.Contains(nameLower, "api-key") ||
		strings.Contains(nameLower, "auth-token") {
		return "***"
	}
	
	return value
}

// maskSensitiveQueryParam masks sensitive query parameter values
func maskSensitiveQueryParam(name, value string) string {
	nameLower := strings.ToLower(name)
	
	if nameLower == "client_secret" ||
		nameLower == "password" ||
		nameLower == "token" ||
		nameLower == "secret" ||
		nameLower == "api_key" ||
		nameLower == "auth_token" {
		return "***"
	}
	
	return value
}

// readRequestBody reads and returns the request body, restoring it for handlers
func readRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	
	// Restore the body for handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	
	return body, nil
}

// formatBodyPreview formats request body for logging with size limit
func formatBodyPreview(body []byte, contentType string, maxSize int) string {
	if len(body) == 0 {
		return "(empty)"
	}
	
	// Limit body size for logging
	preview := body
	if len(preview) > maxSize {
		preview = preview[:maxSize]
	}
	
	bodyStr := string(preview)
	if len(body) > maxSize {
		bodyStr += fmt.Sprintf("\n... (truncated, total size: %d bytes)", len(body))
	}
	
	// Try to format JSON if content type suggests it
	if strings.Contains(strings.ToLower(contentType), "json") {
		// Check if it's valid JSON for pretty printing
		var jsonData interface{}
		if err := json.Unmarshal(preview, &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "  ", "  "); err == nil {
				bodyStr = string(prettyJSON)
				if len(body) > maxSize {
					bodyStr += fmt.Sprintf("\n... (truncated, total size: %d bytes)", len(body))
				}
			}
		}
	}
	
	return bodyStr
}

// DebugMiddleware logs all HTTP request and response details
func DebugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// Create response writer wrapper
		rw := newResponseWriter(w)
		
		// Log request details
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("🔍 DEBUG: %s %s", r.Method, r.URL.Path)
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
		// Log full URL
		log.Printf("📍 URL: %s", r.URL.String())
		
		// Log headers (with masking)
		if len(r.Header) > 0 {
			log.Printf("📋 Headers:")
			for name, values := range r.Header {
				for _, value := range values {
					maskedValue := maskSensitiveHeader(name, value)
					log.Printf("   %s: %s", name, maskedValue)
				}
			}
		}
		
		// Log query parameters (with masking)
		if len(r.URL.Query()) > 0 {
			log.Printf("🔗 Query Parameters:")
			for name, values := range r.URL.Query() {
				for _, value := range values {
					maskedValue := maskSensitiveQueryParam(name, value)
					log.Printf("   %s = %s", name, maskedValue)
				}
			}
		}
		
		// Log request body
		contentType := r.Header.Get("Content-Type")
		body, err := readRequestBody(r)
		if err != nil {
			log.Printf("⚠️  Error reading request body: %v", err)
		} else if len(body) > 0 {
			log.Printf("📦 Request Body (%s, %d bytes):", contentType, len(body))
			bodyPreview := formatBodyPreview(body, contentType, 10240) // 10KB limit
			log.Printf("   %s", bodyPreview)
		} else {
			log.Printf("📦 Request Body: (empty)")
		}
		
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
		// Call the next handler
		next.ServeHTTP(rw, r)
		
		// Calculate duration
		duration := time.Since(startTime)
		
		// Get response headers
		responseHeaders := make(map[string][]string)
		for k, v := range rw.Header() {
			responseHeaders[k] = v
		}
		
		// Log response details
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("✅ RESPONSE: %s %s", r.Method, r.URL.Path)
		log.Printf("   Status: %d", rw.statusCode)
		log.Printf("   Size: %d bytes", rw.bytesWritten)
		log.Printf("   Duration: %d ms", duration.Milliseconds())
		
		// Log response headers
		if len(responseHeaders) > 0 {
			log.Printf("📋 Response Headers:")
			for name, values := range responseHeaders {
				for _, value := range values {
					maskedValue := maskSensitiveHeader(name, value)
					log.Printf("   %s: %s", name, maskedValue)
				}
			}
		}
		
		// Log response body
		responseBody := rw.body.Bytes()
		// Get Content-Type header (case-insensitive)
		contentTypeStr := ""
		for name, values := range responseHeaders {
			if strings.EqualFold(name, "Content-Type") && len(values) > 0 {
				contentTypeStr = values[0]
				break
			}
		}
		
		if len(responseBody) > 0 {
			log.Printf("📦 Response Body (%s, %d bytes):", contentTypeStr, len(responseBody))
			bodyPreview := formatBodyPreview(responseBody, contentTypeStr, 10240) // 10KB limit
			// Split multi-line body preview for better readability
			bodyLines := strings.Split(bodyPreview, "\n")
			for _, line := range bodyLines {
				log.Printf("   %s", line)
			}
		} else {
			log.Printf("📦 Response Body: (empty)")
		}
		
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("")
	})
}

// ValidationMiddleware validates requests against spec schemas
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add schema validation
		// For now, just pass through
		// In the future, validate:
		// - Request body against request schema
		// - Query parameters against parameter definitions
		// - Required parameters are present
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware enforces authentication requirements
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is handled in main.go's existing logic
		// This middleware can be extended for spec-based auth validation
		next.ServeHTTP(w, r)
	})
}
