// web/app/login/login.go

package login

import (
    "crypto/rand"
    "encoding/base64"
    "net/http"

    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"

    "jitsi-auth0-service/platform/authenticator"
)

// Handler for our login.
func Handler(auth *authenticator.Authenticator) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // Generate a secure, random state for OAuth flow
        state, err := generateRandomState()
        if err != nil {
            ctx.String(http.StatusInternalServerError, "Failed to generate state: %s", err)
            return
        }

        // Save the state inside the session to validate response from Auth0
        session := sessions.Default(ctx)
        session.Set("state", state)
        if err := session.Save(); err != nil {
            ctx.String(http.StatusInternalServerError, "Failed to save session: %s", err)
            return
        }

        // Redirect to Auth0 login URL with the state
        redirectURL := auth.AuthCodeURL(state)
        ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
    }
}

// generateRandomState generates a secure random state for OAuth flows
func generateRandomState() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(b), nil
}
