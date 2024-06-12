// platform/router/router.go

package router

import (
	"encoding/gob"
	"net/http"
        "io/ioutil"
        "log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

        "jitsi-auth0-service/platform"
	"jitsi-auth0-service/platform/authenticator"
	"jitsi-auth0-service/platform/middleware"
	"jitsi-auth0-service/web/app/callback"
	"jitsi-auth0-service/web/app/login"
	"jitsi-auth0-service/web/app/logout"
	"jitsi-auth0-service/web/app/user"
)

func getClientIP(c *gin.Context) string {
    // Retrieve IP from the X-Real-IP header set by Nginx
    ip := c.GetHeader("X-Real-IP")
    if ip == "" {
        ip = c.ClientIP()  // Fallback to default method if header is not set
    }
    return ip
}

func authCheckHandler(auth *authenticator.Authenticator, ipStore *platform.IPStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := getClientIP(c)
        if !ipStore.Exists(ip) {
            log.Printf("Access denied for unrecognized IP: %s", ip)
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        log.Printf("Access granted for recognized IP: %s", ip)

        // Try to extract the token from the Authorization header first
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            // If Authorization header is missing, try to extract the token from cookies
            if cookie, err := c.Cookie("auth_token"); err == nil {
                authHeader = "Bearer " + cookie
            }
        }

        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            return
        }

        // Prepare the request to the Auth0 userinfo endpoint
        client := &http.Client{}
        req, err := http.NewRequest("GET", "https://dev-4acakivw13xuzj6b.us.auth0.com/userinfo", nil)
        if err != nil {
            log.Printf("Request creation failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
            return
        }
        req.Header.Add("Authorization", authHeader)

        // Send the request to Auth0
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Error during Auth0 request: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with Auth0"})
            return
        }
        defer resp.Body.Close()

        // Read the response body
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Error reading response body: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response from Auth0"})
            return
        }

        // Check the response status code from Auth0
        if resp.StatusCode != http.StatusOK {
            log.Printf("Auth0 verification failed: %s", string(body))
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "auth0_response": string(body)})
            return
        }

        log.Printf("Token validated successfully, User info: %s", string(body))
        c.JSON(http.StatusOK, gin.H{"success": true, "data": string(body)})
    }
}

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator, ipStore *platform.IPStore) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
        router.LoadHTMLFiles("web/template/user.html", "web/template/home.html")

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "home.html", nil)
	})
	router.GET("/login", login.Handler(auth, ipStore))
        router.GET("/auth_check", authCheckHandler(auth, ipStore))
	router.GET("/callback", callback.Handler(auth))
	router.GET("/user", middleware.IsAuthenticated, user.Handler)
	router.GET("/logout", logout.Handler)

	return router
}
