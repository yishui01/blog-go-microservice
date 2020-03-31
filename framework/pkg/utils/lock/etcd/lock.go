package etcdlock

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/db/etcd"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"math/rand"
	"time"
)

// Exposed errors.
var (
	ErrEmptyKey = errors.New("empty key")
)

const (
	GlobalPrefix = "/micro_lock/"
)

// Conn is the DuMutex's connection and session config
type Conn struct {
	cli    *clientv3.Client
	sesOpt []concurrency.SessionOption
}

type DuMutex struct {
	conf      *Conn
	originKey string
	fullKey   string
	session   *concurrency.Session
	mutex     *concurrency.Mutex
	acquire   bool
}

//DistributeTryLock try acquires a distributed lock  from etcd v3.
func DistributeTryLock(c context.Context, OriginKey string, ttl int, failLockSleepMill int) (*DuMutex, error) {
	cli, err := etcd.GetDefaultClient()
	if err != nil {
		return nil, err
	}
	conn := NewConn(cli, concurrency.WithContext(c), concurrency.WithTTL(ttl))
	m, err := conn.NewMutex(OriginKey)
	if err != nil {
		e := conn.Close()
		if e != nil {
			err = errors.Wrap(e, " conn.NewMutex err:"+err.Error())
		}
		return nil, err
	}
	//nonblock
	if err3 := m.TryLock(c); err3 != nil {
		m.Release(c)
		if errors.Cause(err) == concurrency.ErrLocked {
			if failLockSleepMill < 0 {
				failLockSleepMill = 300
			}
			delay := rand.Intn(failLockSleepMill)
			if delay < failLockSleepMill/2 {
				delay += failLockSleepMill / 2
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
			return nil, nil
		}
		return nil, err
	}

	//lock success
	return m, nil
}

// NewConn save SessionOption and cli to create NewMutex
func NewConn(cli *clientv3.Client, sessionOpt ...concurrency.SessionOption) *Conn {
	return &Conn{
		cli:    cli,
		sesOpt: sessionOpt,
	}
}

// NewMutex create a DuMutex to Lock Or TryLock
func (l *Conn) NewMutex(OriginKey string) (*DuMutex, error) {
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
		conf:      l,
		originKey: OriginKey,
		fullKey:   fullKey,
		session:   se,
		mutex:     m,
	}, nil

}

func (l *Conn) Close() error {
	var err error
	if l.cli != nil {
		err = l.cli.Close()
	}
	l.cli = nil
	return errors.WithStack(err)
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

//unlock and close session and connect
func (d *DuMutex) Release(c context.Context) {
	var err error
	if d.acquire {
		err = d.mutex.Unlock(c)
	}
	if err != nil {
		err = fmt.Errorf("unlock err (%v)", err)
	}

	//close session
	if err2 := d.session.Close(); err2 != nil {
		err = fmt.Errorf("(%v), session close err (%v)", err, err2)
	}

	//close clientv3 connect
	if err3 := d.conf.Close(); err3 != nil {
		err = fmt.Errorf(" (%v), clientv3 close err (%v)", err, err3)
	}

	if err != nil {
		log.SugarWithContext(c).Errorf("DuMutex.Release Err(%#+v)", errors.WithStack(err))
	}
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
