package utils

import (
	"io"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

/*
Debug: 適用于開發過程中詳細的日誌紀錄。
Log.Debug("This is a debug message")

Info: 用于一般信息，表示 app 正常運作的信息。
Log.Info("This is an info message")

Warn: 用于警告信息，表示可能會導致問題的情况。
Log.Warn("This is a warning message")

Error: 用于錯誤信息，表示程序發生了錯誤，但未導致程序停止運行。
Log.Error("This is an error message")

Fatal: 用于嚴重錯誤信息，表示程序將停止運行。調用該方法後會自動調用 os.Exit(1) 退出程序。
Log.Fatal("This is a fatal message")

Panic: 用于非常嚴重的錯誤信息，觸發 panic 導致程序崩潰。調用該方法後會引發 panic。
Log.Panic("This is a panic message")
*/

// Logger 是封裝了 logrus.Logger 的單例結構
type Logger struct {
	*logrus.Logger
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger 返回 Logger 的單例實例
func GetLogger() *Logger {
	once.Do(func() {
		logger := logrus.New()

		// 嘗試打開或創建日誌文件
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.Out = file
		} else {
			// 如果無法打開文件，則使用標準輸出並記錄錯誤
			logger.Out = os.Stdout
			logger.Warnf("Failed to log to file, using default stderr: %v", err)
		}

		// 如果是在 release 模式，就丟棄所有輸出，不做任何事
		if gin.Mode() == gin.ReleaseMode {
			logger.Out = io.Discard
		}

		// 設置日誌格式
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})

		// 設置日誌等級：Debug 模式詳細、Release 模式只記錄 Error 以上
		if gin.Mode() == gin.DebugMode {
			logger.SetLevel(logrus.DebugLevel)
		} else {
			logger.SetLevel(logrus.ErrorLevel)
		}

		// 封裝到 Logger 結構中
		instance = &Logger{Logger: logger}
	})

	return instance
}