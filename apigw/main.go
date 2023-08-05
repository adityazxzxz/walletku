package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var app *firebase.App
var authClient *auth.Client

func initializeFirebase() error {
	opt := option.WithCredentialsFile("credential/service.json") // Replace with the path to your Firebase Admin SDK configuration file
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return err
	}

	authClient, err = app.Auth(ctx)
	if err != nil {
		return err
	}

	return nil
}

func checkToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenArr := strings.Split(tokenString, " ")
		if len(tokenArr) != 2 || tokenArr[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		tokenString = tokenArr[1]

		// Verifikasi token JWT dengan Firebase Auth
		token, err := authClient.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Set informasi pengguna dari token JWT ke konteks untuk digunakan di handler lain
		c.Set("userID", token.UID)

		c.Next()
	}
}

func reverseProxy(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqURL := targetURL + c.Request.URL.String()

		req, err := http.NewRequest(c.Request.Method, reqURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			c.Abort()
			return
		}

		for k, v := range c.Request.Header {
			req.Header.Set(k, v[0])
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
			c.Abort()
			return
		}

		defer resp.Body.Close()

		for k, v := range resp.Header {
			c.Writer.Header().Set(k, v[0])
		}

		c.Writer.WriteHeader(resp.StatusCode)
		c.Writer.Flush()

		if _, err := c.Writer.Write([]byte{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write response"})
			c.Abort()
			return
		}
	}
}

func main() {
	if err := initializeFirebase(); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v\n", err)
		return
	}

	r := gin.Default()

	// Rute tanpa grup, terapkan middleware checkToken pada rute /service1/*any
	r.GET("/service1/*any", checkToken(), reverseProxy("http://localhost:8081"))

	// Rute tanpa grup, tanpa middleware checkToken pada rute /service2/*any
	r.GET("/service2/*any", reverseProxy("http://localhost:8082"))

	r.Run(":8080")
}
