package booking

import (
	"encoding/base64"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/term"
)

type Client struct {
	sync.RWMutex
	token       string
	sessionId   string
	bookingTime time.Time
}

func (c *Client) Login() {
	var token string
	retryIfErr(func() error {
		var err error
		var userName, passwd string
		retryIfErr(func() error {
			fmt.Printf("用户名：")
			n, err := fmt.Scanf("%s", &userName)
			if n == 0 {
				return errors.New("无效的用户名！")
			}
			return err
		})

		retryIfErr(func() error {
			fmt.Printf("密码：")
			p, err := term.ReadPassword(int(syscall.Stdin))
			if len(p) == 0 {
				return errors.New("无效的密码！")
			}
			p, err = AESEncrypt(p, []byte("zjmsa_allhigh@12"))
			if err != nil {
				return errors.Wrap(err, "无效的密码！")
			}
			passwd = base64.StdEncoding.EncodeToString(p)
			return nil
		})

		var code, key int64
		retryIfErr(func() error {
			key = GetCode()
			fmt.Printf("\n验证码已保存至本地，请根据图片输入验证码：")
			_, _ = fmt.Scanf("%d", &code)
			if code == 0 {
				return errors.New("无效的验证码！")
			}
			return VerifyCode(code, key)
		})

		token, err = Login(code, key, userName, passwd)
		return err
	})
	sessionId, err := GetSessionID(token)
	if err != nil {
		panic(err.Error())
	}
	c.setSessionId(sessionId)
	c.token = token
	go c.refreshSessionId()
}

func (c *Client) SetFormInfo() {
	retryIfErr(func() error {
		fmt.Printf("设置抢票时间（格式：2023-10-22 12:00）：")
		var inputDate, inputTime string
		_, _ = fmt.Scanf("%s %s", &inputDate, &inputTime)
		if len(inputTime) == 0 || len(inputTime) == 0 {
			return errors.New("无效的时间！")
		}
		bookingTime, err := parseTime(inputDate + " " + inputTime)
		if err != nil {
			return errors.New("时间格式错误，请按照格式：2023-10-22 12:00")
		}
		c.bookingTime = bookingTime
		fmt.Println("抢票时间设置成功：", bookingTime.String())
		return nil
	})
	go c.countdown()
}

func (c *Client) setSessionId(sessionId string) {
	c.Lock()
	defer c.Unlock()
	c.sessionId = sessionId
}

func (c *Client) getSessionId() string {
	c.RLock()
	defer c.RUnlock()
	return c.sessionId
}

func (c *Client) refreshSessionId() {
	ticker := time.NewTicker(time.Minute)
	for ; true; <-ticker.C {
		sessionId, err := GetSessionID(c.token)
		if err == nil && len(sessionId) != 0 {
			c.setSessionId(sessionId)
		}
	}
}

func (c *Client) countdown() {
	ticker := time.NewTicker(time.Second)
	for ; true; <-ticker.C {
		now := time.Now()
		if now.Equal(c.bookingTime) || now.After(c.bookingTime) {
			return
		}
		duration := c.bookingTime.Sub(now)
		fmt.Printf("\r距离抢票开始还剩：%s", fmtDuration(duration))
	}
}

func fmtDuration(d time.Duration) string {
	var result string
	hour := int64(d.Hours())
	if hour != 0 {
		result += fmt.Sprintf("%d小时", hour)
	}
	minutes := int64(d.Minutes()) % 60
	if minutes != 0 {
		result += fmt.Sprintf("%d分", minutes)
	}
	sec := int64(d.Seconds()) % 60
	result += fmt.Sprintf("%d秒", sec)
	return result
}

func parseTime(t string) (time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.ParseInLocation("2006-01-02 15:04", t, loc)
}

func retryIfErr(fun func() error) {
	for {
		if err := fun(); err != nil {
			fmt.Println(err.Error())
		} else {
			break
		}
	}
}
