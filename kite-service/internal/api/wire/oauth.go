package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// OAuthAuthorizeParams are the parameters of an OAuth client's authorization
// request, which the consent page passes on from its URL.
type OAuthAuthorizeParams struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	State               string `json:"state"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Resource            string `json:"resource"`
}

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

type OAuthAuthorizeRequest struct {
	Params OAuthAuthorizeParams `json:"params"`
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
