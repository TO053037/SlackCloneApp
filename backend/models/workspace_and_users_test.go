package models

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspaceAndUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 10; i++ {
		workspaceId := rand.Int()
		for j := 0; j < 10; j++ {
			userId := rand.Uint32()
			roleId := j%4 + 1
			wau := NewWorkspaceAndUsers(workspaceId, userId, roleId)
			assert.Equal(t, workspaceId, wau.WorkspaceId)
			assert.Equal(t, userId, wau.UserId)
			assert.Equal(t, roleId, wau.RoleId)
		}
	}
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 10; i++ {
		workspaceId := rand.Int()
		for j := 0; j < 10; j++ {
			userId := rand.Uint32()
			roleId := j%4 + 1
			wau := NewWorkspaceAndUsers(workspaceId, userId, roleId)
			assert.Empty(t, wau.Create())
			assert.NotEmpty(t, wau.Create())
			wau = NewWorkspaceAndUsers(workspaceId, userId, (roleId+1)%4)
			assert.NotEmpty(t, wau.Create())
		}
	}
}

func TestGetWorkspaceAndUserByWorkspaceIdAndUserId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 10; i++ {
		workspaceId := rand.Int()
		for j := 0; j < 10; j++ {
			userId := rand.Uint32()
			roleId := j%4 + 1
			wau := NewWorkspaceAndUsers(workspaceId, userId, roleId)
			wau.Create()
			getWau, err := GetWorkspaceAndUserByWorkspaceIdAndUserId(workspaceId, userId)
			assert.Empty(t, err)
			assert.Equal(t, workspaceId, getWau.WorkspaceId)
			assert.Equal(t, userId, getWau.UserId)
			assert.Equal(t, roleId, getWau.RoleId)

			_, err = GetWorkspaceAndUserByWorkspaceIdAndUserId(rand.Int(), userId)
			assert.NotEmpty(t, err)
		}
	}
}

func TestDeleteWorkspaceAndUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 10; i++ {
		workspaceId := rand.Int()
		for j := 0; j < 10; j++ {
			userId := rand.Uint32()
			roleId := j%4 + 1
			wau := NewWorkspaceAndUsers(workspaceId, userId, roleId)
			assert.Empty(t, wau.Create())
			_, err := GetWorkspaceAndUserByWorkspaceIdAndUserId(wau.WorkspaceId, wau.UserId)
			assert.Empty(t, err)
			err = wau.DeleteWorkspaceAndUser()
			assert.Empty(t, err)
			_, err = GetWorkspaceAndUserByWorkspaceIdAndUserId(wau.WorkspaceId, wau.UserId)
			assert.NotEmpty(t, err)
			err = wau.DeleteWorkspaceAndUser()
			assert.Empty(t, err)
		}
	}
}

func TestGetRoleIdByWorkspaceIdAndUserId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wau := NewWorkspaceAndUsers(rand.Int(), rand.Uint32(), 3)
	wau.Create()
	roleId, err := GetRoleIdByWorkspaceIdAndUserId(wau.WorkspaceId, wau.UserId)
	assert.Equal(t, wau.RoleId, roleId)
	assert.Empty(t, err)

	_, err = GetRoleIdByWorkspaceIdAndUserId(rand.Int(), wau.UserId)
	assert.NotEmpty(t, err)
}

func TestGetWAUsByUserId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("1 ??????????????????????????????", func(t *testing.T) {
		testCases := 10
		userId := rand.Uint32()
		workspaceIds := make([]int, testCases)
		for i := 0; i < testCases; i++ {
			wau := NewWorkspaceAndUsers(int(rand.Uint64()), userId, rand.Int()%4+1)
			assert.Empty(t, wau.Create())
			workspaceIds[i] = wau.WorkspaceId
		}
		res, err := GetWAUsByUserId(userId)
		assert.Empty(t, err)
		assert.Equal(t, testCases, len(res))
		for _, wau := range res {
			assert.Equal(t, userId, wau.UserId)
			assert.Contains(t, workspaceIds, wau.WorkspaceId)
			assert.NotEqual(t, 0, wau.RoleId)
		}
	})
	t.Run("2 ?????????????????????????????????", func(t *testing.T) {
		res, err := GetWAUsByUserId(rand.Uint32())
		assert.Empty(t, err)
		assert.Equal(t, 0, len(res))
	})
}

func TestGetWAUsByWorkspaceId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("1 ??????????????????????????????", func(t *testing.T) {
		testCases := 10
		workspaceId := int(rand.Uint64())
		waus := make([]WorkspaceAndUsers, testCases)
		for i := 0; i < testCases; i++ {
			wau := NewWorkspaceAndUsers(workspaceId, rand.Uint32(), rand.Int()%4+1)
			assert.Empty(t, wau.Create())
			waus[i] = *wau
		}
		res, err := GetWAUsByWorkspaceId(workspaceId)
		assert.Empty(t, err)
		assert.Equal(t, testCases, len(res))
		for _, wau := range waus {
			assert.Contains(t, res, wau)
		}
	})

	t.Run("2 ???????????????????????????????????????", func(t *testing.T) {
		res, err := GetWAUsByWorkspaceId(int(rand.Uint64()))
		assert.Empty(t, err)
		assert.Equal(t, 0, len(res))
	})
}
