package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func TsTxt(client *openai.Client, filename string, inputdir string, outputdir string, translatehead string) {
	messages := make([]openai.ChatCompletionMessage, 0)
	//读取文件
	fmt.Println("[INFO]开始翻译,如果输出格式不是name|message请用ctrl+c停止程序并重新运行")
	// 读取 filename 文件
	file, err := os.Open(inputdir + filename)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: translatehead,
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: translatehead,
	})
	var count int
	var result string
	for scanner.Scan() {
		//翻译
		//加预设
		fmt.Println("[INFO]原文: " + scanner.Text())
	tsstart:
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: scanner.Text(),
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
			time.Sleep(5 * time.Second)
			goto tsstart
		}

		content := resp.Choices[0].Message.Content
		content = strings.Replace(content, "\n", "", -1)
		//输出
		fmt.Println("[INFO]译文: " + content)
		result = result + content + "\n"
		//清记录,补预设
		count++
		if count >= 20 {
			messages = messages[2:]
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
		}
	}
	fmt.Println("[INFO]翻译完毕,正在创建新文件")
	// 输出到 Translate.txt
	err = ioutil.WriteFile(outputdir+filename, []byte(result), 0644)
	if err != nil {
		fmt.Print("[ERROR]")
		fmt.Print(err)
		fmt.Scanln()
		return
	}

	fmt.Println("[INFO]已写入" + filename)
}
