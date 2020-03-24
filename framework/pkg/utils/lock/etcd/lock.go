package etcdlock

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/db/etcd"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

// Exposed errors.
var (
	ErrEmptyKey = errors.New("empty key")
)

const (
	GlobalPrefix = "/micro_lock/"
)

// Locker is the client for acquiring distributed locks from etcd. It should be
// created from NewLocker() function.
type Lock struct {
	cli    *clientv3.Client
	sesOpt []concurrency.SessionOption
}

type DuMutex struct {
	originKey string
	fullKey   string
	session   *concurrency.Session
	mutex     *concurrency.Mutex
	acquire   bool
}

func DefaultLock() (*Lock, error) {
	cli, err := etcd.GetDefaultClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return NewLock(cli), nil
}

// NewLocker creates a Locker according to the given options.
func NewLock(cli *clientv3.Client, sessionOpt ...concurrency.SessionOption) *Lock {
	return &Lock{
		cli:    cli,
		sesOpt: sessionOpt,
	}
}

// Lock acquires a distributed lock for the specified resource from etcd v3.
func (l *Lock) NewMutex(OriginKey string) (*DuMutex, error) {
	if OriginKey == "" {
		return nil, errors.WithStack(ErrEmptyKey)
	}
	// create two separate sessions for lock competition
	se, err := concurrency.NewSession(l.cli)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fullKey := GlobalPrefix + OriginKey
	m := concurrency.NewMutex(se, fullKey)

	return &DuMutex{
		originKey: OriginKey,
		fullKey:   fullKey,
		session:   se,
		mutex:     m,
	}, nil

}

//none block
func (d *DuMutex) TryLock(c context.Context) error {
	err := d.mutex.TryLock(c)
	if err == nil {
		d.acquire = true
	}
	return errors.WithStack(err)
}

//block
func (d *DuMutex) Lock(c context.Context) error {
	err := d.mutex.Lock(c)
	if err == nil {
		d.acquire = true
	}
	return errors.WithStack(err)
}

//unlock and close session
func (d *DuMutex) Release(c context.Context) error {
	var err error
	if d.acquire {
		err = d.mutex.Unlock(c)
	}
	if err != nil {
		err = fmt.Errorf("unlock err (%v)", err)
	}

	if err2 := d.session.Close(); err2 != nil {
		err = fmt.Errorf("unlock err (%v), session close err (%v)", err, err2)
	}

	return errors.WithStack(err)
}

func (d *DuMutex) OriginKey() string {
	return d.originKey
}

func (d *DuMutex) FullKey() string {
	return d.fullKey
}

func (d *DuMutex) Session() *concurrency.Session {
	return d.session
}

func (d *DuMutex) Mutex() *concurrency.Mutex {
	return d.mutex
}
