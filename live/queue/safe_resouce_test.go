package queue

import (
	"fmt"
	"testing"
)

func TestSafeResource1_AddV1(t *testing.T) {
	s := SafeResource1{}
	s.AddV2(123)
	s.AddV2(234)
}

type sliceTest struct {
	data []int
}

func (st sliceTest) add(val int) {
	st.data = append(st.data, val)
	fmt.Printf("add %p, s.data %v \n", &st, st.data)
}

func TestSlice(t *testing.T) {
	s := sliceTest{}
	s.add(123)
	fmt.Printf("%p, s.data %v \n", &s, s.data)

	s1 := &sliceTest{}
	s1.add(123)
	fmt.Println(s1.data)
}

type mapTest struct {
	data map[string]string
}

func (m mapTest) add(key, val string) {
	m.data[key] = val
}

func TestMap(t *testing.T) {
	m := mapTest{
		data: map[string]string{},
	}
	m.add("key1", "value1")
	fmt.Println(m.data)

	m1 := &mapTest{
		data: map[string]string{},
	}
	m1.add("key2", "value2")
	fmt.Println(m1.data)
}
