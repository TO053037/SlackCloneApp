package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xyproto/randomstring"

	"backend/controllerUtils"
	"backend/models"
)

var channelRouter = SetupRouter()

func createChannelTestFunc(name, description string, isPrivate *bool, jwtToken string, workspaceId int) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ch := controllerUtils.CreateChannelInput{name, description, isPrivate, workspaceId}
	jsonInput, _ := json.Marshal(ch)
	req, err := http.NewRequest("POST", "/api/channel/create", bytes.NewBuffer(jsonInput))
	if err != nil {
		return rr
	}
	req.Header.Set("Authorization", jwtToken)
	channelRouter.ServeHTTP(rr, req)
	return rr
}

func addUserInChannelTestFunc(channelId int, userId uint32, jwtToken string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	jsonInput, _ := json.Marshal(controllerUtils.AddUserInChannelInput{ChannelId: channelId, UserId: userId})
	req, _ := http.NewRequest("POST", "/api/channel/add_user", bytes.NewBuffer(jsonInput))
	req.Header.Set("Authorization", jwtToken)
	channelRouter.ServeHTTP(rr, req)
	return rr
}

func deleteUserFromChannelTestFunc(channelId, workspaceId int, userId uint32, jwtToken string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	cau := models.NewChannelsAndUses(channelId, userId, false)
	jsonInput, _ := json.Marshal(cau)
	req, _ := http.NewRequest("DELETE", "/api/channel/delete_user/"+strconv.Itoa(workspaceId), bytes.NewBuffer(jsonInput))
	req.Header.Set("Authorization", jwtToken)
	channelRouter.ServeHTTP(rr, req)
	return rr
}

func deleteChannelTestFunc(channelId, workspaceId int, jwtToken string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ch := models.NewChannel(channelId, "", "", false, false, workspaceId)
	jsonInput, _ := json.Marshal(ch)
	req, _ := http.NewRequest("DELETE", "/api/channel/delete", bytes.NewBuffer(jsonInput))
	req.Header.Set("Authorization", jwtToken)
	channelRouter.ServeHTTP(rr, req)
	return rr
}

func getChannelsByUserTestFunc(workspaceId int, jwtToken string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/channel/get_by_user_and_workspace/"+strconv.Itoa(workspaceId), nil)
	req.Header.Set("Authorization", jwtToken)
	channelRouter.ServeHTTP(rr, req)
	return rr
}

func TestCreateChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. ??????????????? 200
	// 3. body???????????????????????????????????? 400
	// 4. request??????user????????????workspace?????????????????????????????? 404
	// 5. ????????????????????????channel????????????workspace??????????????????????????? 409

	t.Run("1", func(t *testing.T) {
		userName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		description := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code)

		rr := loginTestFunc(userName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createWorkSpaceTestFunc(workspaceName, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createChannelTestFunc(channelName, description, &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ch := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), ch)
		assert.NotEqual(t, 0, ch.ID)
		assert.Equal(t, channelName, ch.Name)
		assert.Equal(t, isPrivate, ch.IsPrivate)
		assert.Equal(t, false, ch.IsArchive)
	})

	t.Run("3", func(t *testing.T) {
		userName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		description := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code)

		rr := loginTestFunc(userName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createWorkSpaceTestFunc(workspaceName, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createChannelTestFunc("", description, &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"name not found\"}", rr.Body.String())

		rr = createChannelTestFunc(channelName, description, nil, lr.Token, w.ID)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"is_private not found\"}", rr.Body.String())
		rr = createChannelTestFunc("", description, &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"name not found\"}", rr.Body.String())
	})

	t.Run("4", func(t *testing.T) {
		userName := randomstring.EnglishFrequencyString(30)
		createWorkspaceUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		description := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(createWorkspaceUserName, "pass").Code)

		rr := loginTestFunc(createWorkspaceUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createWorkSpaceTestFunc(workspaceName, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = loginTestFunc(userName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		lr = new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createChannelTestFunc(channelName, description, &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"user not found in workspace\"}", rr.Body.String())
	})

	t.Run("5", func(t *testing.T) {
		userName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		description := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code)

		rr := loginTestFunc(userName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createWorkSpaceTestFunc(workspaceName, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createChannelTestFunc(channelName, description, &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		rr = createChannelTestFunc(channelName, "", &isPrivate, lr.Token, w.ID)
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Equal(t, "{\"message\":\"already exist same name channel in workspace\"}", rr.Body.String())
	})
}

func TestAddUserInChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. ??????????????? 200
	// 2. body????????????????????????(channel_id, user_id) 400
	// 3. ?????????????????????user???workspace?????????????????????????????? 404
	// 4. ???????????????user???workspace?????????????????????????????? 404
	// 6. ???????????????user????????????channel?????????????????????????????? 409
	// 7. ?????????????????????user???????????????????????????????????????????????? 403

	t.Run("1", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, alr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		cau := new(models.ChannelsAndUsers)
		json.Unmarshal(([]byte)(byteArray), cau)
		assert.Equal(t, c.ID, cau.ChannelId)
		assert.Equal(t, alr.UserId, cau.UserId)
		assert.ElementsMatch(t, false, cau.IsAdmin)
	})

	t.Run("2", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, alr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = addUserInChannelTestFunc(0, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"channel_id not found\"}", rr.Body.String())

		rr = addUserInChannelTestFunc(c.ID, 0, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"user_id not found\"}", rr.Body.String())
	})

	t.Run("3", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		createWorkspaceUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(createWorkspaceUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = loginTestFunc(createWorkspaceUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = createWorkSpaceTestFunc(workspaceName, clr.Token, clr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, alr.UserId, clr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, clr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"request user not found in workspace\"}", rr.Body.String())

	})

	t.Run("4", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"added user not found in workspace\"}", rr.Body.String())
	})

	t.Run("6", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, alr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token).Code)

		rr = addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusConflict, rr.Code)

		assert.Equal(t, "{\"message\":\"already exist user in channel\"}", rr.Body.String())
	})

	t.Run("7", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		addUserName := randomstring.EnglishFrequencyString(30)
		createChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(addUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(createChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(addUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		alr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), alr)

		rr = loginTestFunc(createChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, alr.UserId, rlr.Token).Code)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, clr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, clr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = addUserInChannelTestFunc(c.ID, rlr.UserId, clr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)

		rr = addUserInChannelTestFunc(c.ID, alr.UserId, rlr.Token)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "{\"message\":\"no permission adding user in channel\"}", rr.Body.String())
	})
}

func TestDeleteUserFromChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. ???????????????(private channel) 200
	// 2. ???????????????(public channel) 200
	// 3. body????????????????????????(channel_id, user_id) 400
	// 4. delete?????????user???workspace?????????????????? 404
	// 5. request??????user???workspace?????????????????? 404
	// 6. channel???????????????????????? 404
	// 7. channel???workspace???????????????????????? 404
	// 8. channel???name???general????????? 400
	// 9. delete?????????user???channel?????????????????? 404
	// 10. delete?????????????????????user?????????????????????????????????(private channel) 403
	// 11. delete?????????????????????user?????????????????????????????????(public channel) 403
	// 12. channel??????????????????????????????????????? 400

	t.Run("1", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, rlr.Token).Code)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		cau := new(models.ChannelsAndUsers)
		json.Unmarshal(([]byte)(byteArray), cau)
		assert.Equal(t, c.ID, cau.ChannelId)
		assert.Equal(t, dlr.UserId, cau.UserId)
	})

	t.Run("2", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, rlr.Token).Code)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		cau := new(models.ChannelsAndUsers)
		json.Unmarshal(([]byte)(byteArray), cau)
		assert.Equal(t, c.ID, cau.ChannelId)
		assert.Equal(t, dlr.UserId, cau.UserId)
	})

	t.Run("3", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, rlr.Token).Code)

		rr = deleteUserFromChannelTestFunc(0, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"user_id or channel_id not found\"}", rr.Body.String())
		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, 0, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"user_id or channel_id not found\"}", rr.Body.String())
	})

	t.Run("4", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"user not found in workspace\"}", rr.Body.String())
	})

	t.Run("5", func(t *testing.T) {
		createChannelUserName := randomstring.EnglishFrequencyString(30)
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(createChannelUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(createChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, clr.Token, clr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, clr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, clr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, clr.Token).Code)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"request user not found in workspace\"}", rr.Body.String())
	})

	t.Run("6", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, rlr.Token).Code)

		rr = deleteUserFromChannelTestFunc(-1, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"sql: no rows in result set\"}", rr.Body.String())
	})

	t.Run("7", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		workspaceName2 := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createWorkSpaceTestFunc(workspaceName2, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w2 := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w2)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w2.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"channel not found in workspace\"}", rr.Body.String())
	})

	t.Run("8", func(t *testing.T) {
		// TODO ??????????????????????????????????????????????????????

		// createWorkspace func????????????general channel????????????????????????????????????????????????ID????????????????????????????????????DB??????????????????????????????????????????

		// requestUserName := "testDeleteUserFromChannelRequestUserName8"
		// deleteUserName := "testDeleteUserFromChannelDeleteUserName8"
		// workspaceName := "testDeleteUserFromChannelWorkspaceName8"
		// channelName := "general"

		// assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		// assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		// rr := loginTestFunc(requestUserName, "pass")
		// assert.Equal(t, http.StatusOK, rr.Code)
		// byteArray, _ := ioutil.ReadAll(rr.Body)
		// rlr := new(LoginResponse)
		// json.Unmarshal(([]byte)(byteArray), rlr)

		// rr = loginTestFunc(deleteUserName, "pass")
		// assert.Equal(t, http.StatusOK, rr.Code)
		// byteArray, _ = ioutil.ReadAll(rr.Body)
		// dlr := new(LoginResponse)
		// json.Unmarshal(([]byte)(byteArray), dlr)

		// rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		// assert.Equal(t, http.StatusOK, rr.Code)
		// byteArray, _ = ioutil.ReadAll(rr.Body)
		// w := new(models.Workspace)
		// json.Unmarshal(([]byte)(byteArray), w)

		// assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		// rr = createChannelTestFunc(channelName, "des", true, rlr.Token, w.ID)
		// assert.Equal(t, http.StatusOK, rr.Code)
		// byteArray, _ = ioutil.ReadAll(rr.Body)
		// c := new(models.Channel)
		// json.Unmarshal(([]byte)(byteArray), c)

		// assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token).Code)

		// rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		// assert.Equal(t, http.StatusBadRequest, rr.Code)
		// assert.Equal(t, "{\"message\":\"don't delete general channel\"}", rr.Body.String())
	})

	t.Run("9", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"user not found in channel\"}", rr.Body.String())
	})

	t.Run("10", func(t *testing.T) {
		createChannelUserName := randomstring.EnglishFrequencyString(30)
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(createChannelUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(createChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, clr.Token, clr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, clr.Token).Code)
		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, rlr.UserId, clr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, clr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, clr.Token).Code)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "{\"message\":\"not permission deleting user in channel\"}", rr.Body.String())
	})

	t.Run("11", func(t *testing.T) {
		createChannelUserName := randomstring.EnglishFrequencyString(30)
		requestUserName := randomstring.EnglishFrequencyString(30)
		deleteUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(createChannelUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(deleteUserName, "pass").Code)

		rr := loginTestFunc(createChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(deleteUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		dlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), dlr)

		rr = createWorkSpaceTestFunc(workspaceName, clr.Token, clr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, dlr.UserId, clr.Token).Code)
		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, rlr.UserId, clr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, clr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, dlr.UserId, clr.Token).Code)

		rr = deleteUserFromChannelTestFunc(c.ID, w.ID, dlr.UserId, rlr.Token)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "{\"message\":\"not permission deleting user in channel\"}", rr.Body.String())
	})

	// TODO test 12 ????????????????????????????????????

}

func TestDeleteChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. ??????????????? 200
	// 2. body?????????????????????????????????????????????(channel_id, workspace_id) 400
	// 3. request??????user???workspace?????????????????????????????? 404
	// 4. request??????user???delete??????????????????????????? 403
	// 5. channel???????????????????????? 404
	// 6. channel???workspace???????????????????????? 404

	t.Run("1", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(c.ID, w.ID, rlr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ch := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), ch)

		assert.Equal(t, c.ID, ch.ID)
		assert.Equal(t, w.ID, ch.WorkspaceId)
	})

	t.Run("2", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(0, w.ID, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"channel_id or workspace_id not found\"}", rr.Body.String())

		rr = deleteChannelTestFunc(c.ID, 0, rlr.Token)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "{\"message\":\"channel_id or workspace_id not found\"}", rr.Body.String())
	})

	t.Run("3", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		createChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(createChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = loginTestFunc(createChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		clr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), clr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(c.ID, -1, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"sql: no rows in result set\"}", rr.Body.String())
	})

	t.Run("4", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(c.ID, w.ID, ilr.Token)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "{\"message\":\"no permission deleting channel\"}", rr.Body.String())
	})

	t.Run("5", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(-1, w.ID, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"sql: no rows in result set\"}", rr.Body.String())
	})

	t.Run("6", func(t *testing.T) {
		requestUserName := randomstring.EnglishFrequencyString(30)
		inChannelUserName := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		workspaceName2 := randomstring.EnglishFrequencyString(30)
		channelName := randomstring.EnglishFrequencyString(30)
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(requestUserName, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(inChannelUserName, "pass").Code)

		rr := loginTestFunc(requestUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		rlr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), rlr)

		rr = loginTestFunc(inChannelUserName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		ilr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), ilr)

		rr = createWorkSpaceTestFunc(workspaceName, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		rr = createWorkSpaceTestFunc(workspaceName2, rlr.Token, rlr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w2 := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w2)

		assert.Equal(t, http.StatusOK, addUserWorkspaceTestFunc(w2.ID, 4, ilr.UserId, rlr.Token).Code)

		rr = createChannelTestFunc(channelName, "des", &isPrivate, rlr.Token, w2.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		c := new(models.Channel)
		json.Unmarshal(([]byte)(byteArray), c)

		assert.Equal(t, http.StatusOK, addUserInChannelTestFunc(c.ID, ilr.UserId, rlr.Token).Code)

		rr = deleteChannelTestFunc(c.ID, w.ID, rlr.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"sql: no rows in result set\"}", rr.Body.String())
	})
}

func TestGetChannelsByUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. general channel?????????channel????????????????????? 200
	// 2. user???workspace?????????????????????????????? 404

	t.Run("1 ??????????????????????????????", func(t *testing.T) {
		channelCount := 10
		userName := randomstring.EnglishFrequencyString(30)
		workspaceName1 := randomstring.EnglishFrequencyString(30)
		workspaceName2 := randomstring.EnglishFrequencyString(30)
		isPrivate := true
		channelNames1 := make([]string, channelCount)
		channelNames2 := make([]string, channelCount)

		for i := 0; i < channelCount; i++ {
			channelNames1[i] = randomstring.EnglishFrequencyString(30)
			channelNames2[i] = randomstring.EnglishFrequencyString(30)
		}

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code)

		rr := loginTestFunc(userName, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr)

		rr = createWorkSpaceTestFunc(workspaceName1, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w1 := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w1)

		rr = createWorkSpaceTestFunc(workspaceName2, lr.Token, lr.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w2 := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w2)

		channelIds := make([]int, channelCount)
		for i := 0; i < channelCount; i++ {
			rr = createChannelTestFunc(channelNames1[i], "des", &isPrivate, lr.Token, w1.ID)
			assert.Equal(t, http.StatusOK, rr.Code)
			byteArray, _ = ioutil.ReadAll(rr.Body)
			ch := new(models.Channel)
			json.Unmarshal(([]byte)(byteArray), ch)
			channelIds[i] = ch.ID

			assert.Equal(t, http.StatusOK, createChannelTestFunc(
				channelNames2[i],
				"des",
				&isPrivate,
				lr.Token,
				w2.ID,
			).Code)
		}

		rr = getChannelsByUserTestFunc(w1.ID, lr.Token)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		chs := make([]models.Channel, 0)
		json.Unmarshal(([]byte)(byteArray), &chs)
		assert.Equal(t, channelCount+1, len(chs))
		for _, ch := range chs {
			if ch.Name == "general" {
				continue
			}
			assert.Contains(t, channelIds, ch.ID)
			assert.Contains(t, channelNames1, ch.Name)
		}
	})

	t.Run("2 user???workspace??????????????????????????????", func(t *testing.T) {
		channelCount := 10
		userName1 := randomstring.EnglishFrequencyString(30)
		userName2 := randomstring.EnglishFrequencyString(30)
		workspaceName := randomstring.EnglishFrequencyString(30)
		isPrivate := true
		channelNames := make([]string, channelCount)

		for i := 0; i < channelCount; i++ {
			channelNames[i] = randomstring.EnglishFrequencyString(30)
		}

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName1, "pass").Code)
		assert.Equal(t, http.StatusOK, signUpTestFunc(userName2, "pass").Code)

		rr := loginTestFunc(userName1, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ := ioutil.ReadAll(rr.Body)
		lr1 := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr1)

		rr = loginTestFunc(userName2, "pass")
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		lr2 := new(LoginResponse)
		json.Unmarshal(([]byte)(byteArray), lr2)

		rr = createWorkSpaceTestFunc(workspaceName, lr1.Token, lr1.UserId)
		assert.Equal(t, http.StatusOK, rr.Code)
		byteArray, _ = ioutil.ReadAll(rr.Body)
		w := new(models.Workspace)
		json.Unmarshal(([]byte)(byteArray), w)

		channelIds := make([]int, channelCount)
		for i := 0; i < channelCount; i++ {
			rr = createChannelTestFunc(channelNames[i], "des", &isPrivate, lr1.Token, w.ID)
			assert.Equal(t, http.StatusOK, rr.Code)
			byteArray, _ = ioutil.ReadAll(rr.Body)
			ch := new(models.Channel)
			json.Unmarshal(([]byte)(byteArray), ch)
			channelIds[i] = ch.ID
		}

		rr = getChannelsByUserTestFunc(w.ID, lr2.Token)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "{\"message\":\"request user not found in workspace\"}", rr.Body.String())
	})
}
