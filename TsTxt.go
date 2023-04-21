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
	var result string
	var checkcount int
	var tempres string
	for scanner.Scan() {
		//翻译
		//加预设
	tsstart:
		fmt.Println("[INFO]原文: " + scanner.Text())
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: translatehead,
		})
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
		//仅标点符号校验检查
		if OnlyMarkCheck(scanner.Text()) {
			tempres = scanner.Text()
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
		tempres = content
		tempres = strings.Replace(tempres, "\n", "", -1)
		//输出
		fmt.Println("[INFO]译文: " + tempres)
		result = result + tempres + "\n"
		//清记录,补预设
		if len(messages) >= 60 {
			messages = messages[3:]
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
