package filewriter

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//一个文件输出位置就是一个FileWriter，如果要输出到多个地方的file，就要在外面构建多个FileWriter
type FileWriter struct {
	opt    rotateOption       //轮换配置
	dir    string             //输出目录
	fname  string             //日志文件名称
	ch     chan *bytes.Buffer //日志会首先写入该chan, 随后由另一个goroutine异步落盘
	stdlog *log.Logger        //发生错误时打印一些错误输出到stderr
	pool   *sync.Pool         //ch中的bytes.Buffer池，写日志时先取出一个buffer，将日志copy到该buffer，再将buffer入队到chan

	lastRotateFormat string //最后发生日志轮替的日期，默认创建时会设置为当前日期如： 2018-05-06，默认每天轮替
	lastSplitNum     int    //和lastRotateFormat配合，代表该轮替周期内第几次轮替，初始为第0次，到新的一天时会重新从0计数

	writeLock sync.Mutex //日志轮替、日志写入文件是两个goroutine，避免同时进行，需要锁
	current   *wrapFile  //os.File的包装，每次写入磁盘就是往这个current文件写
	filesList *list.List //该目录下的日志文件链表，链表的每个元素都是一个 rotateItem 结构体

	closed int32 //ch是否关闭
	wg     sync.WaitGroup
}

//每一个文件都会构造出一个rotateItem
type rotateItem struct {
	rotateTime int64  //文件名中的日期
	rotateNum  int    //文件名中轮换到第几个了，默认是第0个
	fname      string //文件名全名
}

//遍历文件夹下的日志文件，每个存档文件构造一个rotateItem，并由旧到新（旧的在头部）组成一个链表并返回
func parseRotateItem(dir, fname, rotateFormat string) (*list.List, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	//将文件名解析为rotateItem结构体
	parse := func(s string) (rt rotateItem, err error) {
		// remove filename and left "."   e.g.  error.log.2018-09-12.001 -> 2018-09-12.001
		rt.fname = s
		s = strings.TrimLeft(s[len(fname):], ".")
		seqs := strings.Split(s, ".")
		var t time.Time
		switch len(seqs) {
		case 2:
			if rt.rotateNum, err = strconv.Atoi(seqs[1]); err != nil { //当天轮换到第几次了
				return
			}
			fallthrough
		case 1:
			if t, err = time.Parse(rotateFormat, seqs[0]); err != nil {
				return
			}
			rt.rotateTime = t.Unix() //轮换时间
		}
		return
	}

	var items []rotateItem

	//遍历目录下的所有文件
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), fname) && fi.Name() != fname {
			rt, err := parse(fi.Name())
			if err != nil {
				continue
			}
			items = append(items, rt)
		}
	}

	//按照文件生成时间旧到新排序，旧的排在前面
	sort.Slice(items, func(i, j int) bool {
		if items[i].rotateTime == items[j].rotateTime {
			return items[i].rotateNum > items[j].rotateNum
		}
		return items[i].rotateTime > items[j].rotateTime
	})
	l := list.New()

	for _, item := range items {
		l.PushBack(item)
	}

	return l, nil

}

//包下file，添加size字段，主要用于包裹当前写入的目标文件，日志轮替时候方便检查当前日志size是否已经达到轮替标准
type wrapFile struct {
	fsize int64
	fp    *os.File
}

func (w *wrapFile) size() int64 {
	return w.fsize
}

func (w *wrapFile) write(p []byte) (n int, err error) {
	n, err = w.fp.Write(p)
	w.fsize += int64(n)
	return
}

//打开一个文件并将其包裹成一个struct，加了个size字段，方便记录、返回文件size
func newWrapFile(fpath string) (*wrapFile, error) {
	fp, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	return &wrapFile{fp: fp, fsize: fi.Size()}, nil
}

