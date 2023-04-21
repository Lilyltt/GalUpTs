package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type InputMessage struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func TsJson(client *openai.Client, filename string, inputdir string, outputdir string, translatehead string) {
	//注册翻译体
	messages := make([]openai.ChatCompletionMessage, 0)
	//读取文件
	fmt.Println("[INFO]开始翻译,如果输出格式不是name:message请用ctrl+c停止程序并重新运行")
	// 读取 input.json 文件
	file, err := os.Open(inputdir + filename)
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
	var checkcount int
	var tempres string
	for _, inputRecord := range inputRecords {
		outputString := ""
		if inputRecord.Name != "" {
			outputString += inputRecord.Name + ":"
		}
		outputString += inputRecord.Message
		//翻译
	tsstart:
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: translatehead,
		})
		fmt.Println("[INFO]原文: " + outputString)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: outputString,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("[ERROR]翻译出现错误,尝试重新翻译该句,如果持续出错请重启程序或检查代理设置: %v\n", err)
			time.Sleep(5 * time.Second)
			goto tsstart
		}

		content := resp.Choices[0].Message.Content
		//仅标点符号校验检查
		if OnlyMarkCheck(inputRecord.Message) {
			if inputRecord.Name != "" {
				tempres += inputRecord.Name + ":"
			}
			tempres += inputRecord.Message
			fmt.Println("[WARN]原文仅有标点符号,跳过该句检查并保持原样输出")
		}
		//翻译校验检查
		if checkcount != 0 {
			if GptErrCheck(content) == false {
				checkcount = 0
			}
		}
		if GptErrCheck(content) {
			checkcount++
			if checkcount <= 5 {
				fmt.Println("[INFO]译文: " + content)
				messages = messages[:len(messages)-3]
				fmt.Printf("[ERROR]翻译校验检查出问题,正在进行第 %d 次重试修正:\n", checkcount)
				time.Sleep(5 * time.Second)
				goto tsstart
			} else {
				fmt.Printf("[ERROR]翻译校验检查出问题,已进行 %d 次修正依然出错,将跳过该句:\n", checkcount)
				checkcount = 0
			}
		}
		//处理
		tempres = content
		tempres = strings.Replace(tempres, "：", ":", -1)
		tempres = strings.Replace(tempres, "\n", "", -1)
		fmt.Println("[INFO]译文: " + tempres)
		//输出
		if tempres != "" {
			result = result + tempres + "\n"
		}
		//清记录,补预设
		if len(messages) >= 60 {
			messages = messages[3:]
		}
	}
	fmt.Println("[INFO]翻译完毕,正在创建新文件")
	// 构建JSON数组
	// 解析结果并输出到json
	var outputMessages []InputMessage
	for _, entry := range strings.Split(result, "\n") {
		parts := strings.SplitN(entry, ":", 2)
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

	err = ioutil.WriteFile(outputdir+filename, outputBytes, 0644)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}

	fmt.Println("[INFO]已写入" + filename)
}
