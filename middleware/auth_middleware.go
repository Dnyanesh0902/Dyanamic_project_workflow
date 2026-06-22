package middleware

import (
	"fmt"
	"net/http"
	"os"
	"project-workflow-backend/service"
	"project-workflow-backend/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// func TokenAuthentication() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		requestURL := c.Request.RequestURI
// 		fmt.Println("requestURL:", requestURL)
// 		TokenAuthenticationHandler(c)
// 	}
// }

func TokenAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			util.RespondWithError(c, http.StatusUnauthorized, "Missing token.")
			c.Abort()
			return
		}

		token, err := service.ValidateToken(tokenString)
		if err != nil {
			util.RespondWithError(c, http.StatusUnauthorized, "Unauthorized Access.")
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		c.Request.Header.Set("user_id", fmt.Sprintf("%v", claims["uuid"]))
		c.Request.Header.Set("auth_user_id", fmt.Sprintf("%v", claims["id"]))
		c.Request.Header.Set("user_type", fmt.Sprintf("%v", claims["user_type"]))
		c.Request.Header.Set("pincode", fmt.Sprintf("%v", claims["pincode"]))
		c.Next()
	}
}

func TokenAuthenticationHandler(c *gin.Context) {
	fmt.Println("app_env : ", os.Getenv("APP_ENV"))
	authorization := c.Request.Header.Get("Authorization")
	if authorization != "" {
		tokenValidClaims, tokenErr := TokenValid(authorization)
		fmt.Println(tokenValidClaims)
		logrus.Info("Tokendata: ", tokenValidClaims)
		if tokenErr != nil {
			util.UnauthorizedAbortWithJSON(c, "Invalid Token.")
		}
		logrus.Info("tokenValidClaims", tokenValidClaims)
		ssoId := tokenValidClaims["sub"].(string)
		c.Request.Header.Set("sub_id", ssoId)
		c.Next()
	} else {
		util.UnauthorizedAbortWithJSON(c, "Unauthorized Access.")
		return
	}
}
func VerifyToken(tokenString string) (*jwt.Token, error) {

	pubKeyFileExt := "pem"
	if os.Getenv("APP_ENV") == "production" {
		pubKeyFileExt = "pub"
	}
	keyData, err := os.ReadFile("storage/ssl/" + os.Getenv("APP_ENV") + "/jwtRSA256-public-" + os.Getenv("APP_ENV") + "." + pubKeyFileExt)
	if err != nil {
		logrus.Info("testkeyData err", err)
	}
	logrus.Info("VerifyToken@Path : ", "storage/ssl/"+os.Getenv("APP_ENV")+"/jwtRSA256-public-"+os.Getenv("APP_ENV")+"."+pubKeyFileExt)

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		logrus.Info("key err", err)
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}
func TokenValid(tokenString string) (jwt.MapClaims, error) {
	token, err := VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
