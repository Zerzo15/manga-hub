package handler

import (
	"database/sql"
	"mangahub/internal/models"
	"mangahub/internal/repo"
	"net/http"
	"strconv"

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

func (h *LibraryHandler) UpdateProgress(c *gin.Context) {
	userID := c.GetInt("user_id")

	mangaIDStr := c.Param("mangaId")
	mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}
	

	var req models.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current chapter is required"})
		return
	}

	newVersion, err := h.libRepo.UpdateProgress(int64(userID), mangaID, req.CurrentChapter)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "progress updated",
		"manga_id":        mangaID,
		"current_chapter": req.CurrentChapter,
		"version":         newVersion,
	})
}

func (h *LibraryHandler) DeleteManga(c *gin.Context) {
	userID := c.GetInt("user_id")

	mangaIDStr := c.Param("mangaId")
	mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga ID format"})
		return
	}

	if err := h.libRepo.DeleteManga(int64(userID), mangaID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "manga deleted",
		"manga_id": mangaID,
	})
}