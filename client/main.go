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
	logrus.WithFields(logrus.Fields{
		"state": state,
		"code":  code,
	}).Info("Code and state received!")
	if code == "" {
		c.String(http.StatusBadRequest, "Error: No code returned")
		return
	}
	// Back channel token exchange
	tokenServerURL := "http://localhost:8080/token"

	// prepare the form data to send to the Auth Server
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", s.Config.RedirectURI)
	formData.Set("client_id", s.Config.ClientId)
	formData.Set("client_secret", s.Config.ClientSecret)
	logrus.WithFields(logrus.Fields{
		"event": "token_exchange_start",
		"url":   tokenServerURL,
	}).Info("Exchanging code for access token via back-channel")

	// perform the hidden server to server post request
	resp, err := http.PostForm(tokenServerURL, formData)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to the token endpoint")
	}
	defer resp.Body.Close()

	logrus.WithField("http_statuscode", resp.StatusCode).Info("Received token")

	// 1. Read the body
	bodyBytes, _ := io.ReadAll(resp.Body)

	// 2. Parse JSON into a map
	var body map[string]any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		logrus.WithError(err).Error("Body was not valid JSON")
	}

	logrus.WithFields(body).Info("Token received")

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
	s.Logger.Info("Redirecting user to authorization server")
	c.Redirect(http.StatusFound, authURL)
}

func (s *Server) FetchProtectedResource(c *gin.Context) {
	// if client.AccessToken == "" {
	// 	c.String(401, "No token available. Please /authorize first.")
	// 	return
	// }
	// // Access token present.
	// // create new http request
	// req, _ := http.NewRequest("GET", "http://localhost:8080/resource", nil)
	// req.Header.Set("Authorization", "Bearer "+.AccessToken)
	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	logrus.WithError(err).Error("Error fetching resource with new access token")
	// 	c.String(http.StatusInternalServerError, "Error reaching resource server")
	// 	return
	// }
	// defer resp.Body.Close()
	// logrus.Info("Request for resource completed with status", resp.StatusCode)

	// data, _ := io.ReadAll(resp.Body)
	// var bodyMap map[string]any
	// if err := json.Unmarshal(data, &bodyMap); err != nil {
	// 	c.String(500, "Invliad response while accessing protected resource")
	// 	return
	// }
	// logrus.WithFields(bodyMap)
	//c.IndentedJSON(http.StatusOK, bodyMap)
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

	server.Logger.Info("Server started on :9000")

	router.Run(":9000")

}
