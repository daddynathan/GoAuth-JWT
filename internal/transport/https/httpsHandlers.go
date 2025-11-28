package https

import (
	"errors"
	"friend-help/internal/errs"
	"friend-help/internal/model"
	"friend-help/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type HTTPHandlers struct {
	AuthService *service.AuthService
}

func NewHTTPHandlers(AuthService *service.AuthService) *HTTPHandlers {
	return &HTTPHandlers{
		AuthService: AuthService,
	}
}

// @Summary      Регистрация нового пользователя
// @Description  Создает нового пользователя, хеширует пароль, сохраняет в БД и выдает JWT-токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body model.AuthRegReq true "Данные для регистрации (login, email, password)"
// @Success      201  {object}  map[string]interface{} "Успешное создание ресурса и выдан токен"
// @Failure      400  {object}  map[string]interface{} "Некорректный JSON, невалидные поля (min length, email format)"
// @Failure      409  {object}  map[string]interface{} "Пользователь с таким логином или email уже существует (errs.ErrUserExists)"
// @Failure      500  {object}  map[string]interface{} "Внутренняя ошибка сервера (ошибка БД, хеширования пароля, генерации токена)"
// @Router       /auth/reg [post]
func (h *HTTPHandlers) HandlerReg(c *gin.Context) {
	var req model.AuthRegReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input format or missing fields"})
		return
	}
	userID, token, err := h.AuthService.RegNewUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, errs.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user_id": userID,
		"token":   token,
		"message": "Registration successful",
	})
}

// @Summary      Вход пользователя
// @Description  Аутентифицирует пользователя по логину/email и паролю, выдает новый JWT-токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body model.AuthLogReq true "Данные для входа (identifier: login/email, password)"
// @Success      200  {object}  map[string]interface{} "Успешный вход и выдан токен"
// @Failure      400  {object}  map[string]interface{} "Некорректный JSON или невалидные символы в идентификаторе"
// @Failure      401  {object}  map[string]interface{} "Неверный логин/email или пароль"
// @Failure      500  {object}  map[string]interface{} "Внутренняя ошибка сервера (ошибка БД, сравнения хеша, генерации токена)"
// @Router       /auth/login [post]
func (h *HTTPHandlers) HandlerLogin(c *gin.Context) {
	var req model.AuthLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input format or missing fields"})
		return
	}
	user, token, err := h.AuthService.Authenticate(c.Request.Context(), req.Identifier, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidLoginOrPass) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or password"})
			return
		}
		if errors.Is(err, errs.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or password"})
			return
		}
		if errors.Is(err, errs.ErrInvalidLoginChars) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process login"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"token":   token,
		"message": "Login successful",
	})
}

// @Summary      Выход из системы
// @Description  Добавляет переданный JWT-токен в чёрный список до истечения его срока действия. Токен должен быть передан в заголовке Authorization как Bearer-токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Успешный выход из системы"
// @Failure      401  {object}  map[string]interface{}  "Отсутствует или неверный заголовок Authorization"
// @Failure      500  {object}  map[string]interface{}  "Внутренняя ошибка сервера (ошибка взаимодействия с Redis)"
// @Router       /auth/logout [post]
func (h *HTTPHandlers) HandlerLogout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
		return
	}
	tokenString := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = authHeader[7:]
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
		return
	}
	err := h.AuthService.Logout(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(500, gin.H{"error": "logout failed"})
		return
	}
	c.JSON(200, gin.H{"message": "logged out"})
}

// @Summary      Получить профиль пользователя (бета покачто)
// @Description  Возвращает информацию о текущем авторизованном пользователе (из JWT claims)
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Данные профиля"
// @Failure      401 {object} map[string]interface{} "Токен отсутствует, недействителен или отозван"
// @Router       /user/profile [get]
func (h *HTTPHandlers) HandlerGetProfile(c *gin.Context) {
	claims, ok := GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": claims.UserID,
		"role":    claims.Role,
	})
}
