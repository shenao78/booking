package booking

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const ApiURL = "https://zkpt.zj.msa.gov.cn/out-uaa-api/api/v1/"

type CommonResp struct {
	Status  uint64 `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func GetCode() int64 {
	now := time.Now()
	ts := now.UnixMilli()
	key := now.UnixNano() / 100
	url := fmt.Sprintf("%susers/getCode?timestamp=%d&key=%d", ApiURL, ts, key)
	img, err := GetRaw(url)
	if err != nil {
		panic(errors.Wrap(err, "获取验证码"))
	}
	if err := os.WriteFile("code.jpg", img, 0644); err != nil {
		panic(errors.Wrap(err, "保存验证码"))
	}
	return key
}

type verifyCodeResp struct {
	CommonResp
	Data string `json:"data"`
}

func VerifyCode(code, key int64) error {
	url := fmt.Sprintf("%susers/validateCode?validateCode=%d&key=%d", ApiURL, code, key)
	resp := &verifyCodeResp{}
	if err := Get(url, resp); err != nil {
		return err
	}
	if resp.Status != 200 {
		return errors.New("verify验证码失败！")
	}
	if resp.Code != "10000" {
		return errors.New(resp.Message)
	}
	fmt.Println("验证成功")
	return nil
}

type loginResp struct {
	CommonResp
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type loginReq struct {
	LoginName    string `json:"loginName"`
	Password     string `json:"password"`
	ValidateCode string `json:"validateCode"`
	Key          string `json:"key"`
}

func Login(code, key int64, name, passwd string) (string, error) {
	url := fmt.Sprintf("%susers/login", ApiURL)
	req := &loginReq{
		LoginName:    name,
		Password:     passwd,
		ValidateCode: fmt.Sprintf("%d", code),
		Key:          fmt.Sprintf("%d", key),
	}
	payload, _ := json.Marshal(req)
	resp := &loginResp{}
	if err := Post(url, payload, resp); err != nil {
		return "", err
	}
	if resp.Status != 200 {
		return "", errors.New("登录失败！")
	}
	if resp.Code != "10000" {
		return "", errors.New(resp.Message)
	}
	fmt.Println("登录成功")
	return resp.Data.Token, nil
}

func GetSessionID(token string) (string, error) {
	url := "https://zkpt.zj.msa.gov.cn/trafficflow-api/api/v1/out/my-stuff-infos/unreadCount"
	resp := &CommonResp{}
	header, err := GetWithHeader(url, map[string]string{"Authorization": token}, resp)
	if err != nil {
		return "", err
	}
	if resp.Status != 200 {
		return "", errors.New("获取SessionId失败！")
	}
	if resp.Code != "10000" {
		return "", errors.New(resp.Message)
	}
	cookies := header.Get("Set-Cookie")
	for _, cookie := range strings.Split(cookies, ";") {
		cookie = strings.Trim(cookie, " ")
		if strings.HasPrefix(cookie, "JSESSIONID=") {
			return cookie, nil
		}
	}
	return "", errors.New("服务端未返回SessionId！")
}

type anchorageResp struct {
	CommonResp
	Data *Anchorage `json:"data"`
}

func GetAnchorage(anchorId, token, sessionId string) (*Anchorage, error) {
	url := fmt.Sprintf("https://zkpt.zj.msa.gov.cn/trafficflow-api/api/v1/out/apply-anchorages/load/%s", anchorId)
	resp := &anchorageResp{}
	if _, err := GetWithHeader(url, map[string]string{"Authorization": token, "Cookie": sessionId}, resp); err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, errors.New("获取锚地预约信息失败！")
	}
	if resp.Code != "10000" {
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

func Submit(anchorage *Anchorage, token, sessionId string) (*CommonResp, error) {
	url := "https://zkpt.zj.msa.gov.cn/trafficflow-api/api/v1/out/apply-anchorages/saveData"
	resp := &CommonResp{}
	payload, _ := json.Marshal(anchorage)
	if err := PostWithHeader(url, map[string]string{"Authorization": token, "Cookie": sessionId}, payload, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
