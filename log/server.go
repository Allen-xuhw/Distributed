package log

import (
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
)

var log *stdlog.Logger

type fileLog string

//将数据写入文件
func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(data)
}

func Run(destination string) {
	log =stdlog.New(fileLog(destination), "[go] - ", stdlog.LstdFlags)
}

//相当于注册一个逻辑处理函数func(w http.ResponseWriter, r *http.Request)
func RegisterHandlers() {
	//调用函数http.HandleFunc，其参数包含另一个函数
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

func write(message string) {
	log.Printf("%v\n", message)
}