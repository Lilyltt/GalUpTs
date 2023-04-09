package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

const (
	bufferSize = 5000 // 定义读取文件时的缓冲区大小
)

func main() {
	//读取apikey
	file2, err := os.Open("apikey.txt")
	defer file2.Close()
	sc2 := bufio.NewScanner(file2)
	var apikey string
	for sc2.Scan() {
		apikey += sc2.Text()
	}
	if err != nil || apikey == "" {
		fmt.Println("[ERROR]请将包含apikey的apikey.txt放到程序根目录下!")
		os.Exit(0)
	}
	//读取翻译头
	file3, err := os.Open("head.txt")
	defer file3.Close()
	sc3 := bufio.NewScanner(file3)
	var head string
	for sc3.Scan() {
		head += sc3.Text()
	}
	if err != nil || head == "" {
		fmt.Println("[WARN]未自定义翻译头,已设置默认值为 翻译下面这段话为简体中文,保留原格式 ,如需修改请将包含翻译头的head.txt放到程序根目录下!")
		head = "翻译下面这段话为简体中文,保留原格式 "
	}
	// 创建翻译文件
	fileresult, err3 := os.Create("translation.txt")
	if err3 != nil {
		panic(err3)
	}
	defer fileresult.Close()
	writer := bufio.NewWriter(fileresult)
	file, err := os.Open("input.txt")
	if err != nil {
		fmt.Println("[ERROR]请将包含需要翻译内容的input.txt放到程序根目录下!")
		os.Exit(0)
	}
	defer file.Close()
	fmt.Println("[INFO]已读取到apikey:" + apikey + "\n[INFO]当前使用模型:GPT3Dot5Turbo\n[INFO]当前翻译头:" + head + "\n[INFO]按下enter开始翻译")
	fmt.Scanln()
	fmt.Println("[INFO]已开始翻译,请耐心等待,翻译结果将实时输出至控制台和translate.txt")
	//翻译
	var count = 1
	var buffer strings.Builder              // 用于存放读取到的字符
	var leftBraceCount, rightBraceCount int // 记录当前读到的{和},的数量
	sc1 := bufio.NewScanner(file)
	for sc1.Scan() {
		// 读取一行并存放到buffer中
		buffer.WriteString(sc1.Text() + "\n")
		// 在当前行中查找{和},
		leftBraceCount += strings.Count(sc1.Text(), "{")
		rightBraceCount += strings.Count(sc1.Text(), "},")

		// 判断当前{和}的数量是否相等，并且是否达到20个
		if leftBraceCount == rightBraceCount && leftBraceCount == 20 {
			// 输出读取到的内容
			source := head + buffer.String() + "\n\n"
			count++
			// 写入新文件
			_, err = writer.WriteString(ChatGPTTranslation(source, apikey))
			if err != nil {
				panic(err)
			}
			err = writer.Flush()
			if err != nil {
				panic(err)
			}
			fmt.Print(ChatGPTTranslation(source, apikey) + ",\n")
			// 清空buffer
			buffer.Reset()
			// 重置{和}的数量
			leftBraceCount = 0
			rightBraceCount = 0
			//提示
			fmt.Printf("[INFO]已生成第%d段,将进行下一段\n", count-1)
		}
	}
	if err := sc1.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Reading standard input:", err)
	}
	fmt.Printf("[INFO]翻译完成,共翻译%d段,请检查末尾是否缺失,格式是否错误\n", count)
}
func ChatGPTTranslation(message string, apikey string) (result string) {
	client := openai.NewClient(apikey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
	}
	return resp.Choices[0].Message.Content
}
