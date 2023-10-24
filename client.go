package booking

import (
	"encoding/base64"
	"fmt"
	"strings"
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
	anchorage   *Anchorage
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
			p, _ := term.ReadPassword(int(syscall.Stdin))
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
		fmt.Printf("请输入锚地预约ID：")
		var anchorId string
		_, _ = fmt.Scanf("\n%s", &anchorId)
		if len(anchorId) == 0 {
			return errors.New("无效的锚地预约ID！")
		}
		anchorage, err := GetAnchorage(anchorId, c.token, c.getSessionId())
		if err != nil {
			return err
		}
		c.anchorage = anchorage
		fmt.Println("-------------------------------------")
		fmt.Printf(
			"报告单位：%s\n船名：%s\n抛锚时间：%s\n离锚时间：%s\n",
			anchorage.ApplyObject,
			anchorage.ShipNameCh,
			anchorage.ArrangeAnchorTime,
			anchorage.ArrangeMoveAnchorTime,
		)
		fmt.Println("-------------------------------------")
		return nil
	})

	retryIfErr(func() error {
		fmt.Printf("设置预约时间（格式：12:00）：")
		var inputTime string
		_, _ = fmt.Scanf("\n%s", &inputTime)
		if len(inputTime) == 0 {
			return errors.New("无效的时间！")
		}
		bookingTime, err := parseTime(inputTime)
		if err != nil {
			return errors.New("时间格式错误，请按照格式：12:00")
		}
		if bookingTime.Before(time.Now()) {
			return errors.New("预约时间必须大于当前时间！")
		}
		c.bookingTime = bookingTime
		fmt.Println("预约时间设置成功：", bookingTime.String())
		return nil
	})
	go c.countdown()
}

func (c *Client) Submit() {
	anchorage := c.anchorage
	anchorage.IsSubmit = 1
	// TODO fetch from getAnchorGroundList
	anchorage.IsAnchGroundLimit = "1"
	anchorage.DownUploadfileList = []File{}
	if len(anchorage.FileList) != 0 {
		anchorage.DownUploadfileList = anchorage.FileList
	}
	anchorage.StopReasonList = []interface{}{}
	if anchorage.StopReason != nil {
		reasons := strings.Split(anchorage.StopReason.(string), ",")
		for _, r := range reasons {
			anchorage.StopReasonList = append(anchorage.StopReasonList, r)
		}
	}
	anchorage.DownUploadfileList = anchorage.FileList
	duration := c.bookingTime.Sub(time.Now())
	time.Sleep(duration)
	fmt.Println("开始提交锚地预约！！！")

	respCh := make(chan *CommonResp)
	endCh := make(chan interface{})
	c.parallelSubmit(anchorage, respCh, endCh, 10)
	go func() {
		time.Sleep(15 * time.Second)
		close(endCh)
	}()
	for {
		select {
		case resp := <-respCh:
			if resp.Status != 200 {
				fmt.Println("提交锚地预约信息失败!")
			} else if resp.Code != "10000" {
				now := time.Now().Format("15:04:05")
				fmt.Printf("[%s] %s\n", now, resp.Message)
			} else {
				close(endCh)
				fmt.Println("恭喜！预约锚地成功！")
				return
			}
		case <-endCh:
			fmt.Println("15秒内未能预约成功，程序停止提交")
			return
		}
	}
}

func (c *Client) parallelSubmit(anchorage *Anchorage, respCh chan *CommonResp, endCh chan interface{}, numProcessor int) {
	for i := 0; i < numProcessor; i++ {
		go func() {
			for {
				resp, err := Submit(anchorage, c.token, c.getSessionId())
				if err != nil {
					fmt.Println(err.Error())
				}
				select {
				case <-endCh:
					return
				default:
					if resp != nil {
						respCh <- resp
					}
				}
			}
		}()
	}
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
		duration := c.bookingTime.Sub(now)
		if int(duration.Seconds()) == 0 {
			fmt.Println()
			return
		}
		fmt.Printf("\r距离预约开始还剩：%s", fmtDuration(duration))
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
	date := time.Now().Format("2006-01-02")
	return time.ParseInLocation("2006-01-02 15:04", date+" "+t, loc)
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
