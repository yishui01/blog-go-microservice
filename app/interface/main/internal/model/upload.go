package model

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"mime/multipart"
	"path"
	"strings"
	"time"
)

type Upload struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Tag       string `json:"tag"`
	Uploadkey string `json:"uploadkey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

//后台管理列表分页response
type BackListUpload struct {
	Total    int
	PageNum  uint
	PageSize uint
	Lists    []*Upload
}

//后台管理user filter
type BackUploadQuery struct {
	PageNum  uint   `form:"pageNum" validate:"required,numeric,min=1,max=200000"`
	PageSize uint   `form:"pageSize" validate:"required,numeric,min=1,max=200000000"`
	Tag      string `form:"tag"`
	Name     string `form:"name"`
}

type Qiniu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	UrlHost   string
}

type UploadManager struct {
	qiniu *Qiniu
}

func NewUploadManager(qiniu *Qiniu) *UploadManager {
	return &UploadManager{qiniu: qiniu}
}

func (q *UploadManager) UploadFile(c context.Context, fileHeader *multipart.FileHeader, prefix string) (err error, key string, url string) {
	putPolicy := storage.PutPolicy{
		Scope: q.qiniu.Bucket,
	}
	mac := qbox.NewMac(q.qiniu.AccessKey, q.qiniu.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	resumeUploader := storage.NewResumeUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.RputExtra{}
	f, err := fileHeader.Open()
	if err != nil {
		return errors.Wrap(err, "打开上传文件出错"), "", ""
	}
	prefix = strings.Trim(prefix, "/")
	newFileName := fmt.Sprintf("%s/%s%s", prefix, utils.Md5ByTime(fileHeader.Filename), path.Ext(fileHeader.Filename))
	err = resumeUploader.Put(context.Background(), &ret, upToken, newFileName, f, fileHeader.Size, &putExtra)
	if err != nil {
		return errors.Wrap(err, "上传出错"), "", ""
	}
	return nil, ret.Key, strings.Trim(q.qiniu.UrlHost, "/") + "/" + ret.Key
}

func (q *UploadManager) DeleteFile(c context.Context, key string) error {
	mac := qbox.NewMac(q.qiniu.AccessKey, q.qiniu.SecretKey)
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)
	err := bucketManager.Delete(q.qiniu.Bucket, key)
	return errors.WithStack(err)
}
