package booking

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (c *Client) LoginAll() {
	auths, err := readUserPass()
	if err != nil {
		fmt.Println("读取用户密码文件失败：", err)
		os.Exit(-1)
	}
	if len(auths) == 0 {
		panic("请至少指定一个用户！")
	}
	for _, auth := range auths {
		c.login(auth)
	}
}

func (c *Client) login(auth *auth) {
	p := AESEncrypt([]byte(auth.pass), []byte("zjmsa_allhigh@12"))
	passwd := base64.StdEncoding.EncodeToString(p)

	var token string
	var err error
	retryIfErr(func() error {
		var code, key int64
		retryIfErr(func() error {
			key = GetCode()
			fmt.Printf("请根据图片输入验证码：")
			_, _ = fmt.Scanf("%d", &code)
			if code == 0 {
				return errors.New("无效的验证码！")
			}
			return VerifyCode(code, key)
		})

		token, err = Login(code, key, auth.user, passwd)
		return err
	})
	fmt.Printf("%s 登录成功\n", auth.user)
	c.AddUser(&User{
		userName: auth.user,
		token:    token,
	})
}

type auth struct {
	user string
	pass string
}

func readUserPass() ([]*auth, error) {
	exec, _ := os.Executable()
	path := filepath.Join(filepath.Dir(exec), "auth.txt")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var auths []*auth
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "")
		re := regexp.MustCompile("\\s+")
		data := re.Split(line, 2)
		if len(data) < 2 {
			continue
		}
		auths = append(auths, &auth{user: data[0], pass: data[1]})
	}
	return auths, nil
}
