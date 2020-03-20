package log

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/zuiqiangqishao/framework/pkg/mq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

var (
	_defaultConf *KafkaConfig
	_mu          sync.RWMutex
)

//kafka日志文件配置
type KafkaConfig struct {
	Enable    bool
	Topic     string
	Key       string
	Partition int

	Level      string //debug  info  warn  error fatal
	Encode     string //json or console
	CallFull   bool   //whether full call path or short path, default is short
	EnableHost bool   //Add host name and other additional fields
	Brokers    []string
}

type KafWriter struct {
	w *kafka.Writer
}

func DefaultKafkaCore() zapcore.Core {
	return _defaultConf.NewKafkaCore()
}

func (k *KafWriter) Write(b []byte) (n int, err error) {
	s:=make([]byte,len(b))
	copy(s,b) //异步写入要copy slice数据到新的地方，不然这个slice马上要被下一个core进行写了
	msg := kafka.Message{
		Value: s,
	}
	//ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	err = k.w.WriteMessages(context.Background(), msg)

	return
}

func (f *KafkaConfig) NewKafkaCore() zapcore.Core {
	if !f.Enable {
		return nil
	}
	enCfg := zap.NewProductionEncoderConfig()
	if f.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	enCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	levelEnable := zap.NewAtomicLevelAt(convertLogLevel(f.Level))

	encoder := zapcore.NewJSONEncoder(enCfg)
	if f.Encode == "console" {
		encoder = zapcore.NewConsoleEncoder(enCfg)
	}

	kp := newKafkaWriter(mq.GetKafkaConf().Brokers, f.Topic)

	core := zapcore.NewCore(encoder, zapcore.AddSync(&KafWriter{kp}), levelEnable)
	if f.EnableHost {
		core = AddHostInfo(core)
	}
	return core
}

//消息分发策略默认使用轮训策略
func newKafkaWriter(kafkaURL []string, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: kafkaURL,
		Topic:   topic,
		Async:   true,
	})
}

func GetKafkaConf() *KafkaConfig {
	_mu.RLock()
	defer _mu.RUnlock()
	t := *_defaultConf
	return &t
}

func SetKafkaConf(c *KafkaConfig) {
	_mu.Lock()
	defer _mu.Unlock()
	_defaultConf = c
}
