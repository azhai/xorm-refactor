package base

import (
	"fmt"
	"strings"

	utils "github.com/azhai/gozzo-utils/common"
	"github.com/azhai/gozzo-utils/redisw"
	"github.com/gomodule/redigo/redis"
)

const (
	SESS_FOR_EVER   = -1 // 无限
	SESS_NOT_EXISTS = -2 // 不存在
	SESS_ERR_REDIS  = -3 // redis错误

	SESS_PREFIX     = "sess" // 会话缓存前缀
	SESS_LIST_SEP   = ";"    // 角色名之间的分隔符
	SESS_TOKEN_KEY  = "_token_"
	SESS_ONLINE_KEY = "onlines" // 在线用户

	MAX_TIMEOUT = 86400 * 30 // 接近无限时间
)

func SessListJoin(data []string) string {
	return strings.Join(data, SESS_LIST_SEP)
}

func SessListSplit(data string) []string {
	return strings.Split(data, SESS_LIST_SEP)
}

type SessionRegistry struct {
	sessions map[string]*Session
	Onlines  *redisw.RedisHash
	*redisw.RedisWrapper
}

func NewRegistry(w *redisw.RedisWrapper) *SessionRegistry {
	return &SessionRegistry{
		sessions:     make(map[string]*Session),
		Onlines:      redisw.NewRedisHash(w, SESS_ONLINE_KEY, MAX_TIMEOUT),
		RedisWrapper: w,
	}
}

func (sr SessionRegistry) GetKey(token string) string {
	return fmt.Sprintf("%s:%s", SESS_PREFIX, token)
}

func (sr *SessionRegistry) GetSession(token string, timeout int) *Session {
	key := sr.GetKey(token)
	if sess, ok := sr.sessions[key]; ok && sess != nil {
		return sess
	}
	sess := NewSession(sr, key, timeout)
	if _, err := sess.SetVal(SESS_TOKEN_KEY, token); err == nil {
		sr.sessions[key] = sess
	}
	return sess
}

func (sr *SessionRegistry) DelSession(token string) bool {
	key := sr.GetKey(token)
	if sess, ok := sr.sessions[key]; ok {
		succ, err := sess.DeleteAll()
		if succ && err == nil {
			delete(sr.sessions, key)
			return true
		}
	}
	return false
}

type Session struct {
	reg *SessionRegistry
	*redisw.RedisHash
}

func NewSession(reg *SessionRegistry, key string, timeout int) *Session {
	hash := redisw.NewRedisHash(reg.RedisWrapper, key, timeout)
	return &Session{reg: reg, RedisHash: hash}
}

func (sess *Session) GetKey() string {
	token, err := sess.GetString(SESS_TOKEN_KEY)
	if err == nil && token != "" {
		return sess.reg.GetKey(token)
	}
	return ""
}

func (sess *Session) WrapExec(cmd string, args ...interface{}) (interface{}, error) {
	// pp.Println(cmd, args)
	return sess.reg.Exec(cmd, args...)
}

// 添加临时消息
func (sess *Session) AddFlash(messages ...string) (int, error) {
	key := fmt.Sprintf("flash:%s", sess.GetKey())
	args := append([]interface{}{key}, utils.StrToList(messages)...)
	return redis.Int(sess.WrapExec("RPUSH", args...))
}

// 数量n为最大取出多少条消息，-1表示所有消息
func (sess *Session) GetFlashes(n int) ([]string, error) {
	end := -1
	if n > 0 {
		end = n - 1
	}
	key := fmt.Sprintf("flash:%s", sess.GetKey())
	return redis.Strings(sess.WrapExec("LRANGE", key, 0, end))
}

// 绑定用户角色，返回旧的sid
func (sess *Session) BindRoles(uid string, roles []string, kick bool) (string, error) {
	newSid := sess.GetKey()
	oldSid, _ := sess.reg.Onlines.GetString(uid) // 用于踢掉重复登录
	if oldSid == newSid {                        // 同一个token
		oldSid = ""
	}
	_, err := sess.reg.Onlines.SetVal(uid, newSid)
	_, err = sess.SetVal("uid", uid)
	_, err = sess.SetVal("roles", SessListJoin(roles))
	if kick && oldSid != "" { // 清空旧的session
		flashKey := fmt.Sprintf("flash:%s", oldSid)
		sess.reg.Delete(oldSid, flashKey)
	}
	return oldSid, err
}
