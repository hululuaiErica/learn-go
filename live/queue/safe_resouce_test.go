package queue

import (
	"testing"
)

func TestSafeResource1_AddV1(t *testing.T) {
	s := SafeResource1{}
	s.AddV2(123)
	s.AddV2(234)
}
