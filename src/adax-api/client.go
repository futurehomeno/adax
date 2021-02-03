package adax

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
)

const (
	API_URL         = "https://api-1.adax.no/client-api"
	adaxPartnerCode = "adax"
)

type (
	Config struct {
		ErrorCode  int    `json:"errorCode"`
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode"`
		Success    bool   `json:"success"`
	}

	Client struct {
		Oauth2Client *edgeapp.FhOAuth2Client
		HttpClient   *http.Client
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         int    `json:"id"`
		Args         struct {
			Code string `"son:"code"`
		} `json:"args"`
		// States        []State
		// HomesAndRooms []HomesAndRooms
	}

	State struct {
		Users []struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
			Homes  []struct {
				ID    int `json:"id"`
				Rooms []struct {
					ID                int  `json:"id"`
					HeatingEnabled    bool `json:"heatingEnabled"`
					TargetTemperature int  `json:"targetTemperature"`
					Temperature       int  `json:"temperature"`
					Devices           []struct {
						ID         int `json:"id"`
						PowerUsage struct {
							TimeFrom int64 `json:"timeFrom"`
							TimeTo   int64 `json:"timeTo"`
							Energy   int   `json:"energy"`
						} `json:"powerUsage"`
						Online bool `json:"online"`
					} `json:"devices"`
				} `json:"rooms"`
			} `json:"homes"`
		} `json:"users"`
	}

	HomesAndRooms struct {
		Users []struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
			Homes  []struct {
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Rooms []struct {
					ID      int    `json:"id"`
					Name    string `json:"name"`
					Devices []struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"devices"`
				} `json:"rooms"`
			} `json:"homes"`
		} `json:"users"`
	}
)

func NewClient(accessToken, refreshToken string) *Client {

	return &Client{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		HttpClient:   &http.Client{Timeout: 30 * time.Second}, // Very important to set timeout
	}
}

func (clt *Client) SetTokens(accessToken, refreshToken string) {
	clt.AccessToken = accessToken
	clt.RefreshToken = refreshToken
}

func (clt *Client) Login(username, password string) error {
	url := fmt.Sprintf("%s%s", API_URL, "/auth/token")
	method := "POST"

	payloadString := fmt.Sprintf("%s%s%s%s", "grant_type=password&username=", username, "&password=", password)
	payload := strings.NewReader(payloadString)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	type AccessResponse struct {
		oauth2Client *edgeapp.FhOAuth2Client
		httpClient   *http.Client
		accessToken  string `json:"access_token"`
		refreshToken string `json:"refresh_token"`
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	ProcessHTTPResponse(res, err, clt)

	log.Debug("<client> New access token: ", clt.AccessToken)
	defer res.Body.Close()

	return err
}

func (clt *Client) GetUsers(accessToken string) (int, error) {
	url := "https://adax-api-test.azurewebsites.net/r-test-api/rest/users"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", clt.AccessToken))

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if ProcessHTTPResponse(res, err, clt) != nil {
		return 0, err
	}
	return clt.User, nil
}

func (clt *Client) GetCode() (string, error) {
	url := "https://adax-api-test.azurewebsites.net/r-test-api/auth/auth?client_id=future-home-api-test&response_type=code&state=your_state&scope=your_scope&redirect_uri=https://postman-echo.com/get?foo1=bar1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if ProcessHTTPResponse(res, err, clt) != nil {
		return "", err
	}
	defer res.Body.Close()

	return clt.Args.Code, nil
}

func (clt *Client) GetTokens(code string) (string, string, error) {
	url := "https://adax-api-test.azurewebsites.net/r-test-api/auth/token"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf("%s%s%s", "grant_type=authorization_code&code=", code, "&redirect_uri=https%3A%2F%2Fpostman-echo.com%2Fget%3Ffoo1%3Dbar1"))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic ZnV0dXJlLWhvbWUtYXBpLXRlc3Q6djRwaFZVVm1Hbkg2M3hRMk93VTJncTU2MjRtS2F1aVVrd1gySXhEbzRWelF0Tm50SnJBdTRNWnFLRkRCVDRWc3JQcWFJZVBWYkszT0FiNnBaaE1wN0xCNXBqaWsySTh1TkVoY0d3UHNEc0h0T3M1Y0Zoelg4U3JvbG9JSGV4WmU=")

	res, err := client.Do(req)
	if ProcessHTTPResponse(res, err, clt) != nil {
		return "", "", err
	}
	defer res.Body.Close()

	return clt.AccessToken, clt.RefreshToken, nil

}

func (clt *Client) RefreshTokens(refreshToken string) error {
	url := fmt.Sprintf("%s%s%s", API_URL, "/auth/token")
	method := "POST"

	if refreshToken == "" {
		refreshToken = clt.RefreshToken
	}
	if refreshToken == "" {
		log.Error("<client> Empty refresh token")
		return fmt.Errorf("empty refresh token")
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if ProcessHTTPResponse(res, err, clt) != nil {
		return err
	}
	log.Debug("<client> New access token: ", clt.AccessToken)
	defer res.Body.Close()

	return err
}

func (clt *Client) UpdateAuthParameters(mqttBrokerUri string) {
	clt.Oauth2Client.SetParameters(mqttBrokerUri, "", "", 0, 0, 0, 0)
}

func (clt *Client) RefreshAccessToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		refreshToken = clt.RefreshToken
	}
	if refreshToken == "" {
		log.Error("<client> Empty refresh token")
		return "", fmt.Errorf("empty refresh token")
	}
	resp, err := clt.Oauth2Client.ExchangeRefreshToken(refreshToken)
	if err != nil {
		log.Error("can't fetch new access token", err)
		return "", err
	}
	log.Debug("<client> New access token : ", resp.AccessToken)
	clt.AccessToken = resp.AccessToken
	return resp.AccessToken, nil
}

// do a generic HTTP request
func (clt *Client) doHttpRequest(req *http.Request) ([]byte, error) {
	var err error
	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = clt.HttpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			break
		} else if resp.StatusCode == 401 {
			log.Info("Invalid token . Retrying")
			_, err = clt.RefreshAccessToken(clt.RefreshToken)
			if err != nil {
				time.Sleep(time.Second * 5)
			} else {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clt.AccessToken))
			}
		} else if resp.StatusCode != 200 {
			log.Error("Bad HTTP return code ", resp.StatusCode)
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return nil, fmt.Errorf("bad HTTP return code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func ProcessHTTPResponse(resp *http.Response, err error, holder interface{}) error {
	if err != nil {
		log.Error(fmt.Errorf("API does not respond"))
		return err
	}
	defer resp.Body.Close()
	// check http return code
	if resp.StatusCode != 200 {
		//bytes, _ := ioutil.ReadAll(resp.Body)

		log.Error("Bad HTTP return code ", resp.StatusCode)
		return fmt.Errorf("Bad HTTP return code %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(holder)
}