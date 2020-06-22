package feat_test

import (
	"fmt"
	"testing"

	"gitee.com/azhai/xorm-refactor/builtin/base"
	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	"gitee.com/azhai/xorm-refactor/tests/models/cache"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"
	"gitee.com/azhai/xorm-refactor/utils"

	"github.com/stretchr/testify/assert"
)

var (
	username = "admin"
	password = "admin"
	token    = utils.CreateToken([]byte("T"), 2)
)

// 写入Session
func writeSession(sess *base.Session, user *db.User) error {
	roles, err := contrib.GetUserRoles(user)
	if err != nil {
		return err
	}
	_, err = sess.BindRoles(user.Uid, roles, true)
	if err != nil {
		return err
	}
	var ok bool
	data := contrib.GetUserInfo(user)
	ok, err = sess.SaveMap(data, false)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("SaveMap() is failure")
	}
	return nil
}

// 用户登录，成功后信息写入 Session
func TestSess01Login(t *testing.T) {
	user := new(db.User)
	has, err := user.Load("username = ?", username)
	assert.NoError(t, err)
	if has && err == nil {
		ok := utils.VerifyPassword(password, user.Password)
		assert.True(t, ok)

		sess := cache.Session(token)
		err = writeSession(sess, user)
		assert.NoError(t, err)
	}
}

// 测试临时消息
func TestSess02Flash(t *testing.T) {
	sess := cache.Session(token)
	num, err := sess.AddFlash("1 One", "2 Two Double")
	assert.NoError(t, err)
	assert.Equal(t, num, 2)
	var msgs []string
	msgs, err = sess.GetFlashes(-1)
	assert.NoError(t, err)
	assert.Len(t, msgs, 2)
	msgs, err = sess.GetFlashes(1)
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
}

// 测试自动刷新时间
func TestSess02Timeout(t *testing.T) {
	sess := cache.Session(token)
	timeout := sess.GetTimeout(false)
	assert.Greater(t, timeout, cache.SESS_CREATE_TIMEOUT-3)

	// 高于 SESS_RESCUE_TIMEOUT 时间不刷新
	sess.Expire(cache.SESS_RESCUE_TIMEOUT + 3)
	timeout = cache.Session(token).GetTimeout(false)
	assert.Less(t, timeout, cache.SESS_CREATE_TIMEOUT-3)
	assert.Greater(t, timeout, cache.SESS_RESCUE_TIMEOUT)

	// 低于 SESS_RESCUE_TIMEOUT 时间自动刷新
	sess.Expire(cache.SESS_RESCUE_TIMEOUT - 1)
	timeout = cache.Session(token).GetTimeout(false)
	assert.Greater(t, timeout, cache.SESS_CREATE_TIMEOUT-3)
}
