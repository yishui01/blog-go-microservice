package filewriter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const logdir = "testlog"

func touch(dir, name string) {
	os.MkdirAll(dir, 0755)
	fp, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	fp.Close()
}

func TestMain(m *testing.M) {
	ret := m.Run()
	os.RemoveAll(logdir)
	os.Exit(ret)
}

func TestParseRotate(t *testing.T) {
	touch := func(dir, name string) {
		os.MkdirAll(dir, 0755)
		fp, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		fp.Close()
	}
	dir := filepath.Join(logdir, "test-parse-rotate")
	names := []string{"info.log.2018-11-11", "info.log.2018-11-11.001", "info.log.2018-11-11.002", "info.log." + time.Now().Format("2006-01-02") + ".005"}
	for _, name := range names {
		touch(dir, name)
	}
	l, err := parseRotateItem(dir, "info.log", "2006-02-02")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(names), l.Len())

	rt := l.Front().Value.(rotateItem)

	assert.Equal(t, 5, rt.rotateNum)
}

func TestRotateExists(t *testing.T) {
	subDir := "test-rotate-exists"
	dir := filepath.Join(logdir, subDir)
	names := []string{"info.log." + time.Now().Format("2006-01-02") + ".005"}
	for _, name := range names {
		touch(dir, name)
	}
	fw, err := New(logdir+"/"+subDir+"/info.log",
		MaxSize(1024*1024),
		RotateInterval(time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			_, err = fw.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	fw.Close()
	fis, err := ioutil.ReadDir(logdir + "/" + subDir)
	if err != nil {
		t.Fatal(err)
	}
	var fnams []string
	for _, fi := range fis {
		fnams = append(fnams, fi.Name())
	}
	assert.Contains(t, fnams, "info.log."+time.Now().Format("2006-01-02")+".006")
}

func TestMaxFile(t *testing.T) {
	fw, err := New(logdir+"/test-maxfile/info.log",
		MaxSize(1024*1024),
		MaxFile(1),
		RotateInterval(1*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			_, err = fw.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	fw.Close()
	fis, err := ioutil.ReadDir(logdir + "/test-maxfile")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, len(fis) == 2, fmt.Sprintf("expect 2 file get %d", len(fis)))
}

func TestMaxFile2(t *testing.T) {
	files := []string{
		"info.log.2018-12-01",
		"info.log.2018-12-02",
		"info.log.2018-12-03",
		"info.log.2018-12-04",
		"info.log.2018-12-05",
		"info.log.2018-12-05.001",
	}
	for _, file := range files {
		touch(logdir+"/test-maxfile2", file)
	}
	fw, err := New(logdir+"/test-maxfile2/info.log",
		MaxSize(1024*1024),
		MaxFile(3),
		RotateInterval(1*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			_, err = fw.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	fw.Close()
	fis, err := ioutil.ReadDir(logdir + "/test-maxfile2")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, len(fis) == 4, fmt.Sprintf("expect 4 file get %d", len(fis)))
}

func TestFileWriter(t *testing.T) {
	fw, err := New(logdir + "/info.log")
	if err != nil {
		t.Fatal(err)
	}
	defer fw.Close()
	_, err = fw.Write([]byte("Hello World!\n"))
	if err != nil {
		t.Error(err)
	}
}

//BenchmarkFileWriter/syscall               357264              2853 ns/op
//BenchmarkFileWriter/frame-sync            342744              2980 ns/op
//BenchmarkFileWriter/frame-async          1001078              1039 ns/op

//go test -bench BenchmarkFileWriter  -run =^$ -cpu 1,2,4,8
func BenchmarkFileWriter(b *testing.B) {

	sync, err := New(logdir+"/bench/sync.log",
		ChanSize(10240),
		WriteTimeout(time.Second),
		RotateInterval(10*time.Millisecond),
		MaxSize(1024*1024*8), /*80MB*/
	)
	if err != nil {
		b.Fatal(err)
	}

	async, err := New(logdir+"/bench/async.log",
		ChanSize(10240),
		UserBuffer(true),
		WriteTimeout(time.Second),
		RotateInterval(10*time.Millisecond),
		MaxSize(1024*1024*8), /*80MB*/
	)
	if err != nil {
		b.Fatal(err)
	}

	f, err := os.OpenFile(logdir+"/bench/direct.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("err", err)
	}

	msg := "Hello World!\n"
	b.Run("syscall", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = f.Write([]byte(fmt.Sprintf("url is %s", msg)))
			if err != nil {
				b.Error(err)
			}
		}
	})
	f.Close()

	b.Run("frame-sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = sync.Write([]byte(fmt.Sprintf("url is %s", msg)))
			if err != nil {
				b.Error(err)
			}
		}
	})
	sync.Close()

	b.Run("frame-async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = async.Write([]byte(fmt.Sprintf("url is %s", msg)))
			if err != nil {
				b.Error(err)
			}
		}
	})
	async.Close()

	time.Sleep(time.Second * 2)
	fmt.Println(os.RemoveAll(logdir))

}
