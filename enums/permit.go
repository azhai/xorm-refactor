//go:generate enumer -type=Permit -text
package enums

type Permit int

// 操作列表
const (
	View   Permit = 2 << iota // 查看详情
	Draft                     // 草稿状态
	Delete                    // 删除丢弃
	Add                       // 新增添加
	Edit                      // 编辑修改
	Export                    // 导出下载

	Noop  Permit = 0 // 无操作
	Batch Permit = 1 // 批量处理
)

func IsNoop(permit int) bool {
	return Permit(permit) == Noop
}

func ContainPermit(admission, permit int, strict bool) bool {
	if strict && IsNoop(permit) {
		return false
	}
	return admission&permit == permit
}

// 分解出具体权限
func DividePermit(permit int) (codes []Permit, names []string) {
	if IsNoop(permit) {
		return
	}
	for _, c := range PermitValues() {
		if ContainPermit(permit, int(c), true) {
			codes = append(codes, c)
			names = append(names, c.String())
		}
	}
	return
}
