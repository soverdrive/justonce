package justonce_test

import (
	"sync"
	"testing"

	"log"

	"github.com/soverdrive/justonce"
)

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
	var temp = new(TestingStorage)
	storage := temp.Init()

	justonce.Init(storage)

	var err = make([]error, 2)
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go func(e error) {
		var key = "test01"
		instance, err := justonce.New(justonce.DefaultParams)
		log.Println(err)
		err = instance.PreventDuringInterval(key, 10)
		log.Println(instance.GetUniqueID(), instance.GetInstanceCreation())
		e = err
		log.Println(err)
		wg.Done()
	}(err[0])

	wg.Add(1)
	go func(e error) {
		var key = "test01"
		instance, err := justonce.New(justonce.DefaultParams)
		log.Println(err)
		err = instance.PreventDuringInterval(key, 10)
		log.Println(instance.GetUniqueID(), instance.GetInstanceCreation())
		e = err
		log.Println(err)
		wg.Done()
	}(err[1])

	wg.Wait()
	log.Println(err)
	for i := 0; i < 2; i++ {
		if err[i] != nil {
			t.Errorf("%+v\n", err[i])
		}
	}
}
