package httpv

import (
	"github.com/gin-gonic/gin"
	"github.com/camsiabor/qcom/qlog"
)
func Recovery(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return RecoveryWithWriter(f)
}

func RecoveryWithWriter(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				qlog.Log(qlog.ERROR, err);
				//httprequest, _ := httputil.DumpRequest(c.Request, false)
				//goErr := errors.Wrap(err, 3)
				//reset := string([]byte{27, 91, 48, 109})
				//logger.Printf("[Nice Recovery] panic recovered:\n\n%s%s\n\n%s%s", httprequest, goErr.Error(), goErr.Stack(), reset)
				f(c, err)
			}
		}()
		c.Next() // execute all the handlers
	}
}