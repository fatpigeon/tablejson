package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Table struct {
	Heads []string
	Bodys [][]string
}

func getUrlData(url string) *html.Node {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}
	return node

}

func getFileData(file string) *html.Node {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	node, err := html.Parse(f)
	if err != nil {
		panic(err)
	}
	return node
}

func findTable(node *html.Node) []*html.Node {
	tables := []*html.Node{}
	if node.Type == html.ElementNode && node.Data == "table" {
		tables = append(tables, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		tables = append(tables, findTable(c)...)
	}
	return tables
}

/*
* node值可能为空字符串，能拿到下个有效值增加该函数
 */
func realityNode(node *html.Node) *html.Node {
	for node != nil && node.Type != html.ElementNode {
		node = node.NextSibling
	}
	return node
}

type GetDataAdapter func(node *html.Node) string

/*
* 返回带标签的节点数据
 */
func GetNodeXmlData(node *html.Node) string {
	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	} else if node.Type == html.ElementNode {
		data := fmt.Sprintf("<%s>", node.Data)
		for sr := node.FirstChild; sr != nil; sr = sr.NextSibling {
			d := GetNodeXmlData(sr)
			if d != "" {
				data += d
			}
		}
		data += fmt.Sprintf("</%s>", node.Data)
		return data
	}
	return ""
}

/*
* 返回带纯字符串，如果节点下面有多个字符串以空格分割
 */
func GetNodeTextData(node *html.Node) string {
	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	} else if node.Type == html.ElementNode {
		dataList := []string{}
		for sr := node.FirstChild; sr != nil; sr = sr.NextSibling {
			d := GetNodeTextData(sr)
			if d != "" {
				dataList = append(dataList, d)
			}
		}
		data := strings.Join(dataList, " ")
		return data
	}
	return ""
}

/*
* 表的格式要注意几点
* 1. thead中带一个tr
<table>
 <thead>
   <tr>
    <th></th>
	...
   </tr>
 </thead>
 <tbody>
	 ...
 <tbody>
</table>
* 2. thead中直接使用th
<table>
  <thead>
    <th></th>
	...
  </thead>
 <tbody>
  ...
 <tbody>
</table>

* 3. 不带thead的
<table>
  <tbody>
  ...
  <tbody>
</table>
* 4. tbody也不带的
<table>
  <tr></tr>
  ...
</table>
*/
func CreateTable(node *html.Node, get GetDataAdapter) Table {
	t := Table{}
	curr := realityNode(node.FirstChild)
	//thead parse
	if curr != nil && curr.Data == "thead" {
		thead := curr
		curr = realityNode(curr.NextSibling)
		tr := realityNode(thead.FirstChild)
		if tr == nil || tr.Data != "tr" {
			tr = thead
		}
		for th := realityNode(tr.FirstChild); th != nil; th = realityNode(th.NextSibling) {
			if th.Data == "th" {
				t.Heads = append(t.Heads, get(th))
			}
		}
	}
	//tbody parse
	if curr != nil {
		if curr.Data != "tbody" {
			curr = node
		}
		type MultiLineTd struct {
			Number int
			Data   string
		}
		carry := map[int]*MultiLineTd{}
		for tr := realityNode(curr.FirstChild); tr != nil; tr = realityNode(tr.NextSibling) {
			index := 0
			trData := []string{}
			for td := realityNode(tr.FirstChild); td != nil; {
				if _, ok := carry[index]; ok && carry[index].Number > 0 {
					trData = append(trData, carry[index].Data)
					carry[index].Number--
				} else {
					for _, attr := range td.Attr {
						if attr.Key == "rowspan" {
							val, err := strconv.Atoi(attr.Val)
							if err != nil {
								fmt.Errorf(err.Error())
								break
							}
							carry[index] = &MultiLineTd{val - 1, get(td)}
							break
						}
					}
					trData = append(trData, get(td))
					td = realityNode(td.NextSibling)
				}
				index++
			}
			t.Bodys = append(t.Bodys, trData)
		}
	}
	return t
}

/*
* 入口函数
*
 */
func Portal(url string, file string, mode string) string {
	var node *html.Node
	if url != "" {
		node = getUrlData(url)
	} else {
		node = getFileData(file)
	}
	var adapter GetDataAdapter
	switch mode {
	case "text":
		adapter = GetNodeTextData
	case "xml":
		adapter = GetNodeXmlData
	}
	tNodes := findTable(node)
	tables := []Table{}
	for _, t := range tNodes {
		tables = append(tables, CreateTable(t, adapter))
	}
	b, err := json.Marshal(tables)
	if err != nil {
		panic(err)
	}

	return string(b)
}
