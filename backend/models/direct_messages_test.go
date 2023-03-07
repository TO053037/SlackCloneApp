package models

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xyproto/randomstring"
)

func TestCreateDirectMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	dm := NewDirectMessage(randomstring.EnglishFrequencyString(99), rand.Uint32(), uint(rand.Uint64()))
	res := dm.Create()
	assert.NotEqual(t, uint(0), dm.ID)
	assert.Empty(t, res.Error)
}

func TestGetAllDMsByDLId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	
	t.Run("1 データが存在する場合", func(t *testing.T) {
		dmCount := 4
		dms := make([]*DirectMessage, dmCount)
		sendUserId1 := rand.Uint32()
		sendUserId2 := rand.Uint32()
		dlId := uint(rand.Uint64())
		
		for i := 0; i < dmCount; i ++ {
			if i % 2 == 0 {
				dms[i] = NewDirectMessage(randomstring.EnglishFrequencyString(40), sendUserId1, dlId)
			} else {
				dms[i] = NewDirectMessage(randomstring.EnglishFrequencyString(40), sendUserId2, dlId)
			}
		}

		for _, dm := range dms {
			assert.Empty(t, dm.Create().Error)
		}

		res, err := GetAllDMsByDLId(dlId)
		assert.Empty(t, err)
		assert.Equal(t, dmCount, len(res))

		for i := 0; i < dmCount - 1; i ++ {
			assert.False(t, res[i].CreatedAt.Before(res[i+1].CreatedAt))
		}
	})
	
	t.Run("2 データが存在しない場合", func(t *testing.T) {
		res, err := GetAllDMsByDLId(uint(rand.Uint64()))
		assert.Empty(t, err)
		assert.Empty(t, []DirectMessage{}, res)
	})
}
