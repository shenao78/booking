package booking

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	sync.RWMutex
	users     []*User
	interval  int64
	anchorage *Anchorage
}

func (c *Client) AddUser(user *User) {
	c.users = append(c.users, user)
	sessionId, err := GetSessionID(user.token)
	if err != nil {
		panic(err.Error())
	}
	user.setSessionId(sessionId)
	go user.refreshSessionId()
}

func (c *Client) SetFormInfo() {
	retryIfErr(func() error {
		fmt.Printf("\n请输入锚地预约ID：")
		var anchorId string
		_, _ = fmt.Scanf("%s", &anchorId)
		if len(anchorId) == 0 {
			return errors.New("无效的锚地预约ID！")
		}
		user := c.users[0]
		anchorage, err := GetAnchorage(anchorId, user.token, user.getSessionId())
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
		fmt.Printf("\n设置提交时间间隔 （单位：秒）：")
		var inputSec int64
		_, _ = fmt.Scanf("%d", &inputSec)
		if inputSec == 0 {
			return errors.New("无效的时间！")
		}
		c.interval = inputSec * 1000
		return nil
	})
}

func (c *Client) Submit() {
	anchorage := c.anchorage
	anchorage.normalize()

	fmt.Println("开始提交锚地预约！！！")

	seq := 0
	cnt := len(c.users)
	avgSubmitInterval := c.interval / int64(cnt)
	var prevSubmit int64
	for {
		i := seq % cnt
		user := c.users[i]

		selfInterval := now().UnixMilli() - user.lastSubmitTime
		if selfInterval <= c.interval {
			time.Sleep(time.Duration(c.interval-selfInterval) * time.Millisecond)
		}
		userInterval := now().UnixMilli() - prevSubmit
		if userInterval <= avgSubmitInterval {
			time.Sleep(time.Duration(avgSubmitInterval-userInterval) * time.Millisecond)
		}

		resp, err := Submit(anchorage, user.token, user.getSessionId())
		user.lastSubmitTime = now().UnixMilli()
		prevSubmit = user.lastSubmitTime

		nowStr := now().Format("15:04:05")
		if err != nil {
			fmt.Printf("[%s][%s] %s", nowStr, user.userName, err.Error())
			seq++
			continue
		}
		if resp.Status != 200 {
			fmt.Printf("[%s][%s] 提交锚地预约信息失败!\n", user.userName, nowStr)
		} else if resp.Code != "10000" {
			fmt.Printf("[%s][%s] %s\n", nowStr, user.userName, resp.Message)
		} else {
			fmt.Printf("[%s][%s] 恭喜！预约锚地成功！\n", nowStr, user.userName)
			return
		}
		seq++
	}
}

func now() time.Time {
	return time.Now()
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

type User struct {
	sync.RWMutex
	userName       string
	token          string
	sessionId      string
	lastSubmitTime int64
}

func (u *User) setSessionId(sessionId string) {
	u.Lock()
	defer u.Unlock()
	u.sessionId = sessionId
}

func (u *User) getSessionId() string {
	u.RLock()
	defer u.RUnlock()
	return u.sessionId
}

func (u *User) refreshSessionId() {
	ticker := time.NewTicker(time.Minute)
	for ; true; <-ticker.C {
		sessionId, err := GetSessionID(u.token)
		if err == nil && len(sessionId) != 0 {
			u.setSessionId(sessionId)
		}
	}
}
