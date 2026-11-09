package handler

// controller and service
import (
	"mangahub/internal/models"
	"mangahub/internal/repo"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repo.UserRepository
}

func NewAuthHandler(userRepo *repo.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.LoginRequest 

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

	c.JSON(http.StatusCreated, gin.H{"used_id": id, "username": req.Username})
}
