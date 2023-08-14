package kafkax

import (
	"context"
	"github.com/Shopify/sarama"
)

type ShadowTopicProducer struct {
	sarama.SyncProducer
}

func (p *ShadowTopicProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	ctx, ok := msg.Metadata.(context.Context)
	if ok && isShadow(ctx) {
		msg.Topic = msg.Topic + "_shadow"
	}
	return p.SyncProducer.SendMessage(msg)
}
