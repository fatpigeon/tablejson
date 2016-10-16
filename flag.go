package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type Config struct {
	Url     string          `toml:"url" json:"url"`
	File    string          `toml:"file" json:"file"`
	Mode    string          `toml:"mode" json:"mode"`
	Output  string          `toml:"output" json:"output"`
	Print   bool            `toml:"print" json:"print"`
	Process []ProcessConfig `default: "process"`
}

type ProcessConfig struct {
	Class   string        `default:"class"`
	Index   int           `default:"index"`
	Col     int           `default:"col"`
	Row     int           `default:"row"`
	Value   string        `default:"value"`
	Replace ReplaceConfig `default:"replace"`
}

type ReplaceConfig struct {
	Target string `default:"target"`
	Re     string `default:"re"`
}

func verfiyMode(s string) bool {
	switch s {
	case "text", "xml":
		return true
	}
	return false
}

func (c *Config) PrintConf() {
	s := reflect.ValueOf(c).Elem()
	typeOfT := s.Type()
	NameValuePairs := map[string]interface{}{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		NameValuePairs[typeOfT.Field(i).Name] = f.Interface()
	}
	b, err := json.MarshalIndent(NameValuePairs, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func GetConfig() Config {
	var conf = Config{}
	var confFile string
	flag.StringVar(&confFile, "config", "config.toml", "配置文件，支持json/toml格式")
	var confType string
	flag.StringVar(&confType, "ctype", "", "配置文件类型，不填则根据后缀去判断")
	var url string
	flag.StringVar(&url, "url", "", "爬取的网络链接")
	var file string
	flag.StringVar(&file, "file", "", "html文件名,如果填写了url该参数将失效")
	var mode string
	flag.StringVar(&mode, "mode", "", "保存的数据格式，可以选择text/xml默认为text")
	var output string
	flag.StringVar(&output, "output", "", "输出文件,不填则则输出到标准输出")
	var printConf bool
	flag.BoolVar(&printConf, "print", false, "打印当前配置信息")
	flag.Parse()
	//从配置文件读取配置
	if confFile != "" {
		f, err := ioutil.ReadFile(confFile)
		if err == nil {
			if confType == "toml" || strings.HasSuffix(confFile, ".toml") {
				err = toml.Unmarshal(f, &conf)
				if err != nil {
					panic(err)
				}
			} else if confType == "json" || strings.HasSuffix(confFile, ".json") {
				err = json.Unmarshal(f, &conf)
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println("目前不支持这种格式的配置文件")
			}
		} else {
			fmt.Errorf("%e, %s", err, "配置文件读取失败")
		}
	}
	//从命令行参数读取配置
	if url != "" {
		conf.Url = url
	}
	if file != "" {
		conf.File = file
	}
	if mode != "" {
		conf.Mode = mode
	}
	if output != "" {
		conf.Output = output
	}
	if printConf != false {
		conf.Print = printConf
	}
	if conf.Mode == "" {
		conf.Mode = "text"
	}

	//校验配置是否有效
	verfiy := true
	if verfiyMode(conf.Mode) == false {
		fmt.Println("mode不是有效值", conf.Mode)
		verfiy = false
	}
	if conf.File == "" && conf.Url == "" {
		fmt.Println("文件和链接都为空")
		verfiy = false
	}
	if verfiy == false {
		os.Exit(0)
	}
	return conf
}
