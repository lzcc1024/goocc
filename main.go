package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/lzcc1024/goucc/opencc"
)

func main() {

	d, err := opencc.NewOpenCC("s2t")
	if err != nil {
		fmt.Println(err)
	}

	str := `Go 是一个开源的编程语言，它能让构造简单、可靠且高效的软件变得容易。Go是从2007年末由Robert Griesemer, Rob Pike, Ken Thompson主持开发，后来还加入了Ian Lance Taylor, Russ Cox等人，并最终于2009年11月开源，在2012年早些时候发布了Go 1稳定版本。现在Go的开发已经是完全开放的，并且拥有一个活跃的社区。`

	if ss, err := d.Convert(str); err == nil {
		fmt.Println(ss, err)
	}

	var wg sync.WaitGroup

	s := time.Now().Unix()

	for j := 0; j < 8; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// defer wg.Add(-1)
			for i := 0; i < 1200; i++ {

				if _, err := d.Convert(str); err != nil {
					fmt.Println(err)
				}

			}
		}()
	}

	wg.Wait()

	e := time.Now().Unix()
	fmt.Println(s, e, e-s)

}
