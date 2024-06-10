// platform/router/router.go

package router

import (
	"encoding/gob"
	"net/http"

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
func authCheckHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
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
