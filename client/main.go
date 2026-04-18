package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bikash-das/mini-oauth-system-go/pkg/config"
	"github.com/bikash-das/mini-oauth-system-go/pkg/logger"
	"github.com/bikash-das/mini-oauth-system-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Config *config.Config
	Logger *logrus.Logger
}

func (s *Server) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.String(http.StatusBadRequest, "Error: No code or state returned")
		return
	}

	// prepare the form data to send to the Auth Server
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", s.Config.RedirectURI)
	formData.Set("client_id", s.Config.ClientId)
	formData.Set("client_secret", s.Config.ClientSecret)

	// perform the hidden server to server post request
	s.Logger.WithFields(map[string]any{
		"url":    s.Config.AuthServerTokenURL,
		"method": "POST",
		"form":   formData.Encode(),
	}).Info("Calling token endpoint")

	resp, err := http.PostForm(s.Config.AuthServerTokenURL, formData)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to the token endpoint")
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		s.Logger.WithFields(map[string]any{
			"status": resp.StatusCode,
			"url":    s.Config.AuthServerTokenURL,
			"body":   string(bodyBytes),
		}).Error("Token endpoint returned non 200 response")

		c.JSON(resp.StatusCode, gin.H{
			"error":       "token_request_failed",
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		})
		return
	}

	// 2. Parse JSON into a map
	var body map[string]any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		logrus.WithError(err).Error("Body was not valid JSON: ")
		c.JSON(400, gin.H{
			"error":   "invalid_request",
			"message": "Request body is not valid JSON",
			"body":    body,
		})
		return
	}

	// Save the token
	// Store it (for now, in our global variable)
	//client.AccessToken = bodyMap["access_token"].(string)
	//logrus.WithFields(resp).Info("Credentials")
	c.IndentedJSON(http.StatusOK, body)

}

func (s *Server) Authorize(c *gin.Context) {
	params := map[string]string{
		"response_type": "code",
		"client_id":     s.Config.ClientId,
		"redirect_uri":  s.Config.RedirectURI,
		"scope":         s.Config.Scope,
		"state":         "bik123", // prevents CSRF attacks
	}
	authURL, err := utils.BuildURL(s.Config.AuthServerURL, params)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error: Forming AuthURL")
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (s *Server) FetchProtectedResource(c *gin.Context) {
	accessToken := "very-unsafe-access-token" // after /token store this somewhere safe.
	// userA : accesstoken - then access the resource.
	// todo: checkk if access token is expired for userA
	// todo: using refresh token , fetch the access token then make
	// call to protected resource - using basic auth
	// Header: Basic base64(clientId+clientSecret)
	// Body: { grant_type: 'refresh_token', refresh_token: refreshToken}
	// create new http request
	req, _ := http.NewRequest("GET", "http://localhost:8080/resource", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Error fetching resource with new access token")
		c.String(http.StatusInternalServerError, "Error reaching resource server")
		return
	}
	defer resp.Body.Close()
	logrus.Info("Request for resource completed with status", resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var bodyMap map[string]any
	if err := json.Unmarshal(data, &bodyMap); err != nil {
		c.String(500, "Invliad response while accessing protected resource")
		return
	}
	logrus.WithFields(bodyMap)
	c.IndentedJSON(http.StatusOK, bodyMap)
	c.String(200, "")
}

func main() {

	var log *logrus.Logger = logger.New()
	var cfg *config.Config = config.Load()
	fmt.Println(*cfg)

	var server *Server = &Server{
		Config: cfg,
		Logger: log,
	}

	var router *gin.Engine = gin.New()
	router.Use(gin.Recovery())

	router.GET("/callback", server.Callback)
	router.GET("/authorize", server.Authorize)
	router.GET("/fetch-protected-resource", server.FetchProtectedResource)

	server.Logger.Info("--------- Server started on :9000 ---------- ")

	router.Run(":9000")

}
