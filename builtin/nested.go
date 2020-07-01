package builtin

import (
	"xorm.io/xorm"
)

var NESTED_EDGE_OFFSET = 2

// 嵌套集合树
type NestedMixin struct {
	Lft   int `json:"lft" xorm:"notnull default 0 comment('左边界') INT(10)"`           // 左边界
	Rgt   int `json:"rgt" xorm:"notnull default 0 comment('右边界') index INT(10)"`     // 右边界
	Depth int `json:"depth" xorm:"notnull default 1 comment('高度') index TINYINT(3)"` // 高度
}

type NestedRow struct {
	Id                  int `json:"id" xorm:"notnull pk autoincr INT(10)"`
	ParentIdNihility996 int `json:"-" xorm:"-"` // 阻止对象被混入其他 Model 的特殊字段
	*NestedMixin        `json:",inline" xorm:"extends"`
}

func (n NestedMixin) GetDepth() int {
	return n.Depth
}

// 是否叶子节点
func (n NestedMixin) IsLeaf() bool {
	return n.Rgt-n.Lft == 1
}

// 有多少个子孙节点
func (n NestedMixin) CountChildren() int {
	return (n.Rgt - n.Lft - 1) / 2
}

// 找出所有直系祖先节点
func (n NestedMixin) AncestorsFilter(backward bool) FilterFunc {
	return func(query *xorm.Session) *xorm.Session {
		query = query.Where("rgt > ? AND lft < ?", n.Rgt, n.Lft)
		if backward { // 从子孙往祖先方向排序，即时间倒序
			return query.OrderBy("rgt ASC")
		} else {
			return query.OrderBy("rgt DESC")
		}
	}
}

// 找出所有子孙节点
func (n NestedMixin) ChildrenFilter(rank int, depthFirst bool) FilterFunc {
	return func(query *xorm.Session) *xorm.Session {
		if n.Rgt > 0 && n.Lft > 0 { // 当前不是第0层，即具体某分支以下的节点
			query = query.Where("rgt < ? AND lft > ?", n.Rgt, n.Lft)
		}
		if rank > 0 { // 限制层级
			query = query.Where("depth <= ?", n.Depth+rank)
		}
		if rank != 1 && depthFirst { // 多层先按高度排序
			query = query.OrderBy("depth ASC")
		}
		return query.OrderBy("rgt ASC")
	}
}

// 添加到父节点最末
func (n *NestedMixin) AddToParent(parent *NestedMixin, query *xorm.Session, table string) (err error) {
	query = query.Table(table).OrderBy("rgt DESC")
	if parent == nil {
		n.Depth = 1
	} else {
		n.Depth = parent.Depth + 1
		query = query.Where("rgt < ? AND lft > ?", parent.Rgt, parent.Lft)
	}
	query = query.Where("depth = ?", n.Depth)
	sibling := new(NestedMixin)
	if _, err = query.Get(sibling); err != nil {
		return
	}

	// 重建受影响的左右边界
	if sibling.Depth > 0 {
		n.Lft = sibling.Rgt + 1
	} else if parent != nil {
		n.Lft = parent.Lft + 1
	} else {
		n.Lft = 1
	}
	n.Rgt = n.Lft + 1
	if n.Depth > 1 {
		err = MoveEdge(query, table, n.Lft, NESTED_EDGE_OFFSET)
		parent.Rgt += NESTED_EDGE_OFFSET // 上面的数据更新使 parent.Rgt 变成脏数据
	}
	return
}

// 左右边界整体移动
func MoveEdge(query *xorm.Session, table string, base, offset int) error {
	emptyChanges := map[string]interface{}{}
	// 更新右边界
	query = query.Table(table).Where("rgt >= ?", base)              // 下面的更新lft也要用rgt作为索引
	affected, err := query.Incr("rgt", offset).Update(emptyChanges) // 等同于先 Exec() 再 RowsAffected()
	if affected == 0 || err != nil {
		return err
	}

	// 更新左边界，范围一定在上面更新右边界的所有行之内
	// 要么和上面一起为空，要么比上面少>=n行，n为直系祖先数量
	if affected > 1 {
		query = query.Table(table).Where("rgt >= ? AND lft >= ?", base, base)
		_, err = query.Incr("lft", offset).Update(emptyChanges)
	}
	return err
}

// 按照深度重建，即用之前比它深度小且最近的一个作为父节点
func RebuildNestedByDepth(query *xorm.Session, table string) error {
	var rows []*NestedRow
	err := query.Table(table).OrderBy("id").Find(&rows)
	if err != nil || len(rows) == 0 {
		return err
	}
	left, right := 1, 2
	for i, obj := range rows {
		if i == 0 {
			obj.ParentIdNihility996 = -1 // 无父节点
			rows[i].Lft, rows[i].Rgt = left, right
			continue
		}
		last := rows[i-1]
		if obj.Depth > last.Depth { // 作为第一个子节点
			rows[i].ParentIdNihility996 = i - 1
			rows[i].Lft, rows[i].Rgt = last.Rgt, last.Rgt+1
			rows[i].Depth, rows[i-1].Rgt = last.Depth+1, last.Rgt+2
			continue
		}
		for obj.Depth < last.Depth {
			if last.ParentIdNihility996 < 0 {
				break
			}
			last = rows[last.ParentIdNihility996]
		}
		// 作为兄弟节点
		rows[i].ParentIdNihility996 = last.ParentIdNihility996
		if last.ParentIdNihility996 >= 0 {
			parentNode := rows[last.ParentIdNihility996]
			rows[last.ParentIdNihility996].Rgt = parentNode.Rgt + 2
		}
		rows[i].Lft, rows[i].Rgt = last.Rgt+1, last.Rgt+2
	}
	for _, obj := range rows {
		obj.ParentIdNihility996 = 0 // 忽视此字段
		_, err = query.Table(table).ID(obj.Id).Update(obj)
	}
	return nil
}
