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

	// 1. 正常な場合 200
	// 3. bodyの情報が不足している場合 400
	// 4. requestしたuserが対象のworkspaceに所属していない場合 404
	// 5. すでに同じ名前のchannelが対象のworkspaceに存在している場合 409

	t.Run("1", func(t *testing.T) {
		userName := "testCreateChannelUserName1"
		workspaceName := "testCreateChannelWorkspaceName1"
		channelName := "testCreateChannelName1"
		description := "description1"
		isPrivate := true

		assert.Equal(t, http.StatusOK, signUpTestFunc(userName, "pass").Code, "OK")

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
		userName := "testCreateChannelUserName3"
		workspaceName := "testCreateChannelWorkspaceName3"
		channelName := "testCreateChannelName3"
		description := "description3"
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
		userName := "testCreateChannelUserName4"
		createWorkspaceUserName := "testCreateChannelCreateWorkspaceUserName4"
		workspaceName := "testCreateChannelWorkspaceName4"
		channelName := "testCreateChannelName4"
		description := "description4"
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
		userName := "testCreateChannelUserName5"
		workspaceName := "testCreateChannelWorkspaceName5"
		channelName := "testCreateChannelName5"
		description := "description5"
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

	// 1. 正常な場合 200
	// 2. bodyに不足がある場合(channel_id, user_id) 400
	// 3. リクエストしたuserがworkspaceに参加していない場合 404
	// 4. 追加されるuserがworkspaceに参加していない場合 404
	// 6. 追加されるuserが対象のchannelに既に存在してる場合 409
	// 7. リクエストしたuserにチャンネルの管理権限がない場合 403

	t.Run("1", func(t *testing.T) {
		requestUserName := "testAddUserInChannelRequestUserName1"
		addUserName := "testAddUserInChannelAddUserName1"
		workspaceName := "testAddUserInChannelWorkspaceName1"
		channelName := "testAddUserInChannelChannelName1"
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
		requestUserName := "testAddUserInChannelRequestUserName2"
		addUserName := "testAddUserInChannelAddUserName2"
		workspaceName := "testAddUserInChannelWorkspaceName2"
		channelName := "testAddUserInChannelChannelName2"
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
		requestUserName := "testAddUserInChannelRequestUserName3"
		addUserName := "testAddUserInChannelAddUserName3"
		createWorkspaceUserName := "testAddUserInChannelCreateWorkspaceUserName3"
		workspaceName := "testAddUserInChannelWorkspaceName3"
		channelName := "testAddUserInChannelChannelName3"
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
		requestUserName := "testAddUserInChannelRequestUserName4"
		addUserName := "testAddUserInChannelAddUserName4"
		workspaceName := "testAddUserInChannelWorkspaceName4"
		channelName := "testAddUserInChannelChannelName4"
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
		requestUserName := "testAddUserInChannelRequestUserName6"
		addUserName := "testAddUserInChannelAddUserName6"
		workspaceName := "testAddUserInChannelWorkspaceName6"
		channelName := "testAddUserInChannelChannelName6"
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
		requestUserName := "testAddUserInChannelRequestUserName7"
		addUserName := "testAddUserInChannelAddUserName7"
		createChannelUserName := "testAddUserInChannelCreateUserName7"
		workspaceName := "testAddUserInChannelWorkspaceName7"
		channelName := "testAddUserInChannelChannelName7"
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

	// 1. 正常な場合(private channel) 200
	// 2. 正常な場合(public channel) 200
	// 3. bodyに不足がある場合(channel_id, user_id) 400
	// 4. deleteされるuserがworkspaceにいない場合 404
	// 5. requestしたuserがworkspaceにいない場合 404
	// 6. channelが存在しない場合 404
	// 7. channelがworkspaceに存在しない場合 404
	// 8. channelのnameがgeneralの場合 400
	// 9. deleteされるuserがchannelにいない場合 404
	// 10. deleteする権限がないuserからのリクエストの場合(private channel) 403
	// 11. deleteする権限がないuserからのリクエストの場合(public channel) 403
	// 12. channelがアーカイブされている場合 400

	t.Run("1", func(t *testing.T) {
		requestUserName := "testDeleteUserFromChannelRequestUserName1"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName1"
		workspaceName := "testDeleteUserFromChannelWorkspaceName1"
		channelName := "testDeleteUserFromChannelName1"
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
		requestUserName := "testDeleteUserFromChannelRequestUserName2"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName2"
		workspaceName := "testDeleteUserFromChannelWorkspaceName2"
		channelName := "testDeleteUserFromChannelName2"
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
		requestUserName := "testDeleteUserFromChannelRequestUserName3"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName3"
		workspaceName := "testDeleteUserFromChannelWorkspaceName3"
		channelName := "testDeleteUserFromChannelName3"
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
		requestUserName := "testDeleteUserFromChannelRequestUserName4"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName4"
		workspaceName := "testDeleteUserFromChannelWorkspaceName4"
		channelName := "testDeleteUserFromChannelName4"
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
		createChannelUserName := "testDeleteUserFromChannelCreateChannelUserName5"
		requestUserName := "testDeleteUserFromChannelRequestUserName5"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName5"
		workspaceName := "testDeleteUserFromChannelWorkspaceName5"
		channelName := "testDeleteUserFromChannelName5"
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
		requestUserName := "testDeleteUserFromChannelRequestUserName6"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName6"
		workspaceName := "testDeleteUserFromChannelWorkspaceName6"
		channelName := "testDeleteUserFromChannelName6"
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
		requestUserName := "testDeleteUserFromChannelRequestUserNam7"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName7"
		workspaceName := "testDeleteUserFromChannelWorkspaceName7"
		workspaceName2 := "testDeleteUserFromChannelWorkspaceName7.2"
		channelName := "testDeleteUserFromChannelName7"
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
		// TODO このテストケースは修正する必要あり！

		// createWorkspace funcのなかでgeneral channelを作成しているので、チャンネルのIDが分からない。そのため、DBから情報を取得する必要あり。

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
		requestUserName := "testDeleteUserFromChannelRequestUserName9"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName9"
		workspaceName := "testDeleteUserFromChannelWorkspaceName9"
		channelName := "testDeleteUserFromChannelName9"
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
		createChannelUserName := "testDeleteUserFromChannelCreateChannelUserName10"
		requestUserName := "testDeleteUserFromChannelRequestUserName10"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName10"
		workspaceName := "testDeleteUserFromChannelWorkspaceName10"
		channelName := "testDeleteUserFromChannelName10"
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
		createChannelUserName := "testDeleteUserFromChannelCreateChannelUserName11"
		requestUserName := "testDeleteUserFromChannelRequestUserName11"
		deleteUserName := "testDeleteUserFromChannelDeleteUserName11"
		workspaceName := "testDeleteUserFromChannelWorkspaceName11"
		channelName := "testDeleteUserFromChannelName11"
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

	// TODO test 12 アーカイブされている場合

}

func TestDeleteChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// 1. 正常な場合 200
	// 2. bodyに必要な情報が不足している場合(channel_id, workspace_id) 400
	// 3. requestしたuserがworkspaceに参加していない場合 404
	// 4. requestしたuserにdeleteする権限がない場合 403
	// 5. channelが存在しない場合 404
	// 6. channelがworkspaceに存在しない場合 404

	t.Run("1", func(t *testing.T) {
		testNumber := "1"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber
		channelName := "testDeleteChannelFromWorkspaceChannelName" + testNumber
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
		testNumber := "2"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber
		channelName := "testDeleteChannelFromWorkspaceChannelName" + testNumber
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
		testNumber := "3"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		createChannelUserName := "testDeleteChannelFromWorkspaceCreateChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber
		channelName := "testDeleteChannelFromWorkspaceChannelName" + testNumber
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
		testNumber := "4"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber
		channelName := "testDeleteChannelFromWorkspaceChannelName" + testNumber
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
		testNumber := "5"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber

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
		testNumber := "6"
		requestUserName := "testDeleteChannelFromWorkspaceRequestUserName" + testNumber
		inChannelUserName := "testDeleteChannelFromWorkspaceInChannelUserName" + testNumber
		workspaceName := "testDeleteChannelFromWorkspaceName" + testNumber
		workspaceName2 := "testDeleteChannelFromWorkspaceName" + testNumber + "2"
		channelName := "testDeleteChannelFromWorkspaceChannelName" + testNumber
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
	// if testing.Short() {
	// 	t.Skip("skipping test in short mode.")
	// }

	// 1. general channel以外のchannelが存在する場合 200
	// 2. userがworkspaceに存在していない場合 404

	t.Run("1 データが存在する場合", func(t *testing.T) {
		testNum := "1"
		channelCount := 10
		userName := "testGetChannelsByUser" + testNum
		workspaceName1 := "testGetChannelsByUserWorkspaceName" + testNum + ".1"
		workspaceName2 := "testGetChannelsByUserWorkspaceName" + testNum + ".2"
		isPrivate := true
		channelNames1 := make([]string, channelCount)
		channelNames2 := make([]string, channelCount)

		for i := 0; i < channelCount; i++ {
			channelNames1[i] = "testGetChannelsByUserChannelName" + strconv.Itoa(i) + ".1"
			channelNames2[i] = "testGetChannelsByUserChannelName" + strconv.Itoa(i) + ".2"
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

	t.Run("2 userがworkspaceに存在していない場合", func(t *testing.T) {
		testNum := "2"
		channelCount := 10
		userName1 := "testGetChannelsByUser" + testNum + ".1"
		userName2 := "testGetChannelsByUser" + testNum + ".2"
		workspaceName := "testGetChannelsByUserWorkspaceName" + testNum
		isPrivate := true
		channelNames := make([]string, channelCount)

		for i := 0; i < channelCount; i++ {
			channelNames[i] = "testGetChannelsByUserChannelName" + strconv.Itoa(i)
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
