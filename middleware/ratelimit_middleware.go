package middleware

import (
	"errors"
	"net"
	"os"
	"project-workflow-backend/util"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var (
	mu            sync.Mutex
	clients       = make(map[string]*RateLimiterClient)
	endpointRates = make(map[string]*rate.Limiter)
)

type RateLimiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

/*
Rate Limit: This specifies the average rate at which requests are allowed. For example, if the rate limit is 2 requests per second, on average, the system will process 2 requests every second.

Burst Limit: This represents the maximum number of requests that can be processed without delay, even if it exceeds the average rate limit. In this case, with a burst limit of 4, the system can process up to 4 requests immediately without enforcing the rate limit.
*/

func init() {
	go cleanUpClients()
}

// RateLimiter middleware for rate limiting
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Rate limiter is enabled
		if os.Getenv("ENABLE_RATE_LIMITER") == "Y" {
			// Get the client IP address
			ip, err := getNetworkIP()
			if err != nil {
				logrus.Error("@RateLimiter Failed to find IP address", err)
				util.UnauthorizedAbortWithJSON(c, "Failed to find IP address")
				return
			}

			// Get endpoint rate limiter
			endpoint := c.FullPath()
			limiter, ok := endpointRates[endpoint]
			if !ok {
				logrus.Error("@RateLimiter Rate limiter not configured for this endpoint ", endpoint)

				//set default rate limiter for all other routes
				limiter = rate.NewLimiter(10, 20)
			}

			mu.Lock()
			// Make map key with combination of IP and API endpoint
			clientKey := ip + ":" + endpoint
			if _, found := clients[clientKey]; !found {
				// Add Key in clients map to track the rate limit
				clients[clientKey] = &RateLimiterClient{limiter: limiter}
			}
			// Update the last seen time of the API which is called by the end user
			clients[clientKey].lastSeen = time.Now()
			mu.Unlock()

			// Check whether rate limit exceeded or not
			if !clients[clientKey].limiter.Allow() {
				util.TooManyRequestsAbortWithJSON(c, "Rate limit exceeded")
				return
			}
		}
		c.Next()
	}
}

// cleanUpClients periodically cleans up old client records
func cleanUpClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, c := range clients {
			t, _ := strconv.Atoi(os.Getenv("BURST_LAST_SEEN"))
			if time.Since(c.lastSeen) > time.Duration(t)*time.Second {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// getNetworkIP retrieves the network IP address
func getNetworkIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() && v.IP.To4() != nil {
					return v.IP.String(), nil
				}
			}
		}
	}
	return "", errors.New("no network ip found")
}
