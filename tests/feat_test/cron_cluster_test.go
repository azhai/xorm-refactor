package feat_test

import (
	"testing"
	"time"

	"gitee.com/azhai/xorm-refactor/utils"

	"gitee.com/azhai/xorm-refactor/base"
	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	"gitee.com/azhai/xorm-refactor/tests/models/cron"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

type Dict = map[string]interface{}

var taskData = Dict{
	"is_active": true, "behind": -5, // 提前 5 分钟
	"action_type": "message", "cmd_url": "meeting", // 会议通知
}

// 添加一个会议任务
func TestCron01AddTask(t *testing.T) {
	u := new(db.User)
	has, err := u.Load("username = ?", "admin")
	assert.NoError(t, err)
	if has && err == nil {
		taskData["user_uid"] = u.Uid
	}

	m := new(cron.CronTask)
	err = contrib.TruncTable(m.TableName())
	assert.NoError(t, err)
	if err == nil {
		err = m.Save(taskData)
		assert.NoError(t, err)
	}
}

// 添加三个月的会议时间记录，每周一9:00和周五16:30，但排除节假日
func TestCron02AddRecords(t *testing.T) {
	end := time.Now()
	start := end.AddDate(0, -2, 1-end.Day()) // 最近三个月（包括本月）
	pp.Printf("%s Weekday=%s\n", start, start.Weekday())

	w := start.Weekday()
	switch w {
	case 0:
		start = start.Add(time.Hour * 24)
	case 2, 3, 4:
		delta := time.Duration(5 - w)
		start = start.Add(time.Hour * 24 * delta)
	case 6:
		start = start.Add(time.Hour * 24 * 2)
	}

	var objs []base.IResetTable
	obj := contrib.CronTimerCluster{}
	obj.TaskId, obj.IsActive = 1, true
	for start.Unix() <= end.Unix() {
		ymd := start.Format("2006-01-02")
		obj.RunDate = start
		obj.ClusterMixin = base.NewClusterMonthly(start)
		if start.Weekday() == 1 {
			obj.RunClock = "09:00:00"
			start = start.Add(time.Hour * 24 * 4)
		} else {
			obj.RunClock = "16:30:00"
			start = start.Add(time.Hour * 24 * 3)
		}
		if utils.IsHoliday(ymd) {
			continue
		}
		objs = append(objs, obj)
	}
	_, err := base.ClusterInsertMulti(cron.Engine(), objs, true, true)
	assert.NoError(t, err)
}

// 找出所有周五的会议时间记录
func TestCron03FridayRecords(t *testing.T) {
	filter := func(query *xorm.Session) *xorm.Session {
		return query.Where("run_clock = ?", "16:30:00")
	}
	m := contrib.NewCronTimerCluster()
	query := base.NewClusterQuery(cron.Engine(), m.ClusterMixin)
	query.AddFilter(filter).OrderBy("run_date DESC")
	total, err := query.ClusterCount()
	pp.Println("total:", total)
	assert.NoError(t, err)
	if err == nil && total > 0 {
		var objs []*contrib.CronTimerCluster
		_, err = query.ClusterPaginate(2, 5, &objs)
		assert.NoError(t, err)
		pp.Println(objs)
	}
}
