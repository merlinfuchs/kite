package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type OAuthClient struct {
	// The URL of the client's metadata document.
	ID   string `json:"id"`
	Name string `json:"name"`
	// The host of the ID, which says who the client is. Its name can be
	// anything.
	Domain string `json:"domain"`
}

type OAuthAuthorizeGetResponse struct {
	Client OAuthClient `json:"client"`
	// The scopes the client asks for, the user can grant fewer.
	Scopes []string `json:"scopes"`
}

// OAuthAuthorizeRequest is sent with the client's authorization request in
// the query string, like for OAuthAuthorizeGetResponse.
type OAuthAuthorizeRequest struct {
	// False if the user denied the client.
	Approved bool     `json:"approved"`
	Scopes   []string `json:"scopes"`
	// The apps the client can use, null for all apps of the user.
	AppIDs []string `json:"app_ids" tstype:"string[] | null"`
}

func (req OAuthAuthorizeRequest) Validate() error {
	if !req.Approved {
		return nil
	}

	return validation.ValidateStruct(&req,
		validation.Field(&req.Scopes, apiTokenScopesRules(req.Scopes)...),
		validation.Field(&req.AppIDs, apiTokenAppIDsRules(req.AppIDs)...),
	)
}

type OAuthAuthorizeResponse struct {
	// Where to send the user, back to the client.
	RedirectURL string `json:"redirect_url"`
}
