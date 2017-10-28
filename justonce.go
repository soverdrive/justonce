package justonce

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

var (
	ErrNoUniqueSeed = fmt.Errorf("No UniqueSeed provided")
	ErrNoStorage    = fmt.Errorf("No Storage provided")
	ErrNoUniqueID   = fmt.Errorf("No UniqueID ever generated")
	ErrDuplication  = fmt.Errorf("Duplication detected!")
)

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string, exp int) error
	Delete(key string) error
}

var defaultStorage Storage

func Init(o Storage) {
	defaultStorage = o

	DefaultParams.KVStorage = defaultStorage
}

type justonceInstance struct {
	uniqueID         string
	instanceCreation time.Time
	sleepDuration    time.Duration
	dataStore        Storage
}

func (d justonceInstance) validate() error {
	if d.uniqueID == "" {
		return ErrNoUniqueID
	}

	return nil
}

func (d justonceInstance) PreventDuringInterval(key string, seconds int) error {
	// Check whether a key is already cached
	// This `get` is for client that store the key with timelimit longer than timesleep
	//  or if `PreventDuringInterval` function is called when doing timesleep
	gotCacheVal, _ := d.dataStore.Get(key)
	if gotCacheVal != "" {
		return ErrDuplication
	}

	err := d.dataStore.Set(key, d.uniqueID, seconds)
	if err != nil {
		return err
	}

	time.Sleep(d.sleepDuration)

	gotCacheVal, err = d.dataStore.Get(key)
	if err != nil {
		return err
	}

	if gotCacheVal != d.uniqueID {
		return ErrDuplication
	}

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
	TakeANap:        20 * time.Millisecond,
	isDefault:       true,
}

type Params struct {
	UniqueGenerator UniqueFunc
	UniqueSeed      interface{}
	TakeANap        time.Duration
	KVStorage       Storage
	isDefault       bool
}

func (p Params) validate() error {
	if !p.isDefault && p.UniqueSeed == nil {
		return ErrNoUniqueSeed
	}

	if p.KVStorage == nil {
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

	d.sleepDuration = p.TakeANap
	d.dataStore = p.KVStorage

	if err := d.validate(); err != nil {
		return d, err
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
