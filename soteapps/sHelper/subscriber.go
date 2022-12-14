package sHelper

import (
	"encoding/json"
	"fmt"

	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

const BSLSTREAMNAME = "business-service-layer"
const SALSTREAMNAME = "service-access-layer"

type Subscriber struct {
	Run          *Run
	StreamName   string
	ConsumerName string
	Subject      string
	Schema       *Schema
	Listener     MessageListener

	PullSubscribe   func() sError.SoteError
	GetConsumerInfo func() (*ConsumerInfo, sError.SoteError)
	Fetch           func() ([]Msg, sError.SoteError)
	Publish         func(message interface{}, subject ...string) sError.SoteError
	PublishMessage  func(header RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError

	DoFetch func(consumerInfo *ConsumerInfo) sError.SoteError
}

func NewSubscriber(r *Run, consumerName, subject string, streamName ...string) *Subscriber {
	sLogger.DebugMethod()
	var (
		s Subscriber
	)
	s = Subscriber{
		Run:             r,
		StreamName:      BSLSTREAMNAME,
		ConsumerName:    consumerName,
		Subject:         subject,
		PullSubscribe:   s.subscribe,
		GetConsumerInfo: s.getConsumerInfo,
		Fetch:           s.fetch,
		Publish:         s.publish,
		PublishMessage:  s.publishMessage,
		DoFetch:         s.doFetch,
	}
	if len(streamName) == 1 {
		s.StreamName = streamName[0]
	}
	return &s
}

func (s *Subscriber) Start(msg *Msg) {
	sLogger.Info(fmt.Sprintf("Start Subscription[%v] Subject: %s, Index: %v", msg.Id(), msg.Subject, msg.Index()))
}

func (s *Subscriber) End(msg *Msg, soteErr sError.SoteError) {
	s.Run.returnChain <- &ReturnChain{
		s:       s,
		msg:     msg,
		soteErr: soteErr,
	}
}

func (s *Subscriber) subscribe() sError.SoteError {
	sLogger.DebugMethod()
	return s.Run.myMMPtr.PullSubscribe(s.Subject, s.ConsumerName, s.Run.Env.TestMode)
}

func (s *Subscriber) getConsumerInfo() (*ConsumerInfo, sError.SoteError) {
	sLogger.DebugMethod()
	return s.Run.myMMPtr.GetConsumerInfo(s.StreamName, s.ConsumerName, s.Run.Env.TestMode)
}

func (s *Subscriber) doFetch(consumerInfo *ConsumerInfo) sError.SoteError {
	sLogger.DebugMethod()
	return s.Run.myMMPtr.Fetch(s.ConsumerName, int(consumerInfo.NumPending), false, s.Run.Env.TestMode)
}

func (s *Subscriber) fetch() (messages []Msg, soteErr sError.SoteError) {
	sLogger.DebugMethod()
	var (
		consumerInfo *ConsumerInfo
	)
	s.Run.myMMPtr.Messages = nil // https://sote.myjetbrains.com/youtrack/issue/DO20-233
	if consumerInfo, soteErr = s.GetConsumerInfo(); soteErr.ErrCode == nil && int(consumerInfo.NumPending) > 0 {
		if soteErr = s.DoFetch(consumerInfo); soteErr.ErrCode == nil {
			for index, msg := range s.Run.myMMPtr.Messages {
				messages = append(messages, Msg{
					Subject: msg.Subject,
					Header:  msg.Header,
					Data:    msg.Data,
					index:   index,
					uuid:    UUID(UUIDKind.Short),
				})
				msg.Ack()
			}
		}
	}
	return
}

func (s *Subscriber) publish(message interface{}, subject ...string) sError.SoteError {
	sLogger.DebugMethod()
	if len(subject) == 0 {
		subject = []string{s.Subject}
	}
	return s.Run.myMMPtr.Publish(subject[0], fmt.Sprint(message), s.Run.Env.TestMode)
}

func (s *Subscriber) publishMessage(header RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
	sLogger.DebugMethod()
	m := map[string]interface{}{
		"message-id": header.MessageId,
	}
	if soteErr.ErrCode != nil {
		m["error"] = soteErr
	} else {
		m["message"] = message
	}
	data, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return NewError().InvalidJson(fmt.Sprint(message))
	}
	return s.Publish(string(data), fmt.Sprintf("%v.%v", header.OrganizationId, header.AwsUserName))
}
