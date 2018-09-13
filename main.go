package main

import (
	"fmt"
	"net/http"
	"log"
	"os"
	"encoding/json"
	"strconv"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/static"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

type Response struct {
	Message	string	`json:"message"`
}

type Joke struct {
	ID		int	`json:"id" binding:"required"`
	Likes	int	`json:"likes"`
	Joke	string	`json:"joke" binding:"required"`
}

type JSONWebKeys struct {
	Kty		string		`json:"kty"`
	Kid		string		`json:"kid"`
	Use		string		`json:"use"`
	N		string		`json:"n"`
	E		string		`json:"e"`
	X5c		[]string	`json:"x5c"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

var jokes = []Joke{
	{1, 0, "Did you hear about the restaurant on the moon? Great food, no atmosphere."},
	{2, 0, "What do you call a fake noodle? An Impasta."},
	{3, 0, "How many apples grow on a tree? All of them."},
	{4, 0, "Want to hear a joke about paper? Nevermind it's tearable."},
	{5, 0, "I just watched a program about beavers. It was the best dam program I've ever seen."},
	{6, 0, "Why did the coffee file a police report? It got mugged."},
	{7, 0, "How does a penguin build it's house? Igloos it together."},
}

var jwtMiddleWare *jwtmiddleware.JWTMiddleware

func main() {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: ValidationKeyGetter,
		SigningMethod: jwt.SigningMethodRS256,
	})

	jwtMiddleWare = jwtMiddleware

	// set the router as the default one shipped with gin
	router := gin.Default()

	// serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("./views", true)))

	api := router.Group("/api")

	{
		api.GET("/", func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	// Extra routes
	api.GET("/jokes", authMiddleware(), JokeHandler)
	api.POST("/jokes/like/:jokeID", authMiddleware(), LikeJoke)

	router.Run(":3000")
}


func JokeHandler(context *gin.Context) {
	context.Header("Content-Type", "application/json")
	context.JSON(http.StatusOK, jokes)
}

func LikeJoke(context *gin.Context) {
	if jokeid, err := strconv.Atoi(context.Param("jokeID")); err == nil {
		for i := 0; i < len(jokes); i++ {
			if jokes[i].ID == jokeid {
				jokes[i].Likes += 1
			}
		}

		context.JSON(http.StatusOK, &jokes)
	} else {
		// not a valid joke id
		context.AbortWithStatus(http.StatusNotFound)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		// get the client  secret key
		err := jwtMiddleWare.CheckJWT(context.Writer, context.Request)

		if err != nil {
			// token not found
			fmt.Println(err)
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized"))
			return
		}
	}
}

func ValidationKeyGetter(token *jwt.Token) (interface{}, error) {
	aud := os.Getenv("AUTH0_API_AUDIENCE")
	checkAudience := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
	if !checkAudience {
		return token, errors.New("Invalid audience.")
	}

	// verify iss claim
	iss := os.Getenv("AUTH0_DOMAIN")
	checkIss := token.Claims.(jwt.MapClaims).VerifyAudience(iss, false)
	if !checkIss {
		return token, errors.New("Invalid issuer.")
	}

	cert, err := getPemCert(token)
	if err != nil {
		log.Fatalf("could not get cert: %+v", err)
	}

	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	return result, nil
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(os.Getenv("AUTH0_DOMAIN") + ".well-known/jwks.json")
	fmt.Println(resp, err)
	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	x5c := jwks.Keys[0].X5c
	for k, v := range x5c {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + v + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return cert, errors.New("unable to find appropriate key.")
	}

	return cert, nil
}
