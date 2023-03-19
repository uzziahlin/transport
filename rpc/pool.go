package rpc

import (
	"context"
	"io"
	"sync"
	"time"
)

type Closer interface {
	io.Closer
}

type Pool[T Closer] interface {
	Put(ctx context.Context, t T) error
	Get(ctx context.Context) (T, error)
}

type Conn[T Closer] struct {
	t              T
	lastActiveTime time.Time // 记录当前连接最后一次放回连接池的时间，用来判断是否空闲连接
}

type ConnPool[T Closer] struct {
	idleConns   chan *Conn[T]            // 空闲连接池
	activeCnt   int                      // 目前活跃连接数
	maxActive   int                      // 最大活跃连接数
	maxIdleTime time.Duration            // 最大空闲时间
	waitQ       map[uint64]chan *Conn[T] // 等待队列
	factory     func() T                 // 工厂函数，定义了如何创建T
	mu          sync.Mutex
	seq         uint64
}

// Put 将连接放回连接池
// 先判断等待队列有没有g在等待，如果有，则直接交给对方
// 如果没有g在等待，则将连接放到空闲连接池
// 如果无法放回空闲连接池，则关闭连接
func (c *ConnPool[T]) Put(ctx context.Context, t T) error {
	c.mu.Lock()

	// 先判断等待队列是否为空，不为空直接把连接交给对方
	if len(c.waitQ) > 0 {
		for k, v := range c.waitQ {
			delete(c.waitQ, k)
			// 此处应该解锁，避免阻塞死锁
			c.mu.Unlock()

			// 避免阻塞
			select {
			case v <- &Conn[T]{t: t}:
			default:
			}

			return nil
		}
	}

	// 走到这说明等待队列为空， 则尝试放入空闲连接池
	select {
	case c.idleConns <- &Conn[T]{
		t:              t,
		lastActiveTime: time.Now(),
	}:
	default:
		_ = t.Close()
		c.activeCnt--
	}

	c.mu.Unlock()

	return nil
}

// Get 从连接池获取连接
// 先尝试从空闲连接池获取连接，如果能获取到，则返回
// 如果无法从空闲连接池获取连接，则判断目前活跃连接数是否达到最大活跃连接数，如果没有，则创建一个新的连接
// 否则，则将当前g放入等待队列，等待其他g唤醒
func (c *ConnPool[T]) Get(ctx context.Context) (T, error) {

	var res T

	if ctx.Err() != nil {
		return res, ctx.Err()
	}

	// 先尝试从空闲连接池拿
Loop:
	for {
		select {
		case conn := <-c.idleConns:
			// 如果拿到了，判断当前连接是否超过最大空闲时间，如果超过，则关闭
			if conn.lastActiveTime.Add(c.maxIdleTime).Before(time.Now()) {
				// 走到这里说明，连接超过最大空闲时间，则关闭
				_ = conn.t.Close()
				c.mu.Lock()
				c.activeCnt--
				c.mu.Unlock()
				continue
			}
			// 如果没超过最大空闲时间，则返回
			return conn.t, nil
		default:
			// 走到这里，说明空闲连接池获取不到连接
			break Loop
		}
	}

	// 走到这里，说明空闲连接池里获取不到连接，则判断是否达到最大连接数
	c.mu.Lock()
	if c.activeCnt < c.maxActive {
		res = c.factory()
		c.activeCnt++
		c.mu.Unlock()
		return res, nil
	}

	// 走到这里，说明超过最大活跃连接上线，则阻塞等待唤醒
	q := make(chan *Conn[T], 1)
	seq := c.seq
	c.waitQ[seq] = q
	c.seq++

	// 此处应该解锁，避免阻塞死锁
	c.mu.Unlock()

	select {
	case conn := <-q:
		// 如果收到信号，假设不会过期
		return conn.t, nil
	case <-ctx.Done():
		// 走到这里说明超时了，应该删除等待队列中
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.waitQ, seq)

		// 避免漏信号，如果收到应该转发出去
		select {
		case conn := <-q:
			_ = c.Put(context.TODO(), conn.t)
		default:
		}
		return res, ctx.Err()
	}

}
