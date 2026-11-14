package repo

import (
	"database/sql"
	"mangahub/internal/models"
)

type LibraryRepository struct {
	db *sql.DB	
}

func NewLibraryRepo(db *sql.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

func(r *LibraryRepository) Create(userID, mangaID int64, status string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO user_library (user_id, manga_id, status, current_chapter, version) VALUES (?, ?, ?, 0, 1)", userID, mangaID, status,)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func(r *LibraryRepository) GetLibrary(userID int64) ([]models.LibraryItem, error) {
	rows, err := r.db.Query("SELECT manga_id, status, current_chapter, version FROM user_library WHERE user_id=?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.LibraryItem
	for rows.Next() {
		var item models.LibraryItem 
		err := rows.Scan(
			&item.MangaID, &item.Status, &item.CurrentChapter, &item.Version)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return items, err
	}

	return items, nil
}

func(r *LibraryRepository) ExistsByUserIDAndMangaID(userID, mangaID int64) (bool, error) {
	var exists bool

	err := r.db.QueryRow(` 
		SELECT EXISTS (
		SELECT 1
		FROM user_library 
		WHERE user_id = ? AND manga = ?
		)
	`, userID, mangaID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func(r *LibraryRepository) UpdateProgress(userID, mangaID int64, chapter int) (int, error) {
	res, err := r.db.Exec(`
		UPDATE user_library 
		SET current_chapter = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND manga_id = ?`,
		chapter, userID, mangaID,
	)
	if err != nil {
		return 0, err
	}

	rowAffected, _ := res.RowsAffected()
	if rowAffected == 0 {
		return 0, sql.ErrNoRows
	}

	var newVersion int 
	err = r.db.QueryRow("SELECT version FROM user_id = ?, manga_id = ?", userID, mangaID).Scan(&newVersion)

	return newVersion, nil
}

func(r *LibraryRepository) DeleteManga(userID, mangaID int64) error {
	res, err := r.db.Exec(`
		DELETE FROM user_library 
		WHERE user_id = ? AND manga_id = ?`, 
		userID, mangaID)
	if err != nil {
		return err
	}

	rowAffected, _ := res.RowsAffected()
	if rowAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}