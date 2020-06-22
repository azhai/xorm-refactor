package utils

import (
	"fmt"
	"strings"
)

// 用:号连接两个部分，如果后一部分也存在的话
func ConcatWith(master, slave string) string {
	if slave != "" {
		master += ":" + slave
	}
	return master
}

// 如果本身不为空，在左右两边添加字符
func WrapWith(s, left, right string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", left, s, right)
}

// 将字符串数组转为一般数组
func StrToList(data []string) []interface{} {
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = v
	}
	return result
}

func SprintfString(tpl string, data []string) string {
	return fmt.Sprintf(tpl, StrToList(data)...)
}

// 将多个连续空白缩减为一个空格
func ReduceSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// 删除所有空白，包括中间的
func RemoveSpaces(s string) string {
	subs := map[string]string{
		" ": "", "\n": "", "\r": "", "\t": "", "\v": "", "\f": "",
	}
	return ReplaceWith(s, subs)
}

// 一一对应进行替换，次序不定（因为map的关系）
func ReplaceWith(s string, subs map[string]string) string {
	if s == "" {
		return ""
	}
	var marks []string
	for key, value := range subs {
		marks = append(marks, key, value)
	}
	replacer := strings.NewReplacer(marks...)
	return replacer.Replace(s)
}
