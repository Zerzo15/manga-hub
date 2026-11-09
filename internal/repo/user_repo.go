package repo

import (
	"database/sql"
	"mangahub/internal/models"
)

type UserRepository struct {
	db *sql.DB	
}

func NewUserRepo(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username string, password string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *UserRepository) ExistsByName(username string) (bool, error) {
	var id int32

	//scan the user with the username
	err := r.db.QueryRow("SELECT id FROM users WHERE username = ?", username).
		Scan(&id)
	if err != nil {
		//no user found but no err was also found
		if err == sql.ErrNoRows {
			return false, nil
		}

		//no user found and something is wrong with the database 
		return false, err
	}

	//query found an ID, meaning the user exists
	return true, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{Username: username}

	err := r.db.QueryRow("SELECT id, password FROM users WHERE username=?", username).
		Scan(&user.ID, &user.Password)
	if err != nil {
		return nil, err
	}	

	return user, nil
}