package opencc

import (
	"bufio"
	"io"
	"os"
	"strings"
)

//--------- JSON 映射的结构体 start ------------
// 定义 FileDict 类型  用于扩展方法
type FileDict string

type Dict struct {
	Type  string   `json:"type"`
	File  FileDict `json:"file"` // 这里是 FileDict 类型
	Dicts []*Dict  `json:"dicts"`
}

type Segmentation struct {
	Type string `json:"type"`
	Dict *Dict  `json:"dict"`
}

type ConversionChain struct {
	Dict *Dict `json:"dict"`
}

type Config struct {
	Name            string             `json:"name"`
	Segmentation    Segmentation       `json:"segmentation"`
	ConversionChain []*ConversionChain `json:"conversion_chain"`
}

//--------- JSON 映射的结构体 end ------------

// 单个字典 结构体
type DictsDataSet struct {
	File      FileDict
	IsSegment bool
	DataSet   map[string][]string
	MaxLen    int
	MinLen    int
}

var (
	// 多个字典数据
	dictsDataSet map[FileDict]DictsDataSet

	// 字典map集合
	dictsMap map[string][]string

	// 字典中 字符串最大长度
	dictMaxLen int
	// 字典中 字符串最小长度
	dictMinLen int
)

func loadDictData(c *Config) (err error) {
	dictsDataSet = make(map[FileDict]DictsDataSet)
	dictsMap = make(map[string][]string)
	dictMaxLen = 0
	dictMinLen = 0
	return c.extractData()
}

func (c *Config) extractData() (err error) {
	// 字典数据
	var dSet DictsDataSet

	// 分词文件
	if dSet, err = c.Segmentation.Dict.File.readFile(true); err == nil {
		dictsDataSet[c.Segmentation.Dict.File] = dSet
	}

	// 转换文件
	for _, v := range c.ConversionChain {
		if err = v.extractData(); err != nil {
			return
		}
	}
	return
}

func (c *ConversionChain) extractData() (err error) {
	return c.Dict.extractData()
}

func (d *Dict) extractData() (err error) {
	var (
		ok        bool
		childDict *Dict
		dSet      DictsDataSet
	)

	// 如果是文件
	if len(d.File) > 0 {
		// 判断 dictsDataSet 中否已存在值，不存在则读取
		if _, ok = dictsDataSet[d.File]; !ok {
			if dSet, err = d.File.readFile(false); err == nil {
				dictsDataSet[d.File] = dSet
			}
		}
		return
	}

	// 如果是文件组 则读取子文件
	if d.Dicts != nil && len(d.Dicts) > 0 {
		for _, childDict = range d.Dicts {
			// 递归
			if err = childDict.extractData(); err != nil {
				return
			}
		}
	}
	return
}

func (fd *FileDict) readFile(isSegment bool) (dictsDataSet DictsDataSet, err error) {
	var (
		fileName string
		maxLen   int
		minLen   int
		f        *os.File
		buf      *bufio.Reader
		line     string
		fields   []string
		dataset  map[string][]string
	)
	// 需要读取的文件
	fileName = openCCDataDir + "/dictionary/" + string(*fd)
	// 最大/最小字符长度 用于 正向/逆向/双向分词最大匹配
	maxLen = 0
	minLen = 0
	dataset = make(map[string][]string)

	// 打开文件
	if f, err = os.Open(fileName); err != nil {
		return
	}
	defer f.Close()

	// 读文件的Reader
	buf = bufio.NewReader(f)

	for {
		// 按行读取
		if line, err = buf.ReadString('\n'); err != nil {
			// 遇到文件末尾错误则把错误设置成nil
			if err == io.EOF {
				err = nil
			}
			break
		}
		// 空格分割字符串
		fields = strings.Fields(line)
		if len(fields) > 1 {
			// 调整最大字符串长度
			if len([]rune(fields[0])) > maxLen {
				maxLen = len([]rune(fields[0]))
			}
			// 调整最小字符串长度
			if minLen <= 0 || len([]rune(fields[0])) < minLen {
				minLen = len([]rune(fields[0]))
			}
			// 分割后的第一个字符串作为key,剩下的字符串作为value
			dataset[fields[0]] = fields[1:]
			// 同时把值写到 dictsMap
			dictsMap[fields[0]] = fields[1:]
		}
	}

	// 覆盖字典字符中最大和最小长度
	if maxLen > dictMaxLen {
		dictMaxLen = maxLen
	}
	if minLen < dictMinLen {
		dictMinLen = minLen
	}

	// 返回
	return DictsDataSet{
		File:      *fd,
		IsSegment: isSegment,
		DataSet:   dataset,
		MaxLen:    maxLen,
		MinLen:    minLen,
	}, nil
}
