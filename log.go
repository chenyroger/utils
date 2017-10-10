package utils

import (
	"os"
	"sync"
	"time"
	"fmt"
)

type LogFile struct {
	FileLock *sync.RWMutex
	FileName string
	content  []string
}

func (l *LogFile) Save() error {
	_content := ""
	for _, v := range l.content {
		_content += v + "\n"
	}
	if _content != "" {
		l.FileLock.Lock()
		defer l.FileLock.Unlock()
		//err = os.MkdirAll(logDir, os.ModePerm)
		//if err != nil {
		//	return
		//}
		//

		f, err := os.OpenFile(l.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		l.content = []string{}

		f.Write([]byte(_content))
	}
	return nil
}

func (l *LogFile) SetFileName(fileName string) {
	l.FileName = fileName
}
func (l *LogFile) AddMessage(message string) {
	l.content = append(l.content, fmt.Sprintf("[%v] %v", time.Now().Format("2006-01-02 15:04:05"), message))
}
