package models

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"backend/config"
)

func TestNewWorkspace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 1000; i++ {
		id := i
		name := "testNewWorkspaceName" + strconv.Itoa(i)
		primary_owner_id := uint32(i) + 10
		w := NewWorkspace(id, name, primary_owner_id)
		assert.Equal(t, id, w.ID)
		assert.Equal(t, name, w.Name)
		assert.Equal(t, primary_owner_id, w.PrimaryOwnerId)
	}
}

func TestCreateWorkspace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// 正常な場合
	numbersOfTests := 1
	names := make([]string, numbersOfTests)
	primaryOwnerIds := make([]uint32, numbersOfTests)
	for i := 0; i < numbersOfTests; i++ {
		names[i] = "testCreateWorkspace" + strconv.Itoa(i)
		primaryOwnerIds[i] = uint32(i) + 11
	}

	for i := 0; i < numbersOfTests; i++ {
		w := NewWorkspace(0, names[i], primaryOwnerIds[i])
		err := w.CreateWorkspace()
		assert.Empty(t, err)
	}

	cmd := fmt.Sprintf("SELECT id, name, workspace_primary_owner_id FROM %s WHERE name = ?", config.Config.WorkspaceTableName)
	for i := 0; i < numbersOfTests; i++ {
		row := DbConnection.QueryRow(cmd, names[i])
		var w Workspace
		err := row.Scan(&w.ID, &w.Name, &w.PrimaryOwnerId)
		assert.Empty(t, err)
		assert.NotEqual(t, 0, w.ID)
		assert.Equal(t, names[i], w.Name)
		assert.Equal(t, primaryOwnerIds[i], w.PrimaryOwnerId)
	}

	// workspace nameが既に存在する場合 error
	name := "testCreateWorkspaceDuplicate"
	w := NewWorkspace(0, name, uint32(2))
	err := w.CreateWorkspace()
	assert.Empty(t, err)
	w2 := NewWorkspace(0, name, uint32(1))
	err = w2.CreateWorkspace()
	assert.NotEmpty(t, err)
}

func TestRenameWorkspaceName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	w := NewWorkspace(3333, "old name", 3)
	w.CreateWorkspace()
	w.Name = "new name"
	err := w.RenameWorkspaceName()
	assert.Empty(t, err)
	w2, err := GetWorkspaceByName("new name")
	assert.Empty(t, err)
	assert.Equal(t, w2.ID, w.ID)
	assert.Equal(t, w2.Name, w.Name)
	assert.Equal(t, w2.PrimaryOwnerId, w.PrimaryOwnerId)
}

func TestIsExistWorkspaceAndUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	w := NewWorkspace(0, "testIsExistWorkspaceAndUser", 4)
	w.CreateWorkspace()
	assert.Equal(t, true, IsExistWorkspaceById(w.ID))
	assert.Equal(t, false, IsExistWorkspaceById(-1))
}
