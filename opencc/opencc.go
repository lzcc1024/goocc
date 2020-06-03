package opencc

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// type Opencc interface {}

type OpenCC struct {
	config       *Config
	dictsDataSet map[FileDict]DictsDataSet
	dictsMap     map[string][]string
	dictMaxLen   int
	dictMinLen   int
}

var (
	openCCDataDir string
	punctuations  []string = []string{
		" ", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "-", "=", "_", "=", "`", "~", "[", "]", "{", "}", "\\", "|", ";", ":", "'", "\"", ",", ".", "/", "<", ">", "?", "　", "～", "！", "￥", "…", "×", "（", "）", "—", "【", "】", "、", "；", "：", "‘", "’", "“", "”", "，", "。", "《", "》", "？", "／", "＼", "「", "」", "－", "．", "·"}
)

func init() {
	// 从参数中获取数据的目录
	flag.StringVar(&openCCDataDir, "openccd", "", "OpenCC Data Dir")
	flag.Parse()
}

func NewOpenCC(conversions string) (*OpenCC, error) {
	var (
		fileRealPath        string
		conversionsFileName string
		err                 error
		body                []byte
		config              *Config
	)

	// 判断从参数中获取的目录是否存在，，赋值 openCCDataDir
	if _, err = os.Stat(openCCDataDir); err != nil {
		fileRealPath, _ = filepath.Abs(os.Args[0])
		openCCDataDir = filepath.Dir(fileRealPath) + "/data"
	} else {
		openCCDataDir, _ = filepath.Abs(openCCDataDir)
	}

	// JSON配置文件的路径
	conversionsFileName = openCCDataDir + "/config/" + conversions + ".json"

	// 读取JSON配置文件
	body, err = ioutil.ReadFile(conversionsFileName)
	if err != nil {
		return nil, err
	}

	// 解析JSON 映射到结构体
	if err = json.Unmarshal(body, &config); err != nil {
		return nil, err
	}

	// 提取出所有用到的字典文件名
	if err = loadDictData(config); err != nil {
		return nil, err
	}

	// 返回 OpenCC 实例对象
	return &OpenCC{config: config, dictsDataSet: dictsDataSet, dictsMap: dictsMap, dictMaxLen: dictMaxLen, dictMinLen: dictMinLen}, nil
}

func (oc *OpenCC) Convert(content string) (text string, err error) {
	// 空对象或空字典
	if oc == nil || len(oc.dictsMap) <= 0 {
		return content, nil
	}

	// 空字符串或者纯数字字符串
	if isEmpty(content) || isNum(content) {
		return content, nil
	}

	// 使用标点符号进行句子分隔并转换
	if text, err = oc.splitConvert(content); err != nil {
		return content, nil
	}

	return
}

func (oc *OpenCC) splitConvert(content string) (text string, err error) {
	var (
		sentences     []string
		convertedStr  string
		sentencesText string
	)
	sentences = make([]string, 0, len(content))

	for i, c := range strings.Split(content, "") {
		// 如果第一个字符是标点符号,就直接拼接
		if i == 0 && isPunctuations(c) {
			text += c
			continue
		}

		// 从第二个字符开始，如果是常规标点符号
		if i > 0 && isPunctuations(c) {
			if len(sentences) > 0 {
				// 进行转换
				sentencesText = strings.Join(sentences, "")
				if isChinese(sentencesText) {
					// 语句中有中文
					if convertedStr, err = oc.convertString(sentencesText); err == nil {
						// 转换成功 拼接返回值
						text += convertedStr + c
					} else {
						// 某次转换失败 则原样返回字符串
						return content, nil
					}
				} else {
					// 语句中无中文 拼接返回值
					text += sentencesText + c
				}
				// 清空切片
				sentences = sentences[:0]
			} else {
				// 标点符号，直接拼接
				text += c
			}
			continue
		}
		// 把字存到切片中，待转换
		sentences = append(sentences, c)
	}

	// 全文无标点符号 或 最后一个标点符号后面的字符
	if len(sentences) > 0 {
		sentencesText = strings.Join(sentences, "")

		if isChinese(sentencesText) {
			// 语句中有中文
			if convertedStr, err = oc.convertString(sentencesText); err == nil {
				// 转换成功 拼接返回值
				text += convertedStr
			} else {
				// 转换失败 则原样返回字符串
				return content, nil
			}
		} else {
			text += sentencesText
		}
	}
	// 清空切片
	sentences = sentences[:0]
	return
}

// 判断字符串是否为真空
func isEmpty(str string) bool {
	if len(strings.TrimLeft(str, " ")) > 0 {
		return false
	} else {
		return true
	}
}

// 判断是不是纯数字字符串，64位架构计算机 字符长度为限制309
func isNum(str string) bool {
	_, err := strconv.ParseFloat(strings.Trim(str, " "), 64)
	return err == nil
}

// 判断单个字符是不是常规标点符号
func isPunctuations(character string) bool {
	for _, c := range punctuations {
		if c == character {
			return true
		}
	}
	return false
}

// 判断语句中是否有中文
func isChinese(str string) bool {
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			return true
		}
	}
	return false
}

func (oc *OpenCC) convertString(content string) (text string, err error) {

	var (
		runes  []rune
		maxlen int
	)
	runes = []rune(content)
	text = content

	// 空字典
	if len(oc.dictsMap) <= 0 {
		return content, nil
	}

	// 文本长度 小于 字典中最小字符串长度
	if len(runes) < oc.dictMinLen {
		return content, nil
	}

	if oc.dictMaxLen > len(runes) {
		// 文本长度 小于 字典中最大字符串长度 ，减少循环次数
		maxlen = len(runes)
	} else {
		maxlen = oc.dictMaxLen
	}

	for i := maxlen; i >= oc.dictMinLen; i-- {
		for j := 0; j <= len(runes)-i; j++ {
			if i == 0 || j+i > len(runes) {
				continue
			}

			old := string(runes[j : j+i])
			if new, ok := oc.dictsMap[old]; ok {
				text = strings.Replace(text, old, new[0], 1)
				j = j + i - 1
			}
		}

	}
	return
}
