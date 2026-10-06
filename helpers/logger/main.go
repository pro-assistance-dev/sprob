package logger

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
)

func NewLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logLevel())
	setupOutput(l)
	return l
}

// logLevel: DEBUG по умолчанию раздувает лог (каждая SQL-операция) до ~65 МБ/ч
// на сервис — это и был ENOSPC 05.10. Уровень задаётся LOG_LEVEL (по умолчанию info).
func logLevel() logrus.Level {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

// logMaxAge: окно хранения логов (дни), переопределяется LOG_MAX_AGE_DAYS.
func logMaxAge() time.Duration {
	days, err := strconv.Atoi(os.Getenv("LOG_MAX_AGE_DAYS"))
	if err != nil || days <= 0 {
		days = 3
	}
	return time.Hour * 24 * time.Duration(days)
}

func setupOutput(l *logrus.Logger) {
	path := "logs"
	l.SetOutput(io.Discard) // Send all logs to nowhere by default

	tForm := "%Y-%m-%d_%H:%M"
	ageOpt := rotatelogs.WithMaxAge(logMaxAge())
	rotateTimeOpt := rotatelogs.WithRotationTime(time.Hour)

	infoPath := path + "/info/"
	infoActualFileLink := rotatelogs.WithLinkName(infoPath + "_actual.log")
	infoWriter, _ := rotatelogs.New(infoPath+tForm+".log", infoActualFileLink, ageOpt, rotateTimeOpt)

	errorsPath := path + "/errors/"
	errorsActualFileLink := rotatelogs.WithLinkName(errorsPath + "_actual.log")
	errorsWriter, _ := rotatelogs.New(errorsPath+tForm+".log", errorsActualFileLink, ageOpt, rotateTimeOpt)

	l.Hooks.Add(lfshook.NewHook(
		lfshook.WriterMap{
			// INFO
			logrus.InfoLevel:  infoWriter,
			logrus.DebugLevel: infoWriter,
			logrus.WarnLevel:  infoWriter,
			logrus.TraceLevel: infoWriter,
			// ERROR
			logrus.ErrorLevel: errorsWriter,
			logrus.FatalLevel: errorsWriter,
			logrus.PanicLevel: errorsWriter,
		},
		&logrus.JSONFormatter{PrettyPrint: true, DisableHTMLEscape: true},
	))
}

func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		// query := map[string]string{}
		//err := ctx.Bind(query)
		//if err != nil {
		//	fmt.Println(err)
		//}
		//formBody := map[string]string{}
		//form, err := ctx.MultipartForm()
		//if err != nil {
		//	fmt.Println(err)
		//}
		//err = json.Unmarshal([]byte(form.Value["form"][0]), &formBody)
		//if err != nil {
		//	fmt.Println(err)
		//}

		ctx.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := ctx.Request.Method
		path := ctx.Request.URL.Path
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		clientUserAgent := ctx.Request.UserAgent()
		// referer := ctx.Request.Referer()

		entry := logger.WithFields(logrus.Fields{
			"method":          reqMethod,
			"path":            path,
			"status":          statusCode,
			"latency":         latencyTime,
			"clientIp":        clientIP,
			"clientUserAgent": clientUserAgent,
			"errors":          ctx.Errors.ByType(gin.ErrorTypePrivate).String(),
			// "body":            query,
			//"formBody": formBody,
		})

		if len(ctx.Errors) > 0 {
			entry.Error(ctx.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			switch {
			case statusCode >= http.StatusInternalServerError:
				entry.Error()
			case statusCode >= http.StatusBadRequest:
				entry.Warn()
			default:
				entry.Info()
			}
		}

		ctx.Next()
	}
}
