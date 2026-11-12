package handler

import (
	"mangahub/internal/models"
	"mangahub/internal/repo"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	libRepo *repo.LibraryRepository
}

func NewLibraryHandler(libRepo * repo.LibraryRepository) *LibraryHandler {
	return &LibraryHandler{libRepo: libRepo}
}

func (h *LibraryHandler) AddToLibrary(c *gin.Context) {
	userId := c.GetInt64("userID")

	var req models.AddLibraryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id and status is required"})
		return 
	}

	checker, err := h.libRepo.ExistsByUserIDAndMangaID(userId, req.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return 
	}
	if checker {
		c.JSON(http.StatusConflict, gin.H{"error": "manga already in library"})
		return 
	}

	id, err := h.libRepo.Create(userId, req.MangaID, req.Status) 
	if err != nil {
		return 
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": userId,
		"manga_id": id,
		"status": req.Status,
	})
}

func (h *LibraryHandler) GetLibrary(c *gin.Context) {
	userID := c.GetInt64("user_id")

	items, err := h.libRepo.GetLibrary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if items == nil {
		items = []models.LibraryItem{}
	}

	c.JSON(http.StatusOK, items)

}