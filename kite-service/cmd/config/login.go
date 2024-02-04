package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
	"github.com/urfave/cli/v2"
)

func loginCMD() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate with a Kite server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server",
				Usage: "The Kite server to authenticate with",
				Value: "https://api.kite.onl",
			},
		},
		Action: func(c *cli.Context) error {
			server := c.String("server")

			cfg, err := config.LoadGlobalConfig()
			if err != nil {
				return err
			}

			serverURL, err := url.Parse(server)
			if err != nil {
				return fmt.Errorf("Failed to parse server URL: %v", err)
			}

			return runLogin(serverURL, cfg)
		},
	}
}

func runLogin(serverURL *url.URL, cfg *config.GlobalConfig) error {
	authStartURL := *serverURL
	authStartURL.Path = path.Join(authStartURL.Path, "/v1/auth/cli/start")
	resp, err := http.Post(authStartURL.String(), "application/json", nil)
	if err != nil {
		return fmt.Errorf("Failed to start auth: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read auth start response: %v", err)
	}

	loginResp := wire.AuthCLIStartResponse{}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("Failed to decode auth start response: %v", err)
	}

	if !loginResp.Success {
		return fmt.Errorf("Failed to start auth: %s (%s)", loginResp.Error.Code, loginResp.Error.Message)
	}

	loginCode := loginResp.Data.Code

	authRedirectURL := *serverURL
	authRedirectURL.Path = path.Join(authRedirectURL.Path, "/v1/auth/cli/redirect")
	authRedirectURL.RawQuery = url.Values{
		"code": {loginCode},
	}.Encode()

	fmt.Printf("Please open the following URL in your browser and login with your Discord account:\n\n%s", authRedirectURL.String())
	fmt.Println("\n\nWaiting for you to complete the login ...")

	authCheckURL := *serverURL
	authCheckURL.Path = path.Join(authCheckURL.Path, "/v1/auth/cli/check")
	authCheckURL.RawQuery = url.Values{
		"code": {loginCode},
	}.Encode()

	for {
		resp, err := http.Get(authCheckURL.String())
		if err != nil {
			return fmt.Errorf("Failed to check auth: %v", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Failed to read auth check response: %v", err)
		}

		checkResp := wire.AuthCLICheckResponse{}
		if err := json.Unmarshal(body, &checkResp); err != nil {
			return fmt.Errorf("Failed to decode auth check response: %v", err)
		}

		if !checkResp.Success {
			return fmt.Errorf("Failed to check auth: %s (%s)", checkResp.Error.Code, checkResp.Error.Message)
		}

		if checkResp.Data.Pending {
			time.Sleep(3 * time.Second)
			continue
		}

		token := checkResp.Data.Token

		session := cfg.GetSessionForServer(serverURL.String())
		if session != nil {
			session.Token = token
		} else {
			cfg.Sessions = append(cfg.Sessions, &config.GlobalSessionConfig{
				Server: serverURL.String(),
				Token:  token,
			})
		}

		err = config.WriteGlobalConfig(cfg)
		if err != nil {
			return fmt.Errorf("Failed to write config: %v", err)
		}

		fmt.Println("Successfully logged in and saved session!")
		break
	}

	return nil
}
