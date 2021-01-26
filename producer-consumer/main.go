package main

import (
	"fmt"
	"sync"
	"time"
)

// Ref. https://docs.oracle.com/cd/E19253-01/819-0390/sync-31/index.html

const (
	// BufferSize is buffer size.
	BufferSize = 3
)

type buffer struct {
	wg       *sync.WaitGroup
	buf      []byte
	occupied int
	mu       *sync.Mutex
	nextIn   int
	nextOut  int
	less     *sync.Cond
	more     *sync.Cond
}

func newBuffer() *buffer {
	mu := &sync.Mutex{}
	return &buffer{
		buf:      make([]byte, BufferSize),
		occupied: 0,
		mu:       mu,
		less: &sync.Cond{
			L: mu,
		},
		more: &sync.Cond{
			L: mu,
		},
		wg: &sync.WaitGroup{},
	}
}

func main() {
	b := newBuffer()
	b.wg.Add(11)
	defer b.wg.Wait()

	go consumer(0, b)
	go consumer(1, b)
	go consumer(2, b)
	go consumer(3, b)
	go consumer(4, b)
	go consumer(5, b)
	go producer(10, b, 10)
	go producer(11, b, 11)
	go producer(12, b, 12)
	go producer(13, b, 13)
	go producer(14, b, 14)

}

func producer(num int, b *buffer, item byte) {
	defer b.wg.Done()

	b.mu.Lock()
	fmt.Printf("produce(%d) lock: b.occupied=%d, buf=%v\n", num, b.occupied, b.buf)
	for b.occupied >= BufferSize {
		// cond wait less
		fmt.Printf("produce(%d) is waiting for less signal\n", num)
		b.less.Wait()
		fmt.Printf("produce(%d) quits waiting\n", num)
	}

	if b.occupied >= BufferSize {
		panic("occupied should be equal to or larger than buffer size")
	}

	time.Sleep(time.Millisecond * time.Duration(500))
	b.buf[b.nextIn] = item
	b.nextIn++

	b.nextIn %= BufferSize
	b.occupied++

	// cond signal more
	fmt.Printf("produce(%d) sends more signal: occupied=%d, buf =%v\n\n", num, b.occupied, b.buf)
	b.more.Signal()
	b.mu.Unlock()
}

func consumer(num int, b *buffer) byte {
	defer b.wg.Done()

	b.mu.Lock()
	fmt.Printf("consumer(%d) lock: b.occupied=%d, buf=%v\n", num, b.occupied, b.buf)
	for b.occupied <= 0 {
		fmt.Printf("consumer(%d) is waiting for more signal\n", num)
		b.more.Wait()
		fmt.Printf("consumer(%d) quits waiting\n", num)
	}

	if b.occupied <= 0 {
		panic("occupied should be equal to or smaller than 0")
	}

	time.Sleep(time.Millisecond * time.Duration(500))
	item := b.buf[b.nextOut]
	b.nextOut++

	b.nextOut %= BufferSize
	b.occupied--

	fmt.Printf("consumer(%d) sends less signal: occupied=%d, buf =%v, item = %d\n\n", num, b.occupied, b.buf, item)
	b.less.Signal()
	b.mu.Unlock()

	return item
}
