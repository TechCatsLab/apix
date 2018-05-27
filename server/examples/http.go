/*
 * Revision History:
 *     Initial: 2018/5/27        ShiChao
 */

package main

import (
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"

	"github.com/fengyfei/gu/libs/constants"
	"github.com/fengyfei/gu/libs/http/server"
	"github.com/fengyfei/gu/libs/http/server/middleware"
	"github.com/fengyfei/gu/libs/logger"
)

const (
	tokenExpireInHour = 48

	claimsKey    = "user"
	claimUID     = "uid"
	claimExpire  = "exp"
	respTokenKey = "token"

	invalidUID int32 = -1
)

var (
	tokenHMACKey        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	defaultUserID int32 = 2
	URLMap        map[string]struct{}

	jwtConfig = middleware.JWTConfig{
		Skipper:    customSkipper,
		SigningKey: []byte(tokenHMACKey),
		// use to extract claims from context
		ContextKey: claimsKey,
	}
)

func init() {
	URLMap = make(map[string]struct{})

	URLMap["/login"] = struct{}{}
}

type (
	loginReq struct {
		Name *string `json:"name" validate:"required,alphanum,min=2,max=30"`
		Pwd  *string `json:"pwd" validate:"required,printascii,excludesall=@-,min=6,max=30"`
	}

	echo struct {
		Message *string `json:"message" validate:"required,min=6"`
	}
)

func echoHandler(c *server.Context) error {
	var (
		err  error
		req  echo
		user string
	)

	header := c.GetHeader("auth")
	logger.Debug("header:", header)
	c.SetHeader("auth", "vbnmfs")

	user = c.FormValue("user")
	logger.Debug("user:", user)

	c.SetCookie("id", "123456")
	id, err := c.GetCookie("id")
	if err != nil {
		return err
	}
	logger.Debug("id:", id.Value)

	if err = c.JSONBody(&req); err != nil {
		logger.Error("Parses the JSON request body error:", err)
		return err
	}
	err = c.Validate(&req)
	if err != nil {
		logger.Error(err)
		return err
	}

	return c.ServeJSON(&req)
}

func postHandler(c *server.Context) error {
	// w.Write([]byte("Post\n"))
	return nil
}

func panicHandler(c *server.Context) error {
	panic("Panic testing")
	return nil
}

func loginFilter(c *server.Context) bool {
	if c.Request().RequestURI != "/login" {
		uid := getUID(c)

		if uid != defaultUserID {
			// must call ServeJSON
			c.ServeJSON(map[string]interface{}{constants.RespKeyStatus: constants.ErrPermission})
			return false
		}
	}

	return true
}

func login(c *server.Context) error {
	var (
		err error
		req loginReq
	)

	if err = c.JSONBody(&req); err != nil {
		logger.Error(err)
		return c.ServeJSON(map[string]interface{}{constants.RespKeyStatus: constants.ErrInvalidParam})
	}

	_, token, err := NewToken(defaultUserID)
	if err != nil {
		logger.Error(err)
		return c.ServeJSON(map[string]interface{}{constants.RespKeyStatus: constants.ErrPermission})
	}

	logger.Debug("token:", token)
	return c.ServeJSON(map[string]interface{}{
		constants.RespKeyStatus: constants.ErrSucceed,
		constants.RespKeyToken:  token,
	})
}

func verifyJWT(c *server.Context) error {
	uid := getUID(c)

	if uid != defaultUserID {
		logger.Debug("JWT verify failure...")
		return c.ServeJSON(map[string]interface{}{constants.RespKeyStatus: constants.ErrPermission})
	}

	logger.Debug("JWT verify success!")
	return c.ServeJSON(map[string]interface{}{constants.RespKeyStatus: constants.ErrSucceed})
}

// getUID extract uid from context value
func getUID(c *server.Context) int32 {
	rawclaims := c.Request().Context().Value(claimsKey)
	claims := rawclaims.(jwtgo.MapClaims)

	if uid, ok := claims[claimUID].(float64); ok {
		return int32(uid)
	}

	return invalidUID
}

// NewToken generates a JWT token.
func NewToken(uid int32) (string, string, error) {
	token := jwtgo.New(jwtgo.SigningMethodHS256)

	claims := token.Claims.(jwtgo.MapClaims)
	claims[claimUID] = uid
	claims[claimExpire] = time.Now().Add(time.Hour * tokenExpireInHour).Unix()

	t, err := token.SignedString([]byte(tokenHMACKey))
	return respTokenKey, t, err
}

func customSkipper(c *server.Context) bool {
	if _, ok := URLMap[c.Request().RequestURI]; ok {
		return true
	}

	return false
}

func main() {
	configuration := &server.Configuration{
		Address: "0.0.0.0:9573",
	}

	// routers
	router := server.NewRouter()
	router.Post("/", echoHandler, loginFilter)
	router.Post("/post", postHandler, loginFilter)
	router.Get("/panic", panicHandler, loginFilter)
	router.Post("/login", login)
	router.Post("/verify", verifyJWT, loginFilter)

	// add middlewares
	corsMiddleware := middleware.CORSAllowAll()
	jwtMiddleware := middleware.JWTWithConfig(jwtConfig)

	ep := server.NewEntrypoint(configuration, nil)

	ep.AttachMiddleware(middleware.NegroniRecoverHandler())
	ep.AttachMiddleware(middleware.NegroniLoggerHandler())
	ep.AttachMiddleware(corsMiddleware)
	ep.AttachMiddleware(jwtMiddleware)

	if err := ep.Start(router.Handler()); err != nil {
		logger.Error(err)
		return
	}

	ep.Wait()
}
