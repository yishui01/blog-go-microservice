package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"strconv"
	"strings"
)

const Prefix = "testupload"

//文件列表
func (s *Service) UploadList(c *khttp.Context) {
	query := new(model.BackUploadQuery)
	if err := c.MustBind(query); err != nil {
		return
	}
	uploads, total,err := s.d.BackUploadList(c, query)
	if err != nil {
		log.SugarWithContext(c).Error("err")
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	datas := new(model.BackListUpload)
	datas.Lists = uploads
	datas.Total = total
	datas.PageNum = query.PageNum
	datas.PageSize = query.PageSize
	c.JSON(datas, nil)
}

//上传文件
func (s *Service) Upload(c *khttp.Context) {
	_, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "上传失败："+err.Error()))
		c.Abort()
		return
	}
	err, key, url := s.UploadManager.UploadFile(c, fileHeader, Prefix)
	if err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "server上传失败"+err.Error()))
		c.Abort()
		return
	}

	if err := s.d.UploadCreate(c, fileHeader.Filename, key, url); err != nil {
		log.SugarWithContext(c).Error("s.d.Upload Err:(%#+v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "server Err"))
		c.Abort()
		return
	}

	c.JSON(map[string]string{
		"key": key,
		"url": url,
	}, err)
}

//删除文件
func (s *Service) UploadDelete(c *khttp.Context) {
	idStr := c.Request.Form.Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "id不合法:"+idStr))
		c.Abort()
		return
	}
	//查找本地数据库记录
	file, err := s.d.UploadFind(c, id)
	if err != nil {
		if err == ecode.NothingFound {
			c.JSON(nil, ecode.Error(ecode.RequestErr, "文件不存在"))
		} else {
			c.JSON(nil, ecode.ServerErr)
			log.SugarWithContext(c).Error("s.d.UploadFind Err:(%#+v)", err)
		}
		c.Abort()
		return
	}
	//先删远端
	if err := s.UploadManager.DeleteFile(c, file.Uploadkey); err != nil && !strings.Contains(err.Error(), "no such file") {
		log.SugarWithContext(c).Error("s.d.UploadFind Err:(%#+v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "删除失败"))
		c.Abort()
		return
	}
	//再删除本地数据库记录
	if err := s.d.UploadDelete(c, id); err != nil {
		log.SugarWithContext(c).Error("s.d.UploadDelete Err:(%#+v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "本地记录删除失败"))
		c.Abort()
		return
	}

	c.JSON(nil, nil)

}
