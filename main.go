package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type InputMessage struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func main() {
	//输入apikey
	fmt.Println("[INFO]请输入OpenAi apikey:")
	var apikey string
	fmt.Scanln(&apikey)
	if apikey == "" {
		fmt.Println("[ERROR]必须输入apikey!程序退出")
		fmt.Scanln()
		os.Exit(0)
	}
	//配置代理
	fmt.Println("[INFO]请输入HTTP代理地址 http://localhost:端口号 ,留空则不设置代理")
	var proxy string
	fmt.Scanln(&proxy)
	var client *openai.Client
	if proxy != "" {
		client = SetProxy(apikey, proxy)
		fmt.Println("[INFO]代理设置成功,为" + proxy)
	} else {
		client = openai.NewClient(apikey)
		fmt.Println("[INFO]留空,跳过设置代理")
	}
	messages := make([]openai.ChatCompletionMessage, 0)
	//读取文件
	fmt.Println("[INFO]开始翻译,如果输出格式不是name|message请用ctrl+c停止程序并重新运行")
	// 读取 input.json 文件
	file, err := os.Open("input.json")
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}
	defer file.Close()
	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}
	var inputRecords []InputMessage
	err = json.Unmarshal([]byte(jsonData), &inputRecords)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}
	var result string
	var count int
	for _, inputRecord := range inputRecords {
		outputString := ""
		if inputRecord.Name != "" {
			outputString += inputRecord.Name + "|"
		}
		outputString += inputRecord.Message
		//翻译
		fmt.Println("[INFO]原文: " + outputString)
	tsstart:
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: "trasnlate next line to chinese:\n" + outputString,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("[ERROR]翻译出现错误,尝试重新翻译该句,如果持续出错请重启持续或检查代理设置: %v\n", err)
			goto tsstart
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		//处理
		content = strings.Replace(content, "：", "|", -1)
		content = strings.Replace(content, ":", "|", -1)
		fmt.Println("[INFO]译文: " + content)
		count++
		if count == 20 {
			count = 0
			messages = nil
			fmt.Println("[WARN]达到最大记忆量,已清除对话记录,请注意该句附近的翻译效果")
		}
		result = result + content + "\n"
	}
	fmt.Println("[INFO]翻译完毕,正在创建新文件")
	// 构建JSON数组
	// 解析结果并输出到 Translate.json
	var outputMessages []InputMessage
	for _, entry := range strings.Split(result, "\n") {
		parts := strings.SplitN(entry, "|", 2)
		if len(parts) > 1 {
			outputMessages = append(outputMessages, InputMessage{Name: parts[0], Message: parts[1]})
		} else {
			outputMessages = append(outputMessages, InputMessage{Message: entry})
		}
	}

	outputBytes, err := json.Marshal(outputMessages)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}

	err = ioutil.WriteFile("Translate.json", outputBytes, 0644)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}

	fmt.Println("[INFO]JSON格式化成功，已写入translate.json文件")
	fmt.Scanln()
}
func SetProxy(apikey string, proxy string) *openai.Client {
	config := openai.DefaultConfig(apikey)
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		os.Exit(0)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
	}
	client := openai.NewClientWithConfig(config)
	return client
}
