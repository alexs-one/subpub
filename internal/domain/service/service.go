package service

import (
	"alex-one/subscriber/internal/domain/models"
	"alex-one/subscriber/internal/pkg/subpub"
)

type SubPub interface {
	subpub.SubPub
}

type SubPuber struct {
	sp SubPub
}

func NewSubPuberService(sp SubPub) *SubPuber {
	return &SubPuber{
		sp: sp,
	}
}

func (s *SubPuber) Subscribe(subject string, cb models.MessageHandler) models.Subscription {
	return s.sp.Subscribe(subject, subpub.MessageHandler(cb))
}

func (s *SubPuber) Publish(subject string, msg interface{}) error {
	return s.sp.Publish(subject, msg)
}
