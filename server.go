package main
import (
	"fmt"
	"os"
  "log"
  "net/http"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/cache"
)

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
    os.Exit(-1)
  }
  // redis缓存配置
  redisCache := cache.NewRedis(&cache.RedisOpts{
    Host: os.Getenv("REDIS_HOST"),
    Password: os.Getenv("REDIS_PASSWORD"),
  })
  // 微信公众号配置
  wcConfig := &wechat.Config{
    AppID:          os.Getenv("APP_ID"),
    AppSecret:      os.Getenv("APP_SECRET"),
    Token:          os.Getenv("TOKEN"),
    EncodingAESKey: os.Getenv("ENCODING_AES_KEY"),//消息加解密时用到
    Cache:          redisCache,
  }
  wc := wechat.NewWechat(wcConfig)
	r := gin.Default()
  r.GET("wechat", func (c *gin.Context) {
    callback := c.Query("callback")
    if (callback == "") {
      c.Status(http.StatusBadRequest)
    } else {
      oauth := wc.GetOauth()
      err := oauth.Redirect(c.Writer, c.Request,
        callback,
        "snsapi_userinfo", "erguotou")
      if err != nil {
        fmt.Println(err)
      }
    }
  })

  r.GET("/code2user", func(c *gin.Context) {
    oauth := wc.GetOauth()
    code := c.Query("code")
    resToken, err := oauth.GetUserAccessToken(code)
    if err != nil {
      fmt.Println(err)
      return
    }
    userInfo, err := oauth.GetUserInfo(resToken.AccessToken, resToken.OpenID)
    if err != nil {
      fmt.Println(err)
      return
    }
    fmt.Println(userInfo)
  })
	r.Run(":9800") // listen and serve on 0.0.0.0:8080
}
