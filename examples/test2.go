package main

import (
	"fmt"
	"github.com/everettjf/gossdb"
)

type Book struct {
	ID    string
	Name  string
	Price int64
}

func main() {
	ip := "127.0.0.1"
	port := 8888

	pool, err := gossdb.NewPool(&gossdb.Config{
		Host: ip,
		Port: port,
	})
	if err != nil {
		fmt.Errorf("error new pool %v", err)
		return
	}
	gossdb.Encoding = true

	c, err := pool.NewClient()
	if err != nil {
		fmt.Errorf("new client err=%v", err)
		return
	}

	c.Set("myset", "hello world")
	val, err := c.Get("myset")
	fmt.Println("val=", val)

	c.Set("mysetint", 2008)
	val1, err := c.Get("mysetint")
	fmt.Println("val1=", val1.Int32())

	c.Hset("books", "id1", "book name 1")
	c.Hset("books", "id2", "book name 2")
	c.Hset("books", "id3", "book name 3")
	c.Hset("books", "id4", 8888)

	val2, err := c.Hget("books", "id1")
	fmt.Println("val2(id1)=", val2)

	val3, err := c.Hget("books", "id4")
	fmt.Println("val3(id4)=", val3)

	result, err := c.Hscan("books", "", "", 100)
	if err != nil {
		fmt.Errorf("hscan error = %v", err)
	}
	fmt.Printf("result=%v\n", result)

	c.Hset("booklist", "id1", &Book{"id1", "name1", 1})
	c.Zset("booklist_order_by_price", "id1", 1)
	c.Hset("booklist", "id2", &Book{"id2", "name2", 2})
	c.Zset("booklist_order_by_price", "id2", 2)
	c.Hset("booklist", "id3", &Book{"id3", "name3", 1})
	c.Zset("booklist_order_by_price", "id3", 1)
	c.Hset("booklist", "id4", &Book{"id4", "name4", 3})
	c.Zset("booklist_order_by_price", "id4", 3)
	c.Hset("booklist", "id5", &Book{"id5", "name5", 5})
	c.Zset("booklist_order_by_price", "id5", 5)
	c.Hset("booklist", "id6", &Book{"id6", "name6", 14})
	c.Zset("booklist_order_by_price", "id6", 14)
	c.Hset("booklist", "id7", &Book{"id7", "name7", 13})
	c.Zset("booklist_order_by_price", "id7", 13)
	c.Hset("booklist", "id8", &Book{"id8", "name8", 12})
	c.Zset("booklist_order_by_price", "id8", 12)
	c.Hset("booklist", "id9", &Book{"id9", "name9", 11})
	c.Zset("booklist_order_by_price", "id9", 11)
	c.Hset("booklist", "id10", &Book{"id10", "name10", 10})
	c.Zset("booklist_order_by_price", "id10", 10)

	result1, err := c.Hscan("booklist", "", "", 100)
	fmt.Printf("result1=%v\n", result1)

	for i, b := range result1 {
		fmt.Printf("%v - %v\n", i, b)
	}

	result2, err := c.Zscan("booklist_order_by_price", "", 10, "", 1000)
	fmt.Printf("result2=%v\n", result2)

	for i, b := range result2 {
		fmt.Printf("%v - %v\n", i, b)
	}

	keys := []string{}
	//keys := make([]string, 0, len(result2))
	for k := range result2 {
		keys = append(keys, k)
	}

	result3, err := c.MultiHget("booklist", keys...)
	fmt.Printf("result3=%v\n", result3)
	for i, b := range result3 {
		fmt.Printf("%v - %v\n", i, b)

		var book Book
		b.As(&book)
		fmt.Printf("---book=%v\n", book)
	}

	c.Zset("zsettest", "zsettest_count", 0)
	c.Zincr("zsettest", "zsettest_count", 1)
	c.Zincr("zsettest", "zsettest_count", 1)
	c.Zincr("zsettest", "zsettest_count", 1)
	count, err := c.Zget("zsettest", "zsettest_count")
	if err != nil {
		fmt.Println("zsettest_count error = ", err.Error())
	} else {
		fmt.Printf("zsettest_count = %v\n", count)
	}

}
