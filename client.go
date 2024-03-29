package booking

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	sync.RWMutex
	token     string
	sessionId string
	duration  time.Duration
	anchorage *Anchorage
}

func (c *Client) Login() {
	var token string
	retryIfErr(func() error {
		var err error
		var userName, passwd string

		userName, passwd, err = readUserPass()
		if err != nil {
			fmt.Println("读取用户密码文件失败，请手动输入")
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
			writeUserPass(userName, passwd)
		}

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

func readUserPass() (string, string, error) {
	exec, _ := os.Executable()
	path := filepath.Join(filepath.Dir(exec), "auth.txt")
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", "", err
	}
	auth := strings.Split(string(data), " ")
	if len(auth) != 2 {
		return "", "", errors.New("invalid auth")
	}
	return auth[0], auth[1], nil
}

func writeUserPass(user, pass string) {
	exec, _ := os.Executable()
	path := filepath.Join(filepath.Dir(exec), "auth.txt")
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s %s", user, pass)))
	if _, err := file.Write([]byte(auth)); err != nil {
		panic(err)
	}
}

func (c *Client) SetFormInfo() {
	retryIfErr(func() error {
		fmt.Printf("\n请输入锚地预约ID：")
		var anchorId string
		_, _ = fmt.Scanf("%s", &anchorId)
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
		fmt.Printf("\n设置持续时间间隔 （单位：分钟）：")
		var inputMinutes uint64
		_, _ = fmt.Scanf("%d", &inputMinutes)
		if inputMinutes == 0 {
			return errors.New("无效的时间！")
		}
		c.duration = time.Minute * time.Duration(inputMinutes)
		return nil
	})
}

func (c *Client) Submit() {
	anchorage := c.anchorage
	anchorage.normalize()

	fmt.Println("开始提交锚地预约！！！")

	respCh := make(chan *CommonResp)
	endCh := make(chan interface{})
	c.parallelSubmit(anchorage, respCh, endCh, 20)
	go func() {
		time.Sleep(c.duration)
		close(endCh)
	}()
	for {
		select {
		case resp := <-respCh:
			now := time.Now().Format("15:04:05")
			if resp.Status != 200 {
				fmt.Printf("[%s] 提交锚地预约信息失败!\n", now)
			} else if resp.Code != "10000" {
				fmt.Printf("[%s] %s\n", now, resp.Message)
			} else {
				close(endCh)
				fmt.Printf("[%s] 恭喜！预约锚地成功！\n", now)
				return
			}
		case <-endCh:
			fmt.Printf("%d分钟未能预约成功，程序停止提交\n", int64(c.duration.Minutes()))
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

func retryIfErr(fun func() error) {
	for {
		if err := fun(); err != nil {
			fmt.Println(err.Error())
		} else {
			break
		}
	}
}
