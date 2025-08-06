package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateToken(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"uid": userID,
		"exp": time.Now().Add(time.Duration(config.JWT.Expires) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWT.Secret))
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	
	// 检查用户名是否已存在
	var count int64
	DB.Model(&User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}
	
	// 检查邮箱是否已存在（如果提供了邮箱）
	if req.Email != "" {
		DB.Model(&User{}).Where("email = ?", req.Email).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被注册"})
			return
		}
	}
	
	// 检查手机号是否已存在（如果提供了手机号）
	if req.Phone != "" {
		DB.Model(&User{}).Where("phone = ?", req.Phone).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "手机号已被注册"})
			return
		}
	}
	
	hash, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	
	user := User{
		Username: req.Username,
		Password: hash,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}
	
	if err := DB.Create(&user).Error; err != nil {
		// 检查是否是唯一性约束错误
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "username") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
			} else if strings.Contains(err.Error(), "email") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被注册"})
			} else if strings.Contains(err.Error(), "phone") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "手机号已被注册"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "注册信息已存在"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败: " + err.Error()})
		}
		return
	}
	
	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Token失败"})
		return
	}
	
	c.JSON(http.StatusOK, AuthResponse{Token: token})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	if !checkPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}
	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Token失败"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Token: token})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JWT.Secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token无效"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token无效"})
			return
		}
		c.Set("uid", uint64(claims["uid"].(float64)))
		c.Next()
	}
}
