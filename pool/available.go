//  @time : 2019-05-24 15:11
//  @author : seefan
//  @file : available
//  @software: gossdb

package pool

//Avaliable avaliable interface
type Avaliable interface {
	Pop() int
	Put(int) int
	IsEmpty() bool
}
