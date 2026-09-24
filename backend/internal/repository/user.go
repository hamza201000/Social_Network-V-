package repository

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
)

func (repo *UserRepository) RegisterUser(user models.RegisterRequest) error {

	qeury := `INSERT INTO users 
	(first_name,last_name,username,email,password_hash)
	VALUES ($1, $2, $3, $4, $5)
	`
	result, err := repo.db.Exec(qeury, user.FirstName, user.LastName, user.Username, user.Email, user.Password)
	if err != nil {
		fmt.Println("err database", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("err in result database", err)
		return err
	}
	fmt.Println("userid:", id)
	return nil
}

func (repo *UserRepository) LoginUser(user models.LoginRequest) (*models.User,error) {
	
	var userData models.User
	qeury := `SELECT id,password_hash
	FROM users
	WHERE (email=$1 or username=$1)
	`
	err := repo.db.QueryRow(qeury, user.Email).Scan(&userData.Id,&userData.Password)
	if err == sql.ErrNoRows {
		fmt.Println("err database", err)
		return nil,err
	}

	return &userData,nil
}
