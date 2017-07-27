package justonce_test

import (
	"sync"
	"testing"

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

	var errChan = make(chan error, 2)
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		var key = "test01"
		instance, _ := justonce.New(justonce.DefaultParams)

		err := instance.PreventDuringInterval(key, 10)
		errChan <- err
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		var key = "test01"
		instance, _ := justonce.New(justonce.DefaultParams)

		err := instance.PreventDuringInterval(key, 10)
		errChan <- err
		wg.Done()
	}()

	wg.Wait()
	close(errChan)

	var duplicationDetected bool
	for v := range errChan {
		if v != nil {
			duplicationDetected = true
		}
	}

	if !duplicationDetected {
		t.Errorf("No duplication detected! Duplication should happened")
	}
}
