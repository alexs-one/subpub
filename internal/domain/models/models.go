package models

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}
