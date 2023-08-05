package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"path/filepath"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type UserCredentials struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
	Fullname string `form:"fullname" binding:"required"`
}

type UserResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

var app *firebase.App
var authClient *auth.Client

func initializeFirebase() error {
	opt := option.WithCredentialsFile("credential/service.json")
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

func register(c *gin.Context) {
	var credentials UserCredentials
	if err := c.ShouldBind(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"})
		return
	}

	ktp, err := c.FormFile("ktp")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		return
	}

	// save file into path server
	ktpFileName := filepath.Base(ktp.Filename)

	if err := c.SaveUploadedFile(ktp, filepath.Join("upload", ktpFileName)); err != nil {
		fmt.Println("error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save KTP"})
		return
	}

	params := (&auth.UserToCreate{}).
		Email(credentials.Email).
		Password(credentials.Password).
		EmailVerified(false).
		// PhoneNumber("+15555550100").
		DisplayName(credentials.Fullname).
		PhotoURL("http://www.example.com/12345678/photo.png").
		Disabled(false)

	ctx := context.Background()
	userRecord, err := authClient.CreateUser(ctx, params)

	if err != nil {
		fmt.Println("error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	response := UserResponse{
		UserID: userRecord.UID,
		Email:  userRecord.Email,
	}

	c.JSON(http.StatusCreated, response)
}

func main() {

	if err := initializeFirebase(); err != nil {
		log.Fatalf("Failed initialize firebase: %v\n", err)
		return
	}

	r := gin.Default()

	r.POST("/users", register)

	r.Run(":8080")

	// opt := option.WithCredentialsFile("credential/service.json")
	// app, err := firebase.NewApp(context.Background(), nil, opt)
	// if err != nil {
	// 	log.Fatalf("error initializing app: %v\n", err)
	// }

	// client, err := app.Auth(context.Background())
	// if err != nil {
	// 	log.Fatalf("error getting Auth client: %v\n", err)
	// }

	// params := (&auth.UserToCreate{}).
	// 	Email("user@example.com").
	// 	EmailVerified(false).
	// 	PhoneNumber("+15555550100").
	// 	Password("secretPassword").
	// 	DisplayName("John Doe").
	// 	PhotoURL("http://www.example.com/12345678/photo.png").
	// 	Disabled(false)
	// u, err := client.CreateUser(context.Background(), params)
	// if err != nil {
	// 	log.Fatalf("error creating user: %v\n", err)
	// }
	// log.Printf("Successfully created user: %v\n", u)

}