// New FileWriter A FileWriter is safe for use by multiple goroutines simultaneously.
//创建一个新的FileWriter结构体，说是并发安全的？真的是并发安全的吗？具体有什么用呢？
//需要传入一个完整的带文件名的日志文件路径，可选参数用于修改轮换配置
func New(fullPathName string, fns ...Option) (*FileWriter, error) {
	opt := defaultOption
	for _, fn := range fns {
		fn(&opt)
	}

	fname := filepath.Base(fullPathName)
	if fname == "" {
		return nil, fmt.Errorf("filename can't be empty")
	}
	dir := filepath.Dir(fullPathName)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("%s already exists and not a directory", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s error: %s", dir, err.Error())
		}
	}

	//将当前传入的文件名打开（不存在时创建）后进行包裹，加了个size字段，并设置为FileWriter里面的current
	current, err := newWrapFile(fullPathName)
	if err != nil {
		return nil, err
	}

	stdlog := log.New(os.Stderr, "flog", log.LstdFlags)
	ch := make(chan *bytes.Buffer, opt.ChanSize)

	//遍历该目录下文件，根据日志文件名筛选出存档文件构造出文件链表（并不会将log名字的那个写入文件加进来）
	filesList, err := parseRotateItem(dir, fname, opt.RotateFormat)
	if err != nil {
		filesList = list.New() //清空链表
		stdlog.Printf("parseRotateItem error: %s", err)
	}

	//最后轮替的日期，默认创建时设置为当天时间
	lastRotateFormat := time.Now().Format(opt.RotateFormat)

	var lastSplitNum int
	if filesList.Len() > 0 {
		rt := filesList.Front().Value.(rotateItem)
		//check contains is more easier than compared with timestamp
		if strings.Contains(rt.fname, lastRotateFormat) {
			lastSplitNum = rt.rotateNum //如果有当天轮替的日志，设置轮替到第几次了
		}
	}

	fw := &FileWriter{
		opt:    opt,
		dir:    dir,
		fname:  fname,
		stdlog: stdlog,
		ch:     ch,
		pool:   &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},

		lastSplitNum:     lastSplitNum,     //最后轮替日期内轮替到第几次了
		lastRotateFormat: lastRotateFormat, //最后轮替日期

		filesList: filesList,
		current:   current,
	}

	fw.wg.Add(1)
	go fw.daemon() //监听chan 写入、定时轮替日志

	return fw, nil
}

//定时轮替、将ch中数据取出并异步刷新到磁盘
func (f *FileWriter) daemon() {
	aggsbuf := &bytes.Buffer{}                    //创建一个写入缓冲区，将ch数据取出先放到这里，到时候一次性刷回磁盘
	tk := time.NewTicker(f.opt.RotateInterval)    //轮替检查间隔
	aggstk := time.NewTicker(f.opt.FlushInterval) //落盘刷新间隔
	maxBrokerSize := f.opt.FlushMaxSize           //缓冲区到达该容量时强制写入到磁盘（默认1M）

	var err error

	flushToFile := func() {
		if err = f.write(aggsbuf.Bytes()); err != nil {
			f.stdlog.Printf("write log error: %s", err)
		}
		aggsbuf.Reset()
	}

	//这里将checkRotate单独提出，因为将其放到和写入文件逻辑一个select中的话，如果ch一直有写入的话，轮替会一直得不到进行，造成饥饿
	//但是提出后就会有一个问题，checkRotate和flushToFile不能同时进行，因为checkRotate会关闭写入的file结构体，导致flushToFile失败，互斥锁解决
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				f.stdlog.Printf("recover checkRotate panic, error: %s", err)
			}
		}()
		for {
			select {
			case t := <-tk.C:
				f.checkRotate(t) //内部互斥锁,和flushToFile互斥
			case <-done:
				return
			}
		}
	}()

	for {
		//是否强制刷新
		if aggsbuf.Len() >= maxBrokerSize {
			flushToFile()
		}

		select {
		case buf, ok := <-f.ch:
			if ok {
				aggsbuf.Write(buf.Bytes()) //将ch的数据写到aggsbuf中
				f.putBuf(buf)
			}
		case <-aggstk.C: //定时刷到磁盘
			if aggsbuf.Len() > 0 {
				flushToFile()
			}
		}

		//每次都检查chan是否已经关闭，没关闭就继续监听
		if atomic.LoadInt32(&f.closed) != 1 {
			continue
		}

		// read all buf from channel and break loop
		if err = f.write(aggsbuf.Bytes()); err != nil {
			f.stdlog.Printf("write log error: %s", err)
		}
		for buf := range f.ch {
			if err = f.write(buf.Bytes()); err != nil {
				f.stdlog.Printf("write log error: %s", err)
			}
			f.putBuf(buf)
		}
		break
	}
	close(done)
	if f.current.fp != nil {
		f.current.fp.Close() //关闭current
	}
	f.wg.Done()
}

//向文件中写入数据，这里只是将数据写入到chan中
//chan中的数据是异步写入磁盘的，chan里的数据有可能还未被读取，所以这里返回的是理想状态下的写入字节数
func (f *FileWriter) Write(p []byte) (int, error) {
	// atomic is not necessary
	//如果管道已经关了，那就写不进去了。。。
	if atomic.LoadInt32(&f.closed) == 1 {
		f.stdlog.Printf("%s", p)
		return 0, fmt.Errorf("filewriter already closed")
	}

	//由于写入文件是异步的，所以这里msg传进来后马上复制到内部的buff中，防止外部修改
	buf := f.getBuf()
	buf.Write(p)

	//如果写入超时时间为0，那就是非阻塞写入，写不进去直接返回err
	if f.opt.WriteTimeout == 0 {
		select {
		case f.ch <- buf:
			return len(p), nil
		default:
			// TODO: write discard log to to stdout?
			return 0, fmt.Errorf("log channel is full, discard log")
		}
	}

	//如果设置了超时时间，就是阻塞写入，规定时间内还没写进去就返回err
	timeout := time.NewTimer(f.opt.WriteTimeout)
	select {
	case f.ch <- buf:
		return len(p), nil
	case <-timeout.C:
		// TODO: write discard log to to stdout?
		return 0, fmt.Errorf("write timeout log channel is full, discard log")
	}
}

