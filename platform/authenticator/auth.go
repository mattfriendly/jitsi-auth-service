// platform/authenticator/auth.go

package authenticator

import (
    "context"
    "os"

    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)

// Authenticator is used to authenticate our users.
type Authenticator struct {
    Provider *oidc.Provider
    Config   oauth2.Config
    Verifier *oidc.IDTokenVerifier
}

// New instantiates the *Authenticator.
func New() (*Authenticator, error) {
    provider, err := oidc.NewProvider(context.Background(), "https://"+os.Getenv("AUTH0_DOMAIN")+"/")
    if err != nil {
        return nil, err
    }

    config := oauth2.Config{
        ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
        ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
        RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
        Endpoint:     provider.Endpoint(),
        Scopes:       []string{oidc.ScopeOpenID, "profile"},
    }

    verifier := provider.Verifier(&oidc.Config{ClientID: os.Getenv("AUTH0_CLIENT_ID")})

    return &Authenticator{
        Provider: provider,
        Config:   config,
        Verifier: verifier,
    }, nil
}


// Exchange wraps oauth2.Config's Exchange function
func (a *Authenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
    return a.Config.Exchange(ctx, code)
}

// AuthCodeURL wraps oauth2.Config's AuthCodeURL function
func (a *Authenticator) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
    return a.Config.AuthCodeURL(state, opts...)
}

// VerifyIDToken verifies that an ID Token is valid and returns the corresponding IDToken object.
func (a *Authenticator) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
    // Use the verifier to verify the ID token passed in the HTTP request.
    return a.Verifier.Verify(ctx, rawIDToken)
}
