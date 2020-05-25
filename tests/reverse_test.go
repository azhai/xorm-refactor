package refactor_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"gitee.com/azhai/xorm-refactor"
	"gitee.com/azhai/xorm-refactor/cmd"
	"gitee.com/azhai/xorm-refactor/config"

	"github.com/azhai/gozzo-utils/common"
	"github.com/azhai/gozzo-utils/filesystem"
	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

var (
	configFiles = []string{ // 设置多个路径，方便从子目录下运行
		"./settings.yml", "../settings.yml", "../../settings.yml",
	}
	lockFile    = "./install.lock"
	testSqlFile = "./mysql_test.sql"
)

func checkLockFile(force bool) (fp *os.File, err error) {
	if !force { // 判断 lock 文件是否已存在
		size, exists := filesystem.FileSize(lockFile)
		if exists && size > 0 {
			msg := `
==================================================
                    WARNING
初始化程序已经执行过了。
[DANGER]* 如果你想重新安装并清空这些数据表，
[DANGER]* 找到文件 %s 并删除它！
==================================================
`
			err = fmt.Errorf(fmt.Sprintf(msg, lockFile))
			return
		}
	}
	if fp, err = filesystem.CreateFile(lockFile); err == nil {
		writeLockTime(fp, "[begin] ")
	}
	return
}

func writeLockTime(fp *os.File, label string) {
	now := time.Now().Format("20060102T150405Z")
	fp.WriteString(fmt.Sprintf("%s %s\n", label, now))
}

func createTables(settings config.IReverseSettings) (err error) {
	c, ok := settings.GetConnConfig("default")
	if !ok {
		err = fmt.Errorf("the connection is not found")
		return
	}
	r, _ := config.NewReverseSource(c)
	var db *xorm.Engine
	if db, err = r.Connect(false); err != nil {
		return
	}
	var content []byte
	if content, err = ioutil.ReadFile(testSqlFile); err != nil {
		return
	}
	repl := strings.NewReplacer(
		"{{CURR_MONTH}}", time.Now().Format("200601"),
		"{{PREV_MONTH}}", time.Now().AddDate(0, -1, 0).Format("200601"),
		"{{EARLY_MONTH}}", time.Now().AddDate(0, -2, 0).Format("200601"),
	)
	sql := repl.Replace(string(content))
	_, err = db.Import(strings.NewReader(sql))
	return
}

// 一般运行 go test -v
// 如果要无视 lock 文件，运行 go test -v --args force
func Test01CreateTables(t *testing.T) {
	if lockFile != "" { // 使用 lock 文件，防止删除已有数据表
		force := common.InStringList("force", os.Args)
		fp, err := checkLockFile(force)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer writeLockTime(fp, "[finish]")
	}

	// 导入文件中的 SQL 语句， 会删除并重建相关数据表
	settings, err := cmd.Prepare(configFiles)
	assert.NoError(t, err)
	verbose := cmd.Verbose()
	if verbose {
		pp.Println(settings)
	}
	err = createTables(settings)
	assert.NoError(t, err)
	err = refactor.ExecReverseSettings(settings, verbose)
	assert.NoError(t, err)
}

func Test02Reverse(t *testing.T) {
	fileName := "./models/default/models.go"
	bs, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Greater(t, len(bs), 20) // 是否只生成一个包名，没有其他代码
	// assert.EqualValues(t, "", string(bs))
}
