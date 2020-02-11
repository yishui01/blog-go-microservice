package filewriter

import (
	"fmt"
	"strings"
	"time"
)

const (
	RotateDaily = "2006-01-02"
)

//配置项，主要是轮换配置项，当然还有其他的
type rotateOption struct {
	//轮换配置
	RotateFormat   string        //轮替格式，默认每天轮替 "2006-01-02"
	MaxFile        int           //最多保留存档文件数，0为不限制文件数，默认为0
	MaxSize        int64         //单个文件最大体积,字节为单位,默认1G
	RotateInterval time.Duration //检查轮换的时间间隔，默认10s

	//其他配置
	UseBuffer     bool          //是否使用ch+buffer异步写入磁盘,默认不使用，不使用时每条log直接刷到磁盘
	ChanSize      int           //FileWriter里面ch的个数
	FlushInterval time.Duration //异步将缓冲区刷新到磁盘时间间隔，默认10 毫秒 刷新一次
	FlushMaxSize  int           //缓冲区到达该容量时强制刷新到磁盘，单位字节，默认1MB
	WriteTimeout  time.Duration //默认为0，非阻塞写入，写入channel失败直接返回error
}

var defaultOption = rotateOption{
	RotateFormat:   RotateDaily,
	MaxSize:        1 << 30,
	ChanSize:       1024 * 8,
	FlushInterval:  10 * time.Millisecond,
	FlushMaxSize:   1024 * 1024 * 1024, //1MB
	RotateInterval: 10 * time.Second,
}

type Option func(opt *rotateOption)

// RotateFormat e.g 2006-01-02 meaning rotate log file every day.
// NOTE: format can't contain ".", "." will cause panic ヽ(*。>Д<)o゜.
func RotateFormat(format string) Option {
	if strings.Contains(format, ".") {
		panic(fmt.Sprintf("rotate format can't contain '.' format: %s", format))
	}
	return func(opt *rotateOption) {
		opt.RotateFormat = format
	}
}

// MaxFile default 0, 0 meaning unlimit.
func MaxFile(n int) Option {
	return func(opt *rotateOption) {
		opt.MaxFile = n
	}
}

//MaxSize set max size for single log file
// default 1GB, 0 meaning unlimit.
func MaxSize(n int64) Option {
	return func(opt *rotateOption) {
		opt.MaxSize = n
	}
}

//是否异步写入磁盘
func UserBuffer(use bool) Option {
	return func(opt *rotateOption) {
		opt.UseBuffer = use
	}
}

// ChanSize set internal chan size default 8192 use about 64k memory on x64 platfrom static,
// because filewriter has internal object pool, change chan size bigger may cause filewriter use
// a lot of memory, because sync.Pool can't set expire time memory won't free until program exit.
//异步ch的大小
func ChanSize(n int) Option {
	return func(opt *rotateOption) {
		opt.ChanSize = n
	}
}

//异步轮替检查时间
func RotateInterval(t time.Duration) Option {
	if t <= 0 {
		panic("rotateInterval must be positive time.Duration")
	}
	return func(opt *rotateOption) {
		opt.RotateInterval = t
	}
}

//异步写入ch，超时时间传0代表非阻塞IO，失败会直接return error
func WriteTimeout(t time.Duration) Option {
	if t < 0 {
		panic("WriteTimeout can't less than 0")
	}
	return func(opt *rotateOption) {
		opt.WriteTimeout = t
	}
}

//异步定时刷新到磁盘的频率
func FlushInterval(t time.Duration) Option {
	if t <= 0 {
		panic("FlushInterval must be positive time.Duration")
	}
	return func(opt *rotateOption) {
		opt.FlushInterval = t
	}
}

//异步buffer最大size
func FlushMaxSize(s int) Option {
	if s < 0 {
		panic("FlushMaxSize can't less than 0")
	}
	return func(opt *rotateOption) {
		opt.FlushMaxSize = s
	}
}
