package models

import (
	"fmt"

	"backend/config"
)

type User struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	PassWord string `json:"password"`
}

func NewUser(id uint32, name, password string) *User {
	return &User{ID: id, Name: name, PassWord: password}
}

func (user *User) Create() error {
	cmd := fmt.Sprintf("INSERT INTO %s (id, name, password) VALUES (?, ?, ?)", config.Config.UserTableName)
	_, err := DbConnection.Exec(cmd, user.ID, user.Name, user.PassWord)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func GetUserById(id uint32) (User, error) {
	cmd := fmt.Sprintf("SELECT id, name, password FROM %s WHERE id = ?", config.Config.UserTableName)
	row := DbConnection.QueryRow(cmd, id)
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.PassWord)
	if err != nil {
		return User{}, err
	}
	if user.Name == "" || user.PassWord == "" {
		err = fmt.Errorf("not found id")
		return User{}, err
	}
	return user, nil
}

func GetUserByNameAndPassword(username, password string) (User, error) {
	cmd := fmt.Sprintf("SELECT id, name, password FROM %s WHERE name = ? AND password = ?", config.Config.UserTableName)
	row := DbConnection.QueryRow(cmd, username, password)
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.PassWord)
	return u, err
}
