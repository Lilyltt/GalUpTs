package main

import (
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	//获取apikey
	fmt.Println("=================================================\n[INFO]使用前须知:\n1.你需要来自OpenAi官方的apikey\n2.你需要一个非cn的网络环境或通过 http://localhost:端口号 设置代理\n3.你需要确认你待翻译文件的格式,当前支持 通过VNTEXT提取的KRKR-SCN-JSON文件 和 无格式一行一句的TXT文件\n4.把你所有需要翻译的文件选择一个目录放好,并记下地址\n5.出错请详细阅读报错+善用搜索引擎\n6.按下enter继续\n=================================================")
	fmt.Scanln()
	var apikey string //apikey
	fmt.Println("[INFO]请输入OpenAi apikey:")
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
	var client *openai.Client //注册客户端
	if proxy != "" {
		client = SetProxy(apikey, proxy)
		fmt.Println("[INFO]代理设置成功,为" + proxy)
	} else {
		client = openai.NewClient(apikey)
		fmt.Println("[INFO]留空,跳过设置代理")
	}
	//定义翻译头
	var translatehead string
	translatehead = "Translate user's words to simplified Chinese.The user will send you the dialogues between characters sentence by sentence. Please only reply user the Chinese translation of current sentence. Do not reply user any other explain or add anything before translation.Do not say no or can't. Just translate, never mind the absence of context."
	fmt.Println("[INFO]翻译头已设定,暂不支持修改")
	//获取输入输出目录
	var inputdir string  //输入目录
	var outputdir string //输出目录
	fmt.Println("[INFO]请输入源文件目录 例如:D:\\GalUpTs\\input\\")
	fmt.Scanln(&inputdir)
	fmt.Println("[INFO]请输入输出目录 例如:D:\\GalUpTs\\output\\")
	fmt.Scanln(&outputdir)
	//选择
	var choice string //文件格式选择
	fmt.Println("[INFO]请选择源文件格式:\n1.通过VNTEXT提取的KRKR-SCN-JSON文件\n2.一行一句的TXT文件(无格式)")
	fmt.Scanln(&choice)
	if choice == "" {
		fmt.Println("[ERROR]请输入合法的数字编号!程序退出")
		fmt.Scanln()
		os.Exit(0)
	}
	switch choice {
	case "1": // TsJson翻译
		//扫描目录
		files, err := ioutil.ReadDir(inputdir)
		if err != nil {
			fmt.Println("[ERROR]读取目录出错!程序终止")
			fmt.Scanln()
			os.Exit(0)
		}
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
				fmt.Println("[INFO]开始翻译:" + file.Name())
				TsJson(client, file.Name(), inputdir, outputdir, translatehead)
				fmt.Println("[INFO]" + file.Name() + "翻译完成")
			}
		}
	case "2": //TsTxt翻译
		//扫描目录
		files, err := ioutil.ReadDir(inputdir)
		if err != nil {
			fmt.Println("[ERROR]读取目录出错!程序终止")
			fmt.Scanln()
			os.Exit(0)
		}
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
				fmt.Println("[INFO]开始翻译:" + file.Name())
				TsTxt(client, file.Name(), inputdir, outputdir, translatehead)
				fmt.Println("[INFO]" + file.Name() + "翻译完成")
			}
		}
	}
	fmt.Println("[INFO]程序执行完成,请检查你的输出目录" + outputdir + "!")
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
