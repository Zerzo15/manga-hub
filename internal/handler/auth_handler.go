package handler

// controller and service
import (
	"net/http"

	"mangahub/internal/auth"
	"mangahub/internal/models"
	"mangahub/internal/repo"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo   *repo.UserRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(userRepo *repo.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password is required"})
		return
	}

	checker, err := h.userRepo.ExistsByName(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if checker {
		c.JSON(http.StatusConflict, gin.H{"error": "username in used"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hashed the password"})
		return
	}

	id, err := h.userRepo.Create(req.Username, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "registering error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": id, "username": req.Username})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password is required"})
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong username or password"})
		return
	}

	tokenString, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"token":    tokenString,
	})
}
