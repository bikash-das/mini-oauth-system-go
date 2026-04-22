package main

import (
	"fmt"
	"strings"

	"github.com/bikash-das/mini-oauth-system-go/pkg/config"
	"github.com/bikash-das/mini-oauth-system-go/pkg/logger"
	"github.com/bikash-das/mini-oauth-system-go/pkg/store"
	"github.com/bikash-das/mini-oauth-system-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Logger *logrus.Logger
	Config *config.Config
	Router *gin.Engine
	Store  *store.Store
}

func (s *Server) Authorize(c *gin.Context) {
	redirectUrl := c.Query("redirect_uri")
	state := c.Query("state")
	clientId := c.Query("client_id")
	scope := c.Query("scope")
	// todo: Show login page here.

	code, _ := utils.GenerateRandomCode(6)
	params := map[string]string{
		"code":  code,
		"state": state,
	}
	target, _ := utils.BuildURL(redirectUrl, params)
	// store in db
	s.Store.Set(clientId, map[string]any{
		"state": state, "scope": scope, "code": code, "redirectUrl": redirectUrl})

	s.Logger.Info("Redirecting to ", target)
	// Proof that user is authenticated, but no access token is given.
	// So it redirects the users to the /callback uri with the code,
	// so that the client can request for token using the code and
	// also verify the state value
	// Server to Server communication.

	c.Redirect(302, target)
}

func (s *Server) Token(c *gin.Context) {
	grantType := c.PostForm("grant_type")

	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	logrus.WithFields(logrus.Fields{
		"grant_type":    grantType,
		"code":          code,
		"redirect_uri":  redirectURI,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}).Info("Received token request")

	// grant type check: which oauth flow in use
	if grantType != "authorization_code" && grantType != "refresh_token" {
		c.String(401, "Invalid grant type")
		return
	}

	// if grantType is refresh token then
	// a. invalidate the previous access token - like box.
	// b. issue a new refresh token and access token.
	if grantType == "refresh_token" {
		// todo tusing the refresh token , generate a new access token and refresh token
		// todo Authorization: 'Basic base64(client.clientId + clientSecret)
		// todo cross validate the id and secret which we get in base64 format.
		c.JSON(200, gin.H{
			"access_token":  "very-unsafe-access-token-2",
			"expires_in":    3600,
			"refresh_token": "refresh_12345234234234234234",
			"token_type":    "Bearer",
		})
	}

	data, contains := s.Store.Get(clientID)
	if !contains {
		c.String(401, "Please authorize yourself first.")
		return
	}

	// code check: proof that user authenticated earlier.
	actualCode, ok := data["code"].(string) // assert it's a string
	if !ok {
		c.String(500, "invalid code format")
		return
	}
	if actualCode != code {
		c.String(401, fmt.Sprintf("code didn't match. Provided: %s and Actual: %s", code, actualCode))
		return
	}
	// redirect_uri check: same app that requested code is redeeming it.

	// client secret check: the client itself and not the user.

	// Verify the client secret
	if clientSecret != "B1K2SH12D" {
		s.Logger.Error("Invalid client secret")
		c.String(401, "Invalid client")
		return
	}
	// success
	c.JSON(200, gin.H{
		"access_token":  "very-unsafe-access-token",
		"expires_in":    3600,
		"refresh_token": "refresh_12345",
		"token_type":    "Bearer",
	})

}

func (s *Server) Resource(c *gin.Context) {
	// Validate
	header := c.GetHeader("Authorization")
	fmt.Println(header)
	logrus.Info("header: ", header)
	if header == "" {
		c.IndentedJSON(401, gin.H{"message": "Access denied"})
		return
	}
	accessToken := strings.Split(header, "Bearer ")[1]
	fmt.Println(accessToken)
	fmt.Println("-------------")
	if accessToken != "very-unsafe-access-token" {
		c.IndentedJSON(401, gin.H{"message": "Invalid access token"})
		return
	}
	// todo: Access token valid: We also need to check the scope before accessing the
	// todo resource
	// todo check if token is expired.
	c.IndentedJSON(200, map[string]any{
		"resource_type": "Protected",
	})
}
func (s *Server) NoRoute(c *gin.Context) {
	// collect routes
	var available []map[string]string
	for _, route := range s.Router.Routes() {
		available = append(available, map[string]string{
			"method": route.Method,
			"path":   route.Path,
		})
	}
	s.Logger.WithFields(map[string]any{
		"req":       c.Request.Method + " " + c.Request.URL.Path,
		"available": available,
	}).Warn("Route not found")

	c.JSON(404, gin.H{
		"error":            "not_found",
		"message":          "The requested endpoint does not exist",
		"available_routes": available,
	})
}

func main() {
	fmt.Println("Server starting...")

	r := gin.New()
	r.Use(gin.Recovery())

	cfg := config.Load()
	log := logger.New()
	store := store.NewStore()

	server := &Server{
		Config: cfg,
		Logger: log,
		Router: r,
		Store:  store,
	}

	r.GET("/authorize", server.Authorize)
	r.POST("/token", server.Token)
	r.GET("/resource", server.Resource)
	r.NoRoute(server.NoRoute)

	server.Logger.Info("Server started on :8080")

	r.Run(":8080")
}
