[![](https://goreportcard.com/badge/gitee.com/azhai/xorm-refactor)](https://goreportcard.com/report/gitee.com/azhai/xorm-refactor)

# Xorm-Refactor

一个灵活高效的数据库反转工具。将数据库中的表生成对应的Model和

相应的查询方法，特点是自动套用包名和将已知的 Mixin 嵌入Model 中。

已实现以下功能并有主要功能的测试代码：

* 从数据表生成对应的 Model ，每个连接一个子目录，文件为 models.go
* 同时根据模板生成对应的初始化连接和自定义查询文件，分别为 conn.go 和 queries.go
* 根据已知和已扫描到 Mixin ，嵌入有用相同字段的 Model 中，即自动进行 xorm:“extends” 标注。
* 提供分页和联表查询的方法，特别是可以方便地进行多次 LEFT JOIN 操作。
* 提供高强度密码哈希方法，以及快速生成 流水号/序列号/令牌 的方法。
* 提供嵌套集合树的 Mixin ，方便对多级树状数据进行查询和更新两端数字。
* 提供权限分配和认证的辅助函数和范例，满足多数情况下的鉴权需求。
* 支持分库分表查询

## 常见用法

以MySQL中的一个菜单表为例，建表SQL语句:

```sql
CREATE TABLE `t_menu`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `lft` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '左边界',
  `rgt` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '右边界',
  `depth` tinyint(3) UNSIGNED NOT NULL DEFAULT 1 COMMENT '高度',
  `path` varchar(100) NOT NULL DEFAULT '' COMMENT '路径',
  `title` varchar(50) NOT NULL DEFAULT '' COMMENT '名称',
  `icon` varchar(30) NULL DEFAULT NULL COMMENT '图标',
  `remark` text NULL COMMENT '说明备注',
  `created_at` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp(0) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` timestamp(0) NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_menu_rgt`(`rgt`) USING BTREE,
  INDEX `idx_menu_depth`(`depth`) USING BTREE,
  INDEX `idx_menu_path`(`path`) USING BTREE,
  INDEX `idx_menu_deleted_at`(`deleted_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COMMENT = '菜单' ROW_FORMAT = DYNAMIC;
```

默认在 models/default/ 下生成 3 个代码文件 init.go 、 models.go 和 queries.go

```go
package db
// Filename is models.go

import (
	"gitee.com/azhai/xorm-refactor/base"
)

type Menu struct {
	Id                int `json:"id" xorm:"notnull pk autoincr INT(10)"`
	*base.NestedMixin `json:",inline" xorm:"extends"`
	Path              string `json:"path" xorm:"notnull default '' comment('路径') index VARCHAR(100)"`
	Title             string `json:"title" xorm:"notnull default '' comment('名称') VARCHAR(50)"`
	Icon              string `json:"icon" xorm:"comment('图标') VARCHAR(30)"`
	Remark            string `json:"remark" xorm:"comment('说明备注') TEXT"`
	*base.TimeMixin   `json:",inline" xorm:"extends"`
}
```

```go
package db
// Filename is queries.go

import (
	"time"

	"xorm.io/xorm"
)

func (m *Menu) Load(where interface{}, args ...interface{}) (bool, error) {
	return Table().Where(where, args...).Get(m)
}

func (m *Menu) Save(changes map[string]interface{}) error {
	return ExecTx(func(tx *xorm.Session) (int64, error) {
		if changes == nil || m.Id == 0 {
			changes["created_at"] = time.Now()
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.Id).Update(changes)
		}
	})
}
```

init.go 中含有 Initialize() 方法，通过下面的方法，在程序入口 main() 或 init() 中

初始化数据库连接。

```go
package main

import (
	"gitee.com/azhai/xorm-refactor/setting"
	"my-project/models/db"
	"xorm.io/xorm/log"
)

var verbose bool // 详细输出

func init() {
	cfg, err := setting.ReadSettings("settings.yml")
	if err != nil {
		panic(err)
	}
    logger := &log.DiscardLogger{}
	db.Initialize(cfg, logger, verbose)
}
```

init.go 中含有 QueryAll() 方法用于查询表中多行数据，而 queries.go 中的  Load() 方法

只查询指定的一行数据。配合 NestedMixin 我们可以查询子菜单：

```go
package main

import (
	"my-project/models/db"
)

// 根据 id 找出菜单及其子菜单（最多三层）
func GetMenus(id int) ([]*db.Menu, error) {
	m := new(db.Menu)
	has, err := m.Load("id = ?", id) // 找出指定 id 的菜单
	if err != nil || has == false {
		return nil, err
	}
	filter := m.ChildrenFilter(3, true) // 最多三层子菜单，如不限制传递参数 0
	pageno, pagesize := 0, -1     // 符合条件所有数据，即不分页
	var menus []*db.Menu
	err = db.QueryAll(filter, pageno, pagesize).Find(&menus)
	return menus, err
}
```

## ~~测试~~

~~安装好Go编译器，下载本项目源码解压。~~

~~进入 tests 目录，修改 settings.yml 中的 MySQL 和 Redis 地址、端口、用户名和密码等配置。~~

~~运行 go test -v 将会在数据库中创建表（具体内容请查看 mysql_test.sql 文件），并生成 models~~

~~接着进入 tests 目录下的 crud_test 子目录， 运行 go test -v 执行各种查询、写入、鉴权测试，~~

~~具体执行了什么，请查看屏幕输出和阅读子目录下的 *_test.go 测试文件代码。~~


## 安装

```
go get gitee.com/azhai/xorm-refactor
```

## 编译使用

```
make all
./reverse -c tests/settings.yml
```

## 配置文件

一个典型的配置文件看起来如下：

```yml
debug: true

reverse_target:
   output_dir: "./models"  # 代码生成目录
   init_name_space: "my-project/models" #完整引用model的URL
   template_path: ""       # 生成的模板的路径，优先级比 language 中的默认模板高
   query_template_path: "" # 自定义查询方法模板
   init_template_path: "./data/query_init.tmpl"  # 自定义初始化方法模板
   table_mapper: snake     # 表名到代码类或结构体的映射关系
   column_mapper: snake    # 字段名到代码或结构体成员的映射关系
   multiple_files: false   # 每个model一个go文件
   apply_mixins: true      # 使用已知的Mixin替换部分字段
   mixin_dir_path: ""      # 额外的mixin目录
   mixin_name_space: ""    # 额外的mixin包名

_mysql: &mysql           #共用数据库配置
   driver_name: "mysql"
   params:
      host: "127.0.0.1"
      port: 3306
      username: "root"
      password: ""
      database: "test"
      options: { charset: "utf8" }

connections:
   cache:
      driver_name: "redis"
      params:
         host: "127.0.0.1"
         port: 6379
         password: ""
         database: "0"
   default:
      read_only: false
      table_prefix: "t_" # 表前缀
      include_tables:    # 包含的表，以下可以用
         - "a*"
         - "b*"
      exclude_tables:    # 排除的表，以下可以用
         - "c"
      <<: *mysql         #引用mysql配置
```