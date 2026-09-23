package main

import (
	"backend/database"
	"backend/internal/config"
	"backend/internal/repository"
)

func main() {

	db := database.ConfigDb()
	UserRepo:=repository.NewUserRepository(db)
	
	config.Config()

}
