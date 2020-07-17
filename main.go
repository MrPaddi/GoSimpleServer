package main

import (
	"bytes"
	"container/list"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)
const DEFAULT_PORT = 9000
var RootDir string
var Port int

const DEFAULT_ERROR_MESSAGE = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN"
"http://www.w3.org/TR/html4/strict.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html;charset=utf-8">
<title>Error response</title>
</head>
<body>
<h1>Error response</h1>
<p>Error code: %d</p>
<p>Message: %s.</p>
</body>
</html>
`

func main() {
	parseServerArgs()
	log.Println("==============================================================")
	log.Printf("Start server at port %d ,mapping direcotry is %s .", Port, RootDir)
	log.Println("--------------------------------------------------------------")
	http.HandleFunc("/", serverHandler)
	addr := fmt.Sprintf("0.0.0.0:%d", Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil{
		log.Fatalf("Create server on %s failed: %s \n", addr, err)
	}
}

func serverHandler(w http.ResponseWriter, request *http.Request) {
	rPath, err := url.QueryUnescape(request.RequestURI)
	defer func() {
		log.Printf("%s %s %s", request.Method, rPath, request.Proto)
	}()
	if err != nil{
		w.Write([]byte(fmt.Sprintf(DEFAULT_ERROR_MESSAGE, 404, "Read dest path failed")))
	}
	destPath := RootDir + rPath
	if !IsExist(destPath){
		w.WriteHeader(404)
		log.Printf("File %s not exists", destPath)
		return
	}
	if !IsDir(destPath){
		if content, err := ioutil.ReadFile(destPath); err == nil{
			w.Write(content)
			return
		}
	} else {
		if content, err := listDir(destPath); err == nil{
			w.Write(GenFolderHtml(content, rPath))
			return
		}
	}
	w.Write([]byte(fmt.Sprintf(DEFAULT_ERROR_MESSAGE, 500, "Read dest path failed")))
}

func listDir(dir string) (*list.List, error) {
	names := list.New()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Read directory %s failed!", dir)
		return names, err
	}
	for _, file := range files{
		displayName := file.Name()
		if file.IsDir(){
			displayName += "/"
		}
		linkName := dir + displayName
		linkName = strings.Replace(linkName, RootDir, "", 1)
		linkName = strings.ReplaceAll(linkName,"\\", "/")
		names.PushBack([2]string{linkName, displayName})
	}
	return names, nil
}


func GenFolderHtml(names *list.List, dirpath string) []byte{
	title := fmt.Sprintf("Directory listing for %s", dirpath)
	htmlContent := bytes.NewBuffer([]byte(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN""http://www.w3.org/TR/html4/strict.dtd">`))
	htmlContent.WriteString("<html>\n<head>\n")
	htmlContent.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html\">\n")
	htmlContent.WriteString(fmt.Sprintf("<title>%s</title>\n</head>\n", title))
	htmlContent.WriteString(fmt.Sprintf("<body>\n<h1>%s</h1>\n", title))
	htmlContent.WriteString("<hr>\n<ul>\n")
	for name := names.Front(); name != nil; name=name.Next(){
		a := name.Value.([2]string)
		htmlContent.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`,a[0], a[1]))
	}
	htmlContent.WriteString("</ul>\n<hr>\n</body>\n</html>\n")
	return htmlContent.Bytes()
}



func IsExist(fileAddr string) bool {
	_,err := os.Stat(fileAddr)
	if err!=nil{
		log.Println(err)
		if os.IsExist(err){  // 根据错误类型进行判断
			return true
		}
		return false
	}
	return true
}

func IsDir(fileAddr string) bool{
	s,err:=os.Stat(fileAddr)
	if err!=nil{
		log.Println(err)
		return false
	}
	return s.IsDir()
}


func parseServerArgs(){
	curWS, err := os.Getwd()
	if err != nil{
		log.Fatal("Get current workspace failed")
	}
	destPtr := flag.String("d", curWS, "Dest workspace, default is current directory.")
	portPrt := flag.Int("p", DEFAULT_PORT, "Server Port")
	flag.Parse()
	destFolder, err := filepath.Abs(*destPtr)
	if !IsDir(destFolder) || err!= nil{
		log.Fatalf("Dest folder `%s` is not found!Please input current folder name", *destPtr)
	}
	RootDir = strings.ReplaceAll(destFolder, "\\", "/")
	Port = *portPrt
}
