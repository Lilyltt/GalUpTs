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
	fmt.Println("[INFO]开始翻译,如果输出格式不是name|message请用ctrl+c停止程序并重新运行")
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
	var count int
	for _, inputRecord := range inputRecords {
		outputString := ""
		if inputRecord.Name != "" {
			outputString += inputRecord.Name + "|"
		}
		outputString += inputRecord.Message
		//翻译
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: translatehead,
		})
		fmt.Println("[INFO]原文: " + outputString)
	tsstart:
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
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: translatehead,
		})
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: translatehead,
		})
		//处理
		content = strings.Replace(content, "：", "|", -1)
		content = strings.Replace(content, ":", "|", -1)
		fmt.Println("[INFO]译文: " + content)
		//输出
		if result != "" {
			result = result + content + "\n"
		}
		//清记录,补预设
		count++
		if count >= 20 {
			messages = messages[3:]
		}
		if count%15 == 0 {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: translatehead,
			})
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: translatehead,
			})
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: translatehead,
			})
		}
	}
	fmt.Println("[INFO]翻译完毕,正在创建新文件")
	// 构建JSON数组
	// 解析结果并输出到json
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

	err = ioutil.WriteFile(outputdir+filename, outputBytes, 0644)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}

	fmt.Println("[INFO]已写入" + filename)
}
