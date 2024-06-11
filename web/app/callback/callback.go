// web/app/callback/callback.go

package callback

import (
	"net/http"
        "log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jitsi-auth0-service/platform/authenticator"
)

// Handler for our callback.
func Handler(auth *authenticator.Authenticator) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        session := sessions.Default(ctx)
        if ctx.Query("state") != session.Get("state") {
            ctx.String(http.StatusBadRequest, "Invalid state parameter.")
            return
        }

        // Exchange an authorization code for a token.
        token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
        if err != nil {
            ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
            return
        }

        // Log the Access Token
        log.Printf("Access Token: %s", token.AccessToken)

        // Extract the ID Token from the OAuth2 token.
        rawIDToken, ok := token.Extra("id_token").(string)
        if !ok {
            ctx.String(http.StatusInternalServerError, "ID Token not found in the token response.")
            return
        }

        // Verify the ID Token.
        idToken, err := auth.VerifyIDToken(ctx.Request.Context(), rawIDToken)
        if err != nil {
            ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
            return
        }

        // Log the ID Token
        log.Printf("ID Token: %s", rawIDToken)

        // Extract user profile from the ID Token.
        var profile map[string]interface{}
        if err := idToken.Claims(&profile); err != nil {
            ctx.String(http.StatusInternalServerError, err.Error())
            return
        }

        // Store the access token and profile in the session.
        session.Set("access_token", token.AccessToken)
        session.Set("profile", profile)
        if err := session.Save(); err != nil {
            ctx.String(http.StatusInternalServerError, err.Error())
            return
        }

        // Redirect to the logged-in user page.
        ctx.Redirect(http.StatusTemporaryRedirect, "/user")
    }
}
