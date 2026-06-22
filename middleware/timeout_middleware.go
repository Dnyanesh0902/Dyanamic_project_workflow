package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"project-workflow-backend/util"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		// Use a goroutine to execute the main handler
		done := make(chan bool)
		go func() {
			c.Next()
			done <- true
		}()
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				util.GatewayTimeoutAbortWithJSON(c, "Request Timeout")
			}
		case <-done:
			// Request completed before timeout
		}
	}
}

// Middleware function to check if the request is coming from Postman
func PostmanCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		choiceKey := c.GetHeader("x-choice-key")
		if strings.Contains(strings.ToLower(userAgent), "postman") && choiceKey == "" {
			util.UnauthorizedAbortWithJSON(c, "Unauthorized access from Postman")
			return
		}
		c.Next()
	}
}

var ExcludedLogRoutes = []string{
	//campaign route
	"/campaign/details",
	"/campaign/daily-count",
	"/campaign/status-counts",
	"/campaign/stats",
	"/campaign/counts",
	"/campaign/camp-procedure-list",
	"/campaign/count",
	"/campaign/day-wise-count",
	"/campaign/camp-type-count",
	"/campaign//reason-action/list",
	"/pending-action-listing",
	//beneficiary route
	"/beneficiary/list",
	"/beneficiary/camp-produre-count",
	"/beneficiary/activity-logs",
	"/beneficiary/comment-list",
	"/beneficiary/document-list",
	"/beneficiary/token-no-dropdown",
	"/beneficiary/details",
	"/beneficiary/count",
	"/beneficiary/camp-details",

	//user route
	"/user/list",
	"/user/details",
	"/user/count",

	//menu route
	"/menu/details",
	"/menu/listing",

	//menu-permission route
	"menu-permission/details",
	"menu-permission/listing",

	//roles route
	"roles/list",
	"roles/roletype-list",

	//user-type route
	"user-type/list",

	//doctor-consultations route
	"doctor-consultations/list",

	//guest route
	"/guest/list",

	//schemes route
	"/schemes/list",
	"/schemes/listing",

	// stemi route
	"/stemi/list",

	//pharmacy-consultation route
	"/pharmacy-consultation/list",

	//medicine route
	"/medicine/list",
	"/medicine/medicine-list",
	"/medicine/quantity-list",

	//abha-application route
	"/abha-application/list",

	//scheme-refer route
	"/scheme-refer/list",

	//escalation-action route
	"/escalation-action/list",

	//hospital route
	"/hospital/list",

	//referral-letter route
	"/referral-letter/list",

	//beneficiary-followup route
	"/beneficiary-followup/list",

	//diseases route
	"/diseases/list",

	//blood-test route
	"/blood-test/list",
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip routes defined in config
		for _, route := range ExcludedLogRoutes {
			if strings.HasPrefix(c.Request.RequestURI, route) {
				c.Next()
				return
			}
		}

		method := c.Request.Method
		contentType := c.ContentType()
		start := time.Now()
		routePath := c.FullPath()
		timestamp := time.Now().Format("2006-01-02T15:04:05")

		var requestBody string
		var formData map[string][]string
		var fileSummaries []string

		// Logging data collection
		if method == http.MethodGet && GetBoolEnv("REQUEST_LOG_GET", false) {
			formData = c.Request.URL.Query()
		}

		if method == http.MethodPost && GetBoolEnv("REQUEST_LOG_POST", false) {
			if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "application/octet-stream") && GetBoolEnv("REQUEST_LOG_JSON", false) {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				var compactBody bytes.Buffer
				if err := json.Compact(&compactBody, bodyBytes); err == nil {
					requestBody = compactBody.String()
				} else {
					requestBody = string(bodyBytes) // fallback if compacting fails
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else if strings.Contains(contentType, "application/x-www-form-urlencoded") && GetBoolEnv("REQUEST_LOG_FORM", false) {
				_ = c.Request.ParseForm()
				formData = c.Request.PostForm
			} else if strings.Contains(contentType, "multipart/form-data") && GetBoolEnv("REQUEST_LOG_MULTIPART", false) {
				_ = c.Request.ParseMultipartForm(10 << 20)
				formData = c.Request.PostForm
				if c.Request.MultipartForm != nil {
					for key, files := range c.Request.MultipartForm.File {
						for _, fileHeader := range files {
							fileSummaries = append(fileSummaries, fmt.Sprintf("Field: %s, Filename: %s, Size: %d", key, fileHeader.Filename, fileHeader.Size))
							if GetBoolEnv("REQUEST_LOG_FILE_BLOB", false) {
								file, err := fileHeader.Open()
								if err == nil {
									defer file.Close()
									fileBytes, err := io.ReadAll(file)
									if err == nil {
										// Log first 200 bytes preview (hex)
										previewLen := len(fileBytes)
										if previewLen > 200 {
											previewLen = 200
										}
										fileSummaries = append(fileSummaries, fmt.Sprintf("Blob Preview (first %d bytes): %x", previewLen, fileBytes[:previewLen]))
									}
								}
							}
						}
					}
				}
			}
		}

		c.Next()
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		var formDataStr string
		if len(formData) > 0 {
			// Flatten map[string][]string to map[string]string for logging
			flatMap := make(map[string]string)
			for key, values := range formData {
				if len(values) > 0 {
					flatMap[key] = values[0]
				}
			}
			formDataBytes, _ := json.Marshal(flatMap)
			formDataStr = string(formDataBytes)
		}

		var fileSummariesStr string
		if len(fileSummaries) > 0 {
			fileSummariesBytes, _ := json.Marshal(fileSummaries)
			fileSummariesStr = string(fileSummariesBytes)
		}

		logLine := fmt.Sprintf(
			"Route: %s | Method: %s | Duration: %s | Status Code: %d | Headers: %v",
			routePath, method, duration, statusCode, c.Request.Header,
		)

		if formDataStr != "" {
			logLine += " | Query/Form: " + formDataStr
		}
		if requestBody != "" {
			logLine += " | JSON Body: " + requestBody
		}
		if fileSummariesStr != "" {
			logLine += " | Files: " + fileSummariesStr
		}

		logrus.Infof("[Employee REQUEST LOG - %s] | %s", timestamp, logLine)
	}
}

func GetBoolEnv(key string, defaultVal bool) bool {
	val := strings.ToLower(os.Getenv(key))
	if val == "true" {
		return true
	} else if val == "false" {
		return false
	}
	return defaultVal
}
