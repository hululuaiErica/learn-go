package kafkax

import (
	"github.com/Shopify/sarama"
)

type Producer struct {
}

func (p *Producer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (p *Producer) BeginTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) CommitTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) AbortTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	//TODO implement me
	panic("implement me")
}

func (p *Producer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	//TODO implement me
	panic("implement me")
}
