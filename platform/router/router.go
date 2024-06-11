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

	"jitsi-auth0-service/platform/authenticator"
	"jitsi-auth0-service/platform/middleware"
	"jitsi-auth0-service/web/app/callback"
	"jitsi-auth0-service/web/app/login"
	"jitsi-auth0-service/web/app/logout"
	"jitsi-auth0-service/web/app/user"
)

// Somewhere in your router setup file
/*func authCheckHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract the token from the Authorization header or cookie
        token, err := c.Cookie("auth_token") // or c.GetHeader("Authorization")
        if err != nil {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // Verify the ID Token
        _, err = auth.VerifyIDToken(c.Request.Context(), token)
        if err != nil {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // If token is valid, proceed
        c.Status(http.StatusOK)
    }
}
*/

func authCheckHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract the token from the Authorization header
        authHeader := c.GetHeader("Authorization")
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

        // Token is valid, and userinfo is obtained successfully
        c.JSON(http.StatusOK, gin.H{"success": true, "data": string(body)})
    }
}

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
	router.LoadHTMLGlob("web/template/user.html")
	router.LoadHTMLGlob("web/template/home.html")

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "home.html", nil)
	})
	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth))
	router.GET("/user", middleware.IsAuthenticated, user.Handler)
	router.GET("/logout", logout.Handler)
        router.GET("/auth_check", authCheckHandler(auth))

	return router
}
