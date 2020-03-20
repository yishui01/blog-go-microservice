package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"net/http/httputil"
	"os"
	"runtime"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			var rawReq []byte
			if err := recover(); err != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				if c.Request != nil {
					rawReq, _ = httputil.DumpRequest(c.Request, false)
					pl := fmt.Sprintf("http call panic: %s\n%v\n%s\n", string(rawReq), err, buf)
					fmt.Fprintf(os.Stderr, pl)
					log.ZapWithContext(ctx).Error(pl)
					c.AbortWithStatus(500)
				}
			}
		}()
		c.Next()
	}
}
