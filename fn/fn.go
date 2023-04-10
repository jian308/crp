/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-08 22:30:55
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-09 23:12:51
 */
package fn

import (
	"net"
	"strings"
)

func DeleteSlice3(a []int, elem int) []int {
	j := 0
	for _, v := range a {
		if v != elem {
			a[j] = v
			j++
		}
	}
	return a[:j]
}

func GetIp(conn net.Conn) string {
	RemoteAddr := conn.RemoteAddr().String()
	return RemoteAddr[0:strings.LastIndex(RemoteAddr, ":")]
}
