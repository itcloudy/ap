package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/dot"
)

var (
	graphMap     = map[string][]string{}
	graphDot     = dot.NewGraph("G")
	contr2Contr  = regexp.MustCompile("[^(Join|info|warning|error|LangRes|FindEcosystem|CallContract|ContractAccess|ContractConditions|EvalCondition|ValidateCondition|AddressToId|Contains|Float|HasPrefix|HexToBytes|Int|Len|PubToID|IdToAddress|Money|Replace|Size|Sha256|Sprintf|Str|Substr|UpdateLang|SysParamString|SysParamInt|UpdateSysParam|EcosysParam|DBFind|DBInsert|DBInsertReport|DBUpdate|DBUpdateExt|DBRow|DBIntExt|DBStringExt)]\\s*\\(@?.*?\\)")
	page2Contr   = regexp.MustCompile("\\(.*?Contract:\\s*(@?\\w+)")
	page2Page    = regexp.MustCompile("\\(.*?Page:\\s*(\\w+)")
	contr2Table  = regexp.MustCompile("(?:DBFind|DBInsert|DBUpdate|DBUpdateExt|DBRow)\\s*\\(\\s*[\"\\`]([\\w]+?)[\"`]")
	page2Table   = regexp.MustCompile("DBFind\\s*\\(\\s*Name:\\s*(.*?)[,\\s]|DBFind\\s*\\(\\s*([^:]*?)[\\),\\s]")
	includeBlock = regexp.MustCompile("Include\\s*\\(\\s*Name:\\s*(.*?)[,\\s]|Include\\s*\\(\\s*([^:]*?)[\\),\\s]")
)

func initGraph() {
	graphDot.SetType(dot.DIGRAPH)
	graphDot.Set("rankdir", "LR")
	graphDot.Set("fontsize", "30.0")
	labelGraph := fmt.Sprintf("%s\n%s", strings.Trim(outputName, separator), time.Now().Format(time.RFC850))
	graphDot.Set("label", labelGraph)
}

func addEdges(parentNode *dot.Node, s, dir string) {
	switch dir {
	case dirCon:
		addNode(parentNode, contr2Contr, s, dir, "")
		addNode(parentNode, contr2Table, s, dirTable, "")
	case dirPage:
		addNode(parentNode, page2Contr, s, dirCon, "")
		addNode(parentNode, page2Table, s, dirTable, "")
		addNode(parentNode, page2Page, s, dir, "")
		addNode(parentNode, includeBlock, s, dirBlock, "Include")
	case dirBlock:
		addNode(parentNode, page2Contr, s, dirCon, "")
		addNode(parentNode, page2Table, s, dirTable, "")
		addNode(parentNode, page2Page, s, dir, "")
	case dirMenu:
		addNode(parentNode, page2Page, s, dirPage, "")
		// fmt.Println(graphDot)
	}
}
func createNodeForString(name, dir, value string) {
	switch dir { // parse graph
	case dirPage:
		fallthrough
	case dirCon:
		fallthrough
	case dirTable:
		fallthrough
	case dirBlock:
		fallthrough
	case dirMenu:
		node := dot.NewNode(getNodeName(name, dir))
		if dir == dirPage || dir == dirBlock {
			node.Set("fontcolor", pageColor)
			node.Set("color", pageColor)
		}
		if dir == dirCon {
			node.Set("fontcolor", contrColor)
			node.Set("color", contrColor)
		}
		if dir == dirMenu {
			node.Set("fontcolor", menuColor)
			node.Set("color", menuColor)
		}
		group := parseGroup(name, dir)
		node.Set("group", group)
		if dir != dirTable {
			addEdges(node, value, dir)
		}
		graphDot.AddNode(node)
	}
}
func addNode(parentNode *dot.Node, pat *regexp.Regexp, str, dir, label string) {
	s := strings.Replace(str, "`", `"`, -1)
	arr := pat.FindAllStringSubmatch(s, -1)
	for _, match := range arr {
		for i := range match {
			if i > 0 {
				if match[i] != "" {
					name := getNodeName(match[i], dir)
					if !stringInSlice(graphMap[parentNode.Name()], name) { // check exist node tops
						group := parseGroup(name, dir)
						name = strings.Trim(name, `"`)
						name = strings.Trim(name, "`")
						node := dot.NewNode(name)
						node.Set("group", group)
						if _, ok := graphMap[parentNode.Name()]; !ok {
							graphMap[parentNode.Name()] = []string{}
						}
						edge := dot.NewEdge(parentNode, node)
						if label != "" {
							edge.Set("label", label)
						}
						switch dir {
						case dirPage:
							edge.Set("color", pageColor)
						case dirCon:
							edge.Set("color", contrColor)
						case dirBlock:
							edge.Set("color", pageColor)
						case dirMenu:
							edge.Set("color", menuColor)
						}
						graphDot.AddEdge(edge)
						graphMap[parentNode.Name()] = append(graphMap[parentNode.Name()], name)
					}
				}
			}
		}
	}
}

func getNodeName(name, dir string) (_name string) {
	_name = fmt.Sprintf("%s\n%s", name, strings.TrimSuffix(dir, "s"))
	if strings.Contains(_name, ",") {
		_name = strings.Join(strings.Split(_name, ","), "\n")
	}
	return
}

func writeGraph(name string) {
	outFile, err := os.Create(name)
	if err != nil {
		fmt.Println("error write file:", err)
		return
	}
	defer outFile.Close()
	if _, err := outFile.WriteString(graphDot.String()); err != nil {
		fmt.Println(err)
		return
	}
}

func parseGroup(n, dir string) string {
	name := underscore(n)
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		return strings.ToLower(parts[0])
	}
	return dir
}

var camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")

func underscore(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}
