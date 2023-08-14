package kafkax

import (
	"context"
	"github.com/Shopify/sarama"
	"sync"
)

// 练习实现这个接口，19:40
// 注意，是不同 Kafka 集群
type Producer struct {
	live       sarama.SyncProducer
	shadow     sarama.SyncProducer
	l          sync.Mutex
	isShadowTx bool
	isLiveTx   bool
}

func (p *Producer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	//var stressTestFlag sarama.RecordHeader
	//for _, h := range msg.Headers {
	//	if string(h.Key) == "stress-test" {
	//		stressTestFlag = h
	//		break
	//	}
	//}
	//if string(stressTestFlag.Value) == "true" {
	//	return p.shadow.SendMessage(msg)
	//}
	ctx, ok := msg.Metadata.(context.Context)
	if ok && isShadow(ctx) {
		return p.shadow.SendMessage(msg)
	}
	return p.live.SendMessage(msg)
}

func (p *Producer) SendMessageWithCtx(ctx context.Context, msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	if isShadow(ctx) {
		return p.shadow.SendMessage(msg)
	}
	return p.live.SendMessage(msg)
}

func (p *Producer) SendMessages(msgs []*sarama.ProducerMessage) error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) TxnStatus() sarama.ProducerTxnStatusFlag {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) IsTransactional() bool {
	p.l.Lock()
	defer p.l.Unlock()
	return p.isShadowTx || p.isLiveTx
}

func (p *Producer) BeginTxn() error {
	p.l.Lock()
	defer p.l.Unlock()
	return nil
	// 你怎么知道，是 Live 还是 shadow?
}

func (p *Producer) CommitTxn() error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.isShadowTx {
		p.isShadowTx = false
		return p.shadow.CommitTxn()
	}
	if p.isLiveTx {
		p.isLiveTx = false
		return p.live.CommitTxn()
	}
	return p.live.CommitTxn()
}

func (p *Producer) AbortTxn() error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.isShadowTx {
		p.isShadowTx = false
		return p.shadow.CommitTxn()
	}
	if p.isLiveTx {
		p.isLiveTx = false
		return p.live.CommitTxn()
	}
	return p.live.CommitTxn()
}

func (p *Producer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	p.l.Lock()
	defer p.l.Unlock()
	//TODO implement me
	panic("implement me")
}

func (p *Producer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	//ctx, ok := msg.Headers.(context.Context)
	p.l.Lock()
	defer p.l.Unlock()
	ctx := context.Background()
	if isShadow(ctx) {
		if !p.shadow.IsTransactional() {
			p.shadow.BeginTxn()
		}
		p.isShadowTx = true
		return p.shadow.AddMessageToTxn(msg, groupId, metadata)
	}
	if !p.live.IsTransactional() {
		p.live.BeginTxn()
	}
	return p.live.AddMessageToTxn(msg, groupId, metadata)
}

func isShadow(ctx context.Context) bool {
	return ctx.Value("stress-test") == "true"
}
