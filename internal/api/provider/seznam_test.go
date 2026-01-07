package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supabase/auth/internal/conf"
	"golang.org/x/oauth2"
)

func TestSeznam(t *testing.T) {
	t.Run("GetUserData", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"oauth_user_id":"test_id","email":"test@example.com","firstname":"Test","lastname":"User","avatar_url":"http://example.com/avatar"}`))
		}))
		defer server.Close()

		p := seznamProvider{
			Config: &oauth2.Config{},
			APIURL: server.URL,
		}

		data, err := p.GetUserData(context.Background(), &oauth2.Token{
			AccessToken: "token",
		})
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", data.Emails[0].Email)
		assert.Equal(t, "test_id", data.Metadata.Subject)
		assert.Equal(t, "Test User", data.Metadata.Name)
		assert.Equal(t, "http://example.com/avatar", data.Metadata.Picture)
	})

	t.Run("GetUserData Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		p := seznamProvider{
			Config: &oauth2.Config{},
			APIURL: server.URL,
		}

		_, err := p.GetUserData(context.Background(), &oauth2.Token{
			AccessToken: "token",
		})
		assert.Error(t, err)
	})

	t.Run("GetUserData Missing ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"email":"test@example.com"}`))
		}))
		defer server.Close()

		p := seznamProvider{
			Config: &oauth2.Config{},
			APIURL: server.URL,
		}

		_, err := p.GetUserData(context.Background(), &oauth2.Token{
			AccessToken: "token",
		})
		assert.Error(t, err)
		assert.Equal(t, errors.New("user id was not returned from Seznam API"), err)
	})

	t.Run("NewSeznamProvider", func(t *testing.T) {
		p, err := NewSeznamProvider(conf.OAuthProviderConfiguration{
			ClientID:    []string{"client_id"},
			Secret:      "secret",
			RedirectURI: "http://localhost/callback",
			URL:         "http://example.com",
		}, "")
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})
}
