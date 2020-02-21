package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"go.uber.org/zap"
	"time"
)

// Logger is logger  middleware
func Logger() gin.HandlerFunc {
	const noUser = "no_user"
	return func(c *gin.Context) {

		now := time.Now()
		ip := c.ClientIP()
		req := c.Request
		path := req.URL.Path
		params := req.Form
		var quota float64
		if deadline, ok := c.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		c.Next()

		err := c.Errors
		dt := time.Since(now)

		lf := log.ZapLogger.Info
		errmsg := ""
		isSlow := dt >= (time.Millisecond * 500)
		if isSlow {
			lf = log.ZapLogger.Warn
		}
		if len(err) > 0 {
			lf = log.ZapLogger.Error
		}
		errMsg := ""
		for _, v := range c.Errors {
			errMsg += v.Error() + "; "
		}
		lf("http finish log",
			zap.String("method", req.Method),
			zap.String("ip", ip),
			zap.String("path", path),
			zap.String("params", params.Encode()),
			zap.Int("errCount", len(c.Errors)),
			zap.String("errMsg", errMsg),
			zap.String("stack", fmt.Sprintf("%+v", err)),
			zap.String("err", errmsg),
			zap.Float64("timeout_quota", quota),
			zap.Float64("ts", dt.Seconds()),
			zap.String("source", "http-access-log"),
		)
	}
}
