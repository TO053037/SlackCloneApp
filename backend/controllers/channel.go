package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend/models"
	"backend/utils"
)

func CreateChannel(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	userId, err := Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// urlからパラメータを取得
	workspaceId, err := strconv.Atoi(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyの情報を取得
	var ch models.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channel nameがbodyに含まれているか確認
	if ch.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found channel name"})
		return
	}

	// workspaceIdに対応するworkspaceが存在するか確認
	if !models.IsExistWorkspaceById(workspaceId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found workspace"})
		return
	}

	// 同じ名前のchannelが対応するworkspaceに存在しないか確認
	b, err := ch.IsExistSameNameChannelInWorkspace(workspaceId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if b {
		c.JSON(http.StatusBadRequest, gin.H{"message": "already exist same name channel in workspace"})
		return
	}

	// userが対象のworkspaceに参加しているか確認
	wau := models.NewWorkspaceAndUsers(workspaceId, userId, 0)
	if !wau.IsExistWorkspaceAndUser() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found user in workspace"})
		return
	}

	// channels tableに情報を保存
	if err := ch.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channels_and_workspaces tableに保存する情報を作成し保存
	caw := models.NewChannelsAndWorkspaces(ch.ID, workspaceId)
	if err := caw.Create(); err != nil {
		// TODO DeleteChannelsAndWorkspaces funcを実行する
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channels_and_users tableに保存する情報を作成し保存
	cau := models.NewChannelsAndUses(ch.ID, userId, true)
	if err := cau.Create(); err != nil {
		// TODO DeleteChannel funcを実行する
		// TODO DeleteChannelsAndWorkspace funcを実行

		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ch)
}

func AddUserInChannel(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	userId, err := Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// urlからパラメータを取得
	workspaceId, err := strconv.Atoi(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyの情報を取得
	var cau models.ChannelsAndUsers
	if err := c.ShouldBindJSON(&cau); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyにuserIdとchannelIdが含まれているか確認
	if cau.ChannelId == 0 || cau.UserId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found channel_id or user_id"})
		return
	}

	// とりあえず管理権限はなし
	cau.IsAdmin = false

	// リクエストしたuserがworkspaceに参加してるかを確認
	rwau := models.NewWorkspaceAndUsers(workspaceId, userId, 0)
	if !rwau.IsExistWorkspaceAndUser() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not exist request user in workspace"})
		return
	}

	// 追加されるuserがworkspaceに参加しているかを確認
	awau := models.NewWorkspaceAndUsers(workspaceId, cau.UserId, 0)
	if !awau.IsExistWorkspaceAndUser() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not exist added user in workspace"})
		return
	}

	// 対象のchannelがworkspace内に存在するかを確認
	if !models.IsExistCAWByChannelIdAndWorkspaceId(cau.ChannelId, workspaceId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not exist channel in workspace"})
		return
	}

	// 追加されるユーザーが既に対象のchannelに存在していないかを確認
	if models.IsExistCAUByChannelIdAndUserId(cau.ChannelId, cau.UserId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "already exist user in channel"})
		return
	}

	// リクエストしたuserにchannelの管理権限があるかを確認(結果的にリクエストしたuserがchannelに所属しているかも確認される)
	if !utils.HasPermissionAddingUserInChannel(cau.ChannelId, userId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no permission adding user in channel"})
		return
	}

	// channels_and_users tableに登録
	if err := cau.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cau)
}

func DeleteUserFromChannel(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	userId, err := Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// urlからパラメータを取得
	workspaceId, err := strconv.Atoi(c.Param("workspace_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyを取得
	var cau models.ChannelsAndUsers
	if err := c.ShouldBindJSON(&cau); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyに必要な情報があるかを確認
	if cau.UserId == 0 || cau.ChannelId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found user_id or channel_id"})
		return
	}

	// requestしたuserがworkspaceにいることを確認
	rwau := models.NewWorkspaceAndUsers(workspaceId, userId, 0)
	if !rwau.IsExistWorkspaceAndUser() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found request user in workspace"})
		return
	}

	// deleteされるuserがworkspaceにいることを確認
	wau := models.NewWorkspaceAndUsers(workspaceId, cau.UserId, 0)
	if !wau.IsExistWorkspaceAndUser() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found user in workspace"})
		return
	}

	// channelの情報を取得
	ch, err := models.GetChannelById(cau.ChannelId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channelがworkspaceに存在することを確認
	if !models.IsExistCAWByChannelIdAndWorkspaceId(cau.ChannelId, workspaceId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found channel in workspace"})
		return
	}

	// channelのnameがgeneralでないことを確認
	if ch.Name == "general" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "don't delete general channel"})
		return
	}

	// channelがアーカイブされていないことを確認
	if ch.IsArchive {
		c.JSON(http.StatusBadRequest, gin.H{"message": "don't delete archived channel"})
		return
	}

	// deleteされるuserがchannelに存在することを確認
	if !models.IsExistCAUByChannelIdAndUserId(cau.ChannelId, cau.UserId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found user in channel"})
		return
	}

	// deleteする権限があるかを確認
	if !utils.HasPermissionDeletingUserInChannel(userId, workspaceId, ch) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not permission deleting user in channel"})
		return
	}

	// channels_and_users tableから削除
	if err := cau.Delete(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cau)
}

func DeleteChannel(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	userId, err := Authenticate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyの情報を取得
	var caw models.ChannelsAndWorkspaces
	if err := c.ShouldBindJSON(&caw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// bodyの情報に不足がないか確認
	if caw.ChannelId == 0 || caw.WorkspaceId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found channel_id or workspace_id"})
		return
	}

	// requestしたuserがworkspaceに参加しているかを確認
	wau, err := models.GetWorkspaceAndUserByWorkspaceIdAndUserId(caw.WorkspaceId, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// deleteする権限があるかを確認
	if !utils.HasPermissionDeletingChannel(wau) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no permission deleting channel"})
		return
	}

	// channelが存在するかどうかを確認
	ch, err := models.GetChannelById(caw.ChannelId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channelがworkspaceにあるかどうかを確認
	if !models.IsExistCAWByChannelIdAndWorkspaceId(caw.ChannelId, caw.WorkspaceId) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "not found channel in workspace"})
		return
	}

	// channels tableからデータを削除
	if err := ch.Delete(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channels_and_workspaces tableからデータを削除
	if err := caw.Delete(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// channel_and_users tableからデータを削除
	if err := models.DeleteCAUByChannelId(caw.ChannelId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// TODO roll back func

	c.JSON(http.StatusOK, caw)
}
