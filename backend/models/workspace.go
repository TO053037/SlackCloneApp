package models

import (
	"fmt"

	"backend/config"
)

type Workspace struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	PrimaryOwnerId uint32 `json:"primary_owner_id"`
}

func NewWorkspace(id int, name string, primaryOwnerId uint32) *Workspace {
	return &Workspace{
		ID:             id,
		Name:           name,
		PrimaryOwnerId: primaryOwnerId,
	}
}

func (w *Workspace) SetId() error {
	cmd := fmt.Sprintf("SELECT id FROM %s", config.Config.WorkspaceTableName)
	rows, err := DbConnection.Query(cmd)
	if err != nil {
		return err
	}
	defer rows.Close()
	maxId := 0
	for rows.Next() {
		var id int
		rows.Scan(&id)
		if id > maxId {
			maxId = id
		}
	}
	w.ID = maxId + 1
	return nil
}

func (w *Workspace) Create() error {
	if err := w.SetId(); err != nil {
		return err
	}
	cmd := fmt.Sprintf("INSERT INTO %s (id, name, workspace_primary_owner_id) VALUES ($1, $2, $3)", config.Config.WorkspaceTableName)
	_, err := DbConnection.Exec(cmd, w.ID, w.Name, w.PrimaryOwnerId)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func GetWorkspaceById(id int) (Workspace, error) {
	var w Workspace
	cmd := fmt.Sprintf("SELECT id, name, workspace_primary_owner_id FROM %s WHERE id = $1", config.Config.WorkspaceTableName)
	row := DbConnection.QueryRow(cmd, id)
	err := row.Scan(&w.ID, &w.Name, &w.PrimaryOwnerId)
	return w, err
}

func (w *Workspace) RenameWorkspaceName() error {
	if w.ID == 0 || w.Name == "" {
		return fmt.Errorf("id or newName is empty")
	}
	cmd := fmt.Sprintf("UPDATE %s SET name = $1 WHERE id = $2", config.Config.WorkspaceTableName)
	_, err := DbConnection.Exec(cmd, w.Name, w.ID)
	return err
}