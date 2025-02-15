package user

import (
	"github.com/IlyaZayats/auth/internal/dto"
	"github.com/IlyaZayats/auth/internal/entities"
	"github.com/IlyaZayats/auth/internal/handlers/request"
	"github.com/IlyaZayats/auth/internal/middleware"
	"github.com/IlyaZayats/auth/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type UserHandler struct {
	userService *services.UserService
	engine      *gin.Engine
}

func (h *UserHandler) InitRoutes() {

	h.engine.POST("/auth", h.Auth)
	h.engine.POST("/register", h.Register)
	h.engine.GET("/update_access_token",
		middleware.AuthMiddleware(), h.UpdateAccessToken)
	h.engine.GET("/test", h.Test)

	h.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"},
		AllowCredentials: true,
	}))

	//config := cors.DefaultConfig()
	//config.AllowAllOrigins = true
	//config.AllowCredentials = true
	//config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	//config.AllowHeaders = []string{"Access-Control-Allow-Headers", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"}
	//h.engine.Use(cors.New(config))
}

func NewUserHandler(srv *services.UserService, engine *gin.Engine) (*UserHandler, error) {
	h := &UserHandler{
		userService: srv,
		engine:      engine,
	}
	h.InitRoutes()
	return h, nil
}

func (h *UserHandler) Test(c *gin.Context) {
	c.JSON(http.StatusOK, "krp")
}

func (h *UserHandler) Auth(c *gin.Context) {
	req, ok := request.GetRequest[dto.AuthRequest](c)
	logrus.Debug(req)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth request error", "text": ok})
		return
	}
	user, userTokens, err := h.userService.Auth(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth service error", "text": err.Error()})
		return
	}

	resp := dto.AuthResponse{
		Email:                 user.Email,
		Password:              user.Password,
		AccessTokenString:     userTokens[0].TokenString,
		AccessTokenExpiresAt:  userTokens[0].ExpiresAt,
		AccessTokenIssuedAt:   userTokens[0].IssuedAt,
		RefreshTokenString:    userTokens[1].TokenString,
		RefreshTokenExpiresAt: userTokens[1].ExpiresAt,
		RefreshTokenIssuedAt:  userTokens[1].IssuedAt,
	}

	logrus.Debug(resp)

	c.SetCookie("refresh_token", resp.RefreshTokenString, 60*60*60*60*24, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}

func (h *UserHandler) Register(c *gin.Context) {
	req, ok := request.GetRequest[dto.RegisterRequest](c)
	logrus.Debug(req)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "register request error", "text": ok})
		return
	}
	usr, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "register service error", "text": err.Error()})
		return
	}
	logrus.Debug(usr)

	resp := dto.RegisterResponse{
		Id:         usr.Id,
		LastName:   usr.LastName,
		FirstName:  usr.FirstName,
		MiddleName: usr.MiddleName,
		Email:      usr.Email,
		Password:   usr.Password,
		Passport:   usr.Passport,
		Inn:        usr.Inn,
		Snils:      usr.Snils,
		Birthday:   usr.Birthday,
		Role:       usr.Role,
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}

func (h *UserHandler) UpdateAccessToken(c *gin.Context) {
	//req, ok := request.GetRequest[dto.UpdateAccessTokenRequest](c)
	logrus.Info("старт обработки запроса на обновление токена")
	accessToken := c.MustGet("access_token").(string)
	refreshToken, err := c.Cookie("refresh_token")
	//test, err := c.Cookie()
	logrus.Info("access token: ", accessToken)
	logrus.Info("refresh token: ", refreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error refresh token parsing from cookie", "text": err.Error()})
		return
	}

	//logrus.Debug(req)
	//if !ok {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "update access token request error", "text": ok})
	//	return
	//}

	tokens, err := h.userService.UpdateAccessToken(&entities.Tokens{
		AccessTokenString:  accessToken,
		RefreshTokenString: refreshToken,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update access token service error", "text": err.Error()})
		return
	}
	logrus.Debug(tokens)

	resp := dto.UpdateAccessTokenResponse{
		AccessTokenString:  tokens[0].TokenString,
		RefreshTokenString: tokens[1].TokenString,
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}
