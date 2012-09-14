// example program

package main

import (
	"fmt"
	"github.com/littlewhitetwo/radix/redis"
	"time"
)

var sig chan int

func test(no int, c *redis.Client) {

	var err error

	fmt.Printf("[%d]to get \n", no)
	key := fmt.Sprintf("key:%d", no)

	r := c.Get(key)

	fmt.Printf("[%d]after get\n", no)

	result := 0
	if r.Type == redis.ReplyNil {

		fmt.Printf("[%d]notfound \n", no)
		result = 0

	} else {

		result, err = r.Int()
		if err != nil {
			fmt.Printf("[%d]err:%s\n", no, err)
		}
	}
	fmt.Printf("[%d]mykeys values:%d\n", no, result)
	sig <- no
}

func main() {
	var c *redis.Client

	conf := redis.DefaultConfig()
	//	conf.Address = "192.168.1.211:6379"

	conf.Database = 8
	conf.PoolCapacity = 1
	conf.Timeout = time.Duration(210) * time.Millisecond

	fmt.Printf("conf:%+v \n", conf)

	c = redis.NewClient(conf)

	c.Set("mykey", "111")

	defer c.Close()

	size := 4

	sig = make(chan int, size)

	for i := 0; i < size; i++ {
		go test(i, c)
	}

	fmt.Printf("waiting\n")
	for i := 0; i < size; i++ {
		fmt.Printf("recv %d\n", <-sig)
	}
	fmt.Printf("exit\n")

	return

}
