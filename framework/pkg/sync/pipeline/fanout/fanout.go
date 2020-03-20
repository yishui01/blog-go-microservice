package fanout

import (
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"runtime"
	"sync"
)

//golang 的并发模式之管道模式  https://blog.golang.org/pipelines
//其他参考的项目 https://github.com/sunfmin/fanout
//其实就是把多个input的输入全部丢到一个chan中，然后开多个worker全部从这个chan中取任务，每个worker执行完之将结果发送到全局的res chan中，
//外部调用者只管把任务往input里面丢，然后从res这个chan中取结果就行了。
//就是一个中间层解耦了生产者和消费者，有效降低了生产-消费模型发生deadlock的概率以及编写难度
var (
	// ErrFull chan full.
	ErrFull   = errors.New("fanout: chan full")
	traceTags = []opentracing.Tag{
		{Key: string(ext.SpanKind), Value: "background"},
		{Key: string(ext.Component), Value: "sync/pipeline/fanout"},
	}
)

type options struct {
	worker int //worker数量
	buffer int //中心任务队列的长度
}

// Option fanout option
type Option func(*options)

// Worker specifies the worker of fanout
func Worker(n int) Option {
	if n <= 0 {
		panic("fanout: worker should > 0")
	}
	return func(o *options) {
		o.worker = n
	}
}

// Buffer specifies the buffer of fanout
func Buffer(n int) Option {
	if n <= 0 {
		panic("fanout: buffer should > 0")
	}
	return func(o *options) {
		o.buffer = n
	}
}

//任务node
type item struct {
	f   func(c context.Context)
	ctx context.Context
}

// Fanout async consume data from chan.
type Fanout struct {
	name    string
	ch      chan item //任务队列
	options *options
	waiter  sync.WaitGroup

	ctx    context.Context
	cancel func()
}

// New new a fanout struct.
func New(name string, opts ...Option) *Fanout {
	if name == "" {
		name = "anonymous"
	}
	o := &options{
		worker: 1,
		buffer: 1024,
	}
	for _, op := range opts {
		op(o)
	}
	c := &Fanout{
		ch:      make(chan item, o.buffer),
		name:    name,
		options: o,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.waiter.Add(o.worker)
	for i := 0; i < o.worker; i++ {
		go c.proc()
	}
	return c
}

func (c *Fanout) proc() {
	defer c.waiter.Done()
	for {
		select {
		case t := <-c.ch:
			wrapFunc(t.f)(t.ctx)
			//todo... addMetric
			//_metricChanSize.Set(float64(len(c.ch)), c.name)
			//_metricCount.Inc(c.name)
		case <-c.ctx.Done():
			return
		}
	}
}

func wrapFunc(f func(c context.Context)) (res func(context.Context)) {
	res = func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64*1024)
				buf = buf[:runtime.Stack(buf, false)]
				log.SugarWithContext(ctx).Errorf("panic in fanout proc, err: %s, stack: %s", r, buf)
			}
		}()
		f(ctx)
		if span := opentracing.SpanFromContext(ctx); span != nil {
			span.Finish()
		}
	}
	return
}

// Do save a callback func.
func (c *Fanout) Do(ctx context.Context, f func(ctx context.Context)) (err error) {
	if f == nil || c.ctx.Err() != nil {
		return c.ctx.Err()
	}
	nakeCtx := metadata.WithContext(ctx)
	if span := opentracing.SpanFromContext(ctx); span != nil { //如果里面有trace，就接着加一段，没有就不加
		span := span.Tracer().StartSpan("Fanout:Do")
		setTags(span)
		nakeCtx = opentracing.ContextWithSpan(nakeCtx, span)
	}
	select {
	case c.ch <- item{f: f, ctx: nakeCtx}:
	default:
		err = ErrFull
	}
	//todo... addMetric
	//_metricChanSize.Set(float64(len(c.ch)), c.name)
	return
}

// Close close fanout
func (c *Fanout) Close() error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.cancel()
	c.waiter.Wait()
	return nil
}

func setTags(span opentracing.Span) {
	for _, v := range traceTags {
		span.SetTag(v.Key, v.Value)
	}
}
