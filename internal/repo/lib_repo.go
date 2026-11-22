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
	res, err := r.db.Exec("INSERT INTO user_library (user_id, manga_id, status, current_chap, version) VALUES (?, ?, ?, 0, 1)", userID, mangaID, status,)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func(r *LibraryRepository) GetLibrary(userID int64) ([]models.LibraryItem, error) {
	rows, err := r.db.Query("SELECT manga_id, status, current_chap, version FROM user_library WHERE user_id=?", userID)
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
		WHERE user_id = ? AND manga_id = ?
		)
	`, userID, mangaID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//transactional update to handle multiple query happen at the same time 
//trying to change the data
func(r *LibraryRepository) UpdateProgress(userID, mangaID int64, chap int) (int, error) {
	tx, err := r.db.Begin() 
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE user_library 
		SET current_chap = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND manga_id = ?`,
		chap, userID, mangaID,
	)
	if err != nil {
		return 0, err
	}

	rowAffected, _ := res.RowsAffected()
	if rowAffected == 0 {
		return 0, sql.ErrNoRows
	}

	var newVersion int 
	err = tx.QueryRow("SELECT version FROM user_library WHERE user_id = ? AND manga_id = ?", userID, mangaID).
		Scan(&newVersion)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

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