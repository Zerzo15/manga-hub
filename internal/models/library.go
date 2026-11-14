package models

type AddLibraryRequest struct {
	MangaID int64  `json:"manga_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

type UpdateProgressRequest struct {
	CurrentChapter int `json:"current_chapter" binding:"required, min=0"`
}

type LibraryItem struct {
	MangaID        int64  `json:"manga_id"`
	Status         string `json:"status"`
	CurrentChapter int    `json:"current_chapter"`
	Version        int    `json:"version"`
}