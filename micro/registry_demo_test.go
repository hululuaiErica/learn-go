package micro

type Registry interface {
	Subscribe() <- chan Event
}

type Event struct {

}