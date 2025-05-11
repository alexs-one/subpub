package subpub

import (
	"context"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) Subscription
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subscriber struct {
	callback MessageHandler
	ch       chan interface{}
	stop     chan struct{}
}

type subPub struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscriber
	closed      bool
	closeCh     chan struct{}
	shutdownWg  sync.WaitGroup
}

func NewSubPub() SubPub {
	return &subPub{
		subscribers: make(map[string][]*subscriber),
		closeCh:     make(chan struct{}),
	}
}

func (sp *subPub) Subscribe(subject string, cb MessageHandler) Subscription {
	sub := &subscriber{
		callback: cb,
		ch:       make(chan interface{}, 10),
		stop:     make(chan struct{}),
	}

	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.subscribers[subject] = append(sp.subscribers[subject], sub)
	sp.shutdownWg.Add(1)
	sp.mu.Unlock()

	go func() {
		defer sp.shutdownWg.Done()
		for {
			select {
			case msg := <-sub.ch:
				cb(msg)
			case <-sub.stop:
				return
			case <-sp.closeCh:
				return
			}
		}
	}()

	return &subscription{
		sp:         sp,
		subject:    subject,
		subscriber: sub,
	}
}

func (sp *subPub) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	if sp.closed {
		return nil
	}
	for _, sub := range sp.subscribers[subject] {
		select {
		case sub.ch <- msg:
		default:

		}
	}
	return nil
}

func (sp *subPub) Close(ctx context.Context) error {
	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.closed = true
	close(sp.closeCh)
	for _, subs := range sp.subscribers {
		for _, sub := range subs {
			close(sub.stop)
		}
	}
	sp.mu.Unlock()

	done := make(chan struct{})
	go func() {
		sp.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type subscription struct {
	sp         *subPub
	subject    string
	subscriber *subscriber
}

func (s *subscription) Unsubscribe() {
	s.sp.mu.Lock()
	defer s.sp.mu.Unlock()
	subs := s.sp.subscribers[s.subject]
	for i, sub := range subs {
		if sub == s.subscriber {
			s.sp.subscribers[s.subject] = append(subs[:i], subs[i+1:]...)
			close(sub.stop)
			break
		}
	}
}
