package repository

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
)

func RegisterUser(db *sql.DB, user models.RegisterRequest)error {

	qeury := `INSERT INTO TABLE users 
	(first_name,last_name,username,email,pasword_hash)
	VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.Exec(qeury, user.FirstName, user.LastName, user.Username, user.Email, user.Password)
	if err != nil {
		fmt.Println("err database", err)
		return err
	}
	id ,err:= result.LastInsertId()
	if err != nil {
		fmt.Println("err in result database", err)
		return err
	}
	fmt.Println("userid:", id)
	return  nil
}
