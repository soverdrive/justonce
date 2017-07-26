package justonce

import (
	"fmt"
	"math/rand"
	"time"

	"log"

	"github.com/oklog/ulid"
)

var (
	ErrNoUniqueSeed = fmt.Errorf("No UniqueSeed provided\n")
	ErrNoStorage    = fmt.Errorf("No Storage provided\n")
)

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string, exp int) error
	Delete(key string) error
}

type Instance struct {
	Storage
}

var defaultInstance Instance

func Init(o Storage) {
	defaultInstance = Instance{
		Storage: o,
	}
}

func SpawnStorage(o Storage) Instance {
	return Instance{
		Storage: o,
	}
}

type justonceInstance struct {
	uniqueID         string
	instanceCreation time.Time
	sleepDuration    time.Duration
}

func (d justonceInstance) PreventDuringInterval(key string, seconds int) error {
	err := defaultInstance.Set(key, d.uniqueID, seconds)
	if err != nil {
		return err
	}

	time.Sleep(d.sleepDuration)

	gotCacheVal, err := defaultInstance.Get(key)
	if err != nil {
		defaultInstance.Delete(key)
		return err
	}

	if gotCacheVal != d.uniqueID {
		return fmt.Errorf("Duplication detected! Found %v, expected %v\n", gotCacheVal, d.uniqueID)
	}

	log.Println(gotCacheVal, d.uniqueID)

	return nil
}

func (d justonceInstance) GetUniqueID() string {
	return d.uniqueID
}

func (d justonceInstance) GetInstanceCreation() time.Time {
	return d.instanceCreation
}

type UniqueFunc func(interface{}) string

var DefaultParams = Params{
	UniqueGenerator: getUniqueID,
	TakeANap:        2 * time.Second,
	KVStorage:       defaultInstance,
	isDefault:       true,
}

type Params struct {
	UniqueGenerator UniqueFunc
	UniqueSeed      interface{}
	TakeANap        time.Duration
	KVStorage       Instance
	isDefault       bool
}

func (p Params) validate() error {
	if !p.isDefault && p.UniqueSeed == nil {
		return ErrNoUniqueSeed
	}

	if p.KVStorage == (Instance{}) {
		return ErrNoStorage
	}

	return nil
}

func New(p Params) (justonceInstance, error) {
	var d = justonceInstance{}

	if err := p.validate(); err != nil {
		return d, err
	}

	d.instanceCreation = time.Now()

	if p.isDefault {
		d.uniqueID = p.UniqueGenerator(d.instanceCreation)
	} else {
		d.uniqueID = p.UniqueGenerator(p.UniqueSeed)
	}

	return d, nil
}

func getUniqueID(t interface{}) string {
	switch t.(type) {
	case time.Time:
		return getULID(t.(time.Time))
	}

	return ""
}

func getULID(t time.Time) string {
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
