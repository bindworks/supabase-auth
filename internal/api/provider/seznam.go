package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/supabase/auth/internal/conf"
	"golang.org/x/oauth2"
)

// Seznam provider constants
const (
	defaultSeznamAuthURL  = "https://login.szn.cz/api/v1/oauth/auth"
	defaultSeznamTokenURL = "https://login.szn.cz/api/v1/oauth/token"
	defaultSeznamUserURL  = "login.szn.cz/api/v1/user"
)

type seznamProvider struct {
	*oauth2.Config
	APIURL string
}

type seznamUser struct {
	ID            string `json:"oauth_user_id"`
	Email         string `json:"email"`
	Name          string `json:"firstname"`
	FamilyName    string `json:"lastname"`
	AvatarURL     string `json:"avatar_url"`
	EmailVerified bool   `json:"email_verified"` // This might need to be checked against real response, docs claim 'identity' scope provides email
}

// NewSeznamProvider creates a Seznam OAuth2 identity provider.
func NewSeznamProvider(ext conf.OAuthProviderConfiguration, scopes string) (OAuthProvider, error) {
	if err := ext.ValidateOAuth(); err != nil {
		return nil, err
	}

	oauthScopes := []string{
		"identity",
	}

	if scopes != "" {
		oauthScopes = append(oauthScopes, strings.Split(scopes, ",")...)
	}

	apiPath := chooseHost(ext.URL, defaultSeznamUserURL)

	return &seznamProvider{
		Config: &oauth2.Config{
			ClientID:     ext.ClientID[0],
			ClientSecret: ext.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  defaultSeznamAuthURL,
				TokenURL: defaultSeznamTokenURL,
			},
			RedirectURL: ext.RedirectURI,
			Scopes:      oauthScopes,
		},
		APIURL: apiPath,
	}, nil
}

func (g seznamProvider) GetOAuthToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return g.Exchange(ctx, code, opts...)
}

func (g seznamProvider) RequiresPKCE() bool {
	return true
}

func (g seznamProvider) GetUserData(ctx context.Context, tok *oauth2.Token) (*UserProvidedData, error) {
	var u seznamUser
	if err := makeRequest(ctx, tok, g.Config, g.APIURL, &u); err != nil {
		return nil, err
	}

	if u.ID == "" {
		return nil, errors.New("user id was not returned from Seznam API")
	}

	var data UserProvidedData

	if u.Email != "" {
		data.Emails = append(data.Emails, Email{
			Email:    u.Email,
			Verified: true, // Seznam verifies emails
			Primary:  true,
		})
	}

	data.Metadata = &Claims{
		Issuer:        g.APIURL,
		Subject:       u.ID,
		Name:          strings.TrimSpace(u.Name + " " + u.FamilyName),
		GivenName:     u.Name,
		FamilyName:    u.FamilyName,
		Picture:       u.AvatarURL,
		Email:         u.Email,
		EmailVerified: true, // Seznam verifies emails

		// To be deprecated
		AvatarURL:  u.AvatarURL,
		FullName:   strings.TrimSpace(u.Name + " " + u.FamilyName),
		ProviderId: u.ID,
	}

	return &data, nil
}
