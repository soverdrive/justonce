package justonce_test

import (
	"sync"
	"testing"

	"github.com/soverdrive/justonce"
)

// type Storage interface {
// 	Get(key string) (string, error)
// 	Set(key, value string, exp int) error
// 	Delete(key string) error
// }

type TestingStorage struct {
	s    map[string]string
	lock sync.Mutex
}

func (t *TestingStorage) Init() *TestingStorage {
	t.s = make(map[string]string)

	return t
}

func (t *TestingStorage) Get(key string) (string, error) {
	var v string

	t.lock.Lock()
	v = t.s[key]
	t.lock.Unlock()

	return v, nil
}

func (t *TestingStorage) Set(key, value string, exp int) error {
	t.lock.Lock()
	t.s[key] = value
	t.lock.Unlock()

	return nil
}

func (t *TestingStorage) Delete(key string) error {
	t.lock.Lock()
	delete(t.s, key)
	t.lock.Unlock()

	return nil
}

func TestDuplicatePrevention(t *testing.T) {
	var temp TestingStorage
	storage := temp.Init()

	justonce.Init(storage)

	var err [2]*error
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go func(e *error) {
		var key = "test01"
		instance, _ := justonce.New(justonce.DefaultParams)
		t.Errorf(instance.GetUniqueID())
		err1 := instance.PreventDuringInterval(key, 60)
		e = &err1
		wg.Done()
	}(err[0])
	wg.Add(1)
	go func(e *error) {
		var key = "test02"
		instance, _ := justonce.New(justonce.DefaultParams)
		t.Errorf(instance.GetUniqueID())
		err2 := instance.PreventDuringInterval(key, 60)
		e = &err2
		wg.Done()
	}(err[1])

	wg.Wait()
	for i := 0; i < 2; i++ {
		if err[i] != nil {
			t.Errorf("%+v\n", err[i])
		}
	}
}
