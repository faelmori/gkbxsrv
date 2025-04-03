package sync

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

func createRequestWithAuth(method, url string, authOptions *AuthOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	var version string
	var versionErr error
	version, versionErr = getGoSpiderVersion()
	if versionErr != nil {
		return nil, versionErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GoSpider/"+version)

	switch authOptions.Type {
	case "basic":
		auth := authOptions.BasicAuth.Username + ":" + authOptions.BasicAuth.Password
		basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+basicAuth)
	case "apiKey":
		req.Header.Set("X-API-Key", authOptions.APIKey) // Ou outra forma que a API exigir
	case "oauth2":
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: authOptions.OAuth2.AccessToken},
		)
		client := oauth2.NewClient(context.TODO(), ts)
		req.Header.Set("Authorization", "Bearer "+authOptions.OAuth2.AccessToken)

		res, resErr := client.Do(req)
		if resErr != nil {
			return nil, resErr
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to authenticate: %s", res.Status)
		}
		authOptions.OAuth2.AccessToken = res.Header.Get("Authorization")

	case "jwt":
		// Implementar lógica de autenticação JWT
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authOptions.Type)
	}

	return req, nil
}

func getGoSpiderVersion() (string, error) {
	// TODO: Implementar lógica para obter a versão da aplicação
	return "1.0.0", nil
}
