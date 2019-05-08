/*
@Time : 2019-05-06 20:39
@Author : seefan
@File : errors
@Software: gossdb
*/
package consts

const (
	None = iota
	PoolEmpty
	PoolNotStart
	//连接池状态：关闭
	PoolStop
	//连接池状态：运行
	PoolStart
	//连接池状态：检查
	PoolCheck
)
