package handler

// controller and service
import (
	"mangahub/internal/models"
	"mangahub/internal/repo"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//generate jwt signiture
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type AuthHandler struct {
	userRepo *repo.UserRepository
}

func NewAuthHandler(userRepo *repo.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
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

	c.JSON(http.StatusCreated, gin.H{"used_id": id, "username": req.Username})
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

	//set the exp timer for the token to be 2 hours per session
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"username": user.Username,
		"exp": time.Now().Add(2 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating token"})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"token": tokenString,
	})
}