package main
import (
	"fmt"
	"os"
  "log"
  "net/http"
  "net/url"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/cache"
	"github.com/silenceper/wechat/message"
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
  wxState := os.Getenv("WECHAT_STATE")
  // 服务器的一些配置
  host := os.Getenv("HOST")
  wc := wechat.NewWechat(wcConfig)
  r := gin.Default()

  // 认证接入
  r.GET("/wechat", func (c *gin.Context) {
    // 传入request和responseWriter
    server := wc.GetServer(c.Request, c.Writer)
    //设置接收消息的处理方法
    server.SetMessageHandler(func(msg message.MixMessage) *message.Reply {
      //回复消息：演示回复用户发送的消息
      text := message.NewText(msg.Content)
      return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
    })

    //处理消息接收以及回复
    err := server.Serve()
    if err != nil {
      fmt.Println(err)
      return
    }
    //发送回复的消息
    server.Send()
  })

  // 获取code的中转页面
  r.GET("/auth", func (c *gin.Context) {
    callback := c.Query("callback")
    // 判断是否是正确的url格式
    _, err := url.ParseRequestURI(callback)
    if (err != nil) {
      c.Status(http.StatusBadRequest)
    } else {
      oauth := wc.GetOauth()
      err := oauth.Redirect(c.Writer, c.Request,
        host + "/callback?url=" + url.QueryEscape(callback),
        "snsapi_userinfo", wxState)
      if err != nil {
        fmt.Println(err)
      }
    }
  })

  // code返回后默认回调页面
  r.GET("/callback", func (c *gin.Context) {
    state := c.Query("state")
    if (state == wxState) {
      code := c.Query("code")
      redirect := c.Query("url")
      escRedirect, err := url.QueryUnescape(redirect)
      if err != nil {
        c.String(http.StatusBadRequest, "参数错误")
      } else {
        c.Redirect(http.StatusMovedPermanently, escRedirect + "?code=" + code)
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
