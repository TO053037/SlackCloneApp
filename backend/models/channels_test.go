package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChannel(t *testing.T) {
	name := "testCreateChannelModelName"
	description := "testCreateChannelModelDescription"
	is_private := true
	is_archive := false
	c := NewChannel(0, name, description, is_private, is_archive)
	assert.Empty(t, c.Create())
}

func TestGetChannelById(t *testing.T) {
	name := "testGetChannelByIdName"
	description := "testGetChannelByIdDescription"
	is_private := true
	is_archive := false
	c := NewChannel(0, name, description, is_private, is_archive)
	assert.Empty(t, c.Create())
	assert.NotEqual(t, 0, c.ID)
	c2, err := GetChannelById(c.ID)
	assert.Empty(t, err)
	assert.Equal(t, *c, c2)

	_, err = GetChannelById(-1)
	assert.NotEmpty(t, err)
}

func TestDeleteChannel(t *testing.T) {
	c := NewChannel(0, "testDeleteChannelName", "", true, false)
	assert.Empty(t, c.Create())
	channelId := c.ID
	assert.Empty(t, c.Delete())

	_, err := GetChannelById(channelId)
	assert.NotEmpty(t, err)
}