//关闭异步写入的ch
func (f *FileWriter) Close() error {
	atomic.StoreInt32(&f.closed, 1)
	close(f.ch)
	f.wg.Wait()
	return nil
}

//检查轮替，日期更新或者单个文件过大都会触发轮替
//轮替流程：
//1、先将当前current文件重命名为存档文件
//2、再重新创建一个新的current，和之前的current文件名一样，并赋值给fileWriter的current字段
//3、如果是新的一天则更新lastRotateFormat为新日期并将lastSplitNum置0，如果是当天内多次轮替直接lastSplitNum++
func (f *FileWriter) checkRotate(t time.Time) {
	f.writeLock.Lock() //加锁，防止轮替时还在往文件里面写
	defer f.writeLock.Unlock()
	if f.current.fp == nil {
		return
	}
	formatFname := func(format string, num int) string {
		//num++的原因是由于如果原来有个laravel.2012-02-02.001.log 并且里面已经写了一些日志
		//在当天程序重启后，当前lastSplitNum指向001.log，导致rotate时会重新存档为001.log，覆盖了原来001.log里面的日志
		num++
		//%03d代表不够3位用0补齐
		return fmt.Sprintf("%s.%s.%03d", f.fname, format, num)
	}

	nowDate := t.Format(f.opt.RotateFormat) //当前时间

	//如果有文件数量限制，删除超出保留日期的远古时期文件
	if f.opt.MaxFile != 0 {
		for f.filesList.Len() > f.opt.MaxFile {
			rt := f.filesList.Remove(f.filesList.Front()).(rotateItem)
			fpath := filepath.Join(f.dir, rt.fname)
			if err := os.Remove(fpath); err != nil {
				f.stdlog.Printf("remove file %s error:%s", fpath, err)
			}
		}
	}

	//如果当前时间已经不是filewriter的lastRotateFormat了（默认format是到天数，所以当前是到了新的一天日期）
	//或者当前单个文件体积已经超过了配置的单个文件大小
	//进行轮替
	if nowDate != f.lastRotateFormat || (f.opt.MaxSize != 0 && f.current.size() > f.opt.MaxSize) {
		var err error
		//首先关闭当前写入文件的file结构体，不要再往里写了
		if err = f.current.fp.Close(); err != nil {
			f.stdlog.Printf("close current file error :%s", err)
		}

		f.current.fp = nil

		//当前需要轮替的文件就是 f.fname，这个是不会变的， 将它重命名为date+filename的形式，
		// 如果今天轮替了多次，那就是 laravel.log.2018-09-09.001，laravel.log.2018-09-09.002
		fname := formatFname(f.lastRotateFormat, f.lastSplitNum)
		oldpath := filepath.Join(f.dir, f.fname)
		newpath := filepath.Join(f.dir, fname)

		if err = os.Rename(oldpath, newpath); err != nil {
			f.stdlog.Printf("rename file %s to %s error: %s", oldpath, newpath, err)
			return
		}
		//再将刚刚新生成的日志存档挂在到list链表的尾部
		f.filesList.PushBack(rotateItem{fname: fname})

		//如果当前时间不是最后的轮替日期，代表那就开始新一轮的计数
		if nowDate != f.lastRotateFormat {
			f.lastRotateFormat = nowDate //重新设置日期
			f.lastSplitNum = 0           //当天已进行0次轮替
		} else {
			f.lastSplitNum++ //当天多次轮替，累加次数
		}

		//原本的文件已经被重命名为归档了，这里重新创建一个日志文件，并设置为current，以后就对着这个文件写
		f.current, err = newWrapFile(filepath.Join(f.dir, f.fname))
		if err != nil {
			f.stdlog.Printf("create log file error: %s", err)
		}

	}

}

//直接同步写入磁盘，在轮替进行的时候不能调用该方法，否则会报错
func (f *FileWriter) write(p []byte) error {
	f.writeLock.Lock()
	defer f.writeLock.Unlock()
	// f.current may be nil, if newWrapFile return err in checkRotate, redirect log to stderr
	if f.current == nil || f.current.fp == nil {
		f.stdlog.Printf("can't write log to file, please check stderr log for detail")
		f.stdlog.Printf("%s", p)
	}

	_, err := f.current.write(p)
	return err
}

//归还buffer到池中
func (f *FileWriter) putBuf(buf *bytes.Buffer) {
	buf.Reset()
	f.pool.Put(buf)
}

//从池中取出buff
func (f *FileWriter) getBuf() *bytes.Buffer {
	return f.pool.Get().(*bytes.Buffer)
}
