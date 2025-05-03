package main

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func BuildTree(w []string) *TreeNode {
	if len(w) == 0 {
		return nil
	}

	r := &TreeNode{}
	m := map[int]*TreeNode{}
	m[1] = r

	for i, v := range w {
		if v == "null" {
			continue
		}

		j := i + 1
		vv, ok := m[j]
		if !ok {
			fmt.Println(j)
		}
		num, _ := strconv.Atoi(v)
		vv.Val = num
		if j*2 <= len(w) {
			l := &TreeNode{}
			m[j*2] = l
			vv.Left = l
		}
		if j*2+1 <= len(w) {
			r := &TreeNode{}
			m[j*2+1] = r
			vv.Right = r
		}
	}
	return m[1]
}

type Node struct {
	Val   int
	Left  *Node
	Right *Node
	Next  *Node
}

func main() {
	s := "/sso/inner?state=xcv"
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(u.Query().Get("state"))
	//longestPalindrome("babad")
}

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */

func reverse(x int) int {
	xx := int64(x)
	if xx == 0 {
		return 0
	}
	var ixx int64
	pos := xx < 0
	var ans int64 = 0
	for ixx = 1; ixx*10 < xx; ixx *= 10 {
	}
	for i := 1; ixx > 0; i *= 10 {
		j := xx / ixx
		ans += j * int64(i)
		xx %= ixx
		ixx /= 10
	}

	if ans > math.MaxInt32 {
		return 0
	}

	if pos {
		ans *= -1
	}
	return int(ans)
}

func print() {
	i := 0
	maxi := 10000

	chans := make([]chan int, 5)
	for i := range chans {
		chans[i] = make(chan int)
	}
	g := sync.WaitGroup{}
	g.Add(5)

	for j := range 5 {
		m := j
		go func() {
			for {
				r, ok := <-chans[m]
				if i > maxi {
					if ok {
						for _, v := range chans {
							close(v)
						}
					}
					g.Done()
					return
				}

				fmt.Printf("协程%d: %d\n", m, r)
				i++
				chans[(m+1)%5] <- i
			}
		}()
	}
	chans[0] <- i
	g.Wait()
}

func test1() {
	const (
		goroutineCount = 5
		target         = 10000
	)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		cond          = sync.NewCond(&mu)
		current int32 = 1
	)

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				mu.Lock()
				for int(current)%goroutineCount != id {
					if int(current) > target {
						mu.Unlock()
						return
					}
					cond.Wait()
				}

				if int(current) > target {
					cond.Broadcast()
					mu.Unlock()
					return
				}

				fmt.Printf("协程%d: %d\n", id, current)
				atomic.AddInt32(&current, 1)
				cond.Broadcast()
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("所有协程完成打印")
}

func getNext(s string) []int {
	ne := make([]int, len(s))
	ne[1] = 0
	for i, j := 2, 0; i <= len(s)-1; i++ {
		for j > 0 && s[i] != s[j+1] {
			j = ne[j]
		}
		if s[i] == s[j+1] {
			j++
		}
		ne[i] = j
	}
	return ne
}

func kmp(s string, p string) int {
	fmt.Println(s, p)
	ne := getNext(p)
	fmt.Println(ne)

	for i, j := 1, 0; i <= len(s)-1; i++ {
		for j > 0 && p[j+1] != s[i] {
			j = ne[j]
		}
		if s[i] == p[j+1] {
			j++
		}
		if j == len(p)-1 {
			return j
		}
	}
	return -1
}

func strStr(haystack string, needle string) int {
	return kmp(" "+haystack, " "+needle)
}

func timekeep(f func()) time.Duration {
	t1 := time.Now()
	f()
	return time.Now().Sub(t1)
}
