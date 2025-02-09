package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/siriusa51/webtty/session"
	templates "github.com/siriusa51/webtty/templates"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type RouterConfig struct {
	Host       string
	Port       int
	PrefixPath string
	IndexFile  string
	Workdir    string
	Command    string
	ExtraEnv   []string
}

func NewHandler(config RouterConfig, log *slog.Logger, mgr *session.SessionManager) http.Handler {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	router.SetHTMLTemplate(templates.GetTemplate("*"))

	ctrl := NewController(
		ControllerConfig{Workdir: config.Workdir, Command: config.Command, ExtraEnv: config.ExtraEnv},
		log, mgr,
	)
	prefixPath := strings.TrimRight(config.PrefixPath, "/")

	if prefixPath == "" {
		prefixPath = "/"
	}

	router.GET(prefixPath, func(context *gin.Context) {
		if config.IndexFile != "" {
			content, err := os.ReadFile(config.IndexFile)
			if err != nil {
				context.String(http.StatusInternalServerError, "failed to read index file: %v", err)
				return
			}

			context.Data(http.StatusOK, "text/html", content)
			return
		}

		context.HTML(http.StatusOK, "index.html", gin.H{
			"prefix_path": prefixPath,
		})
	})

	router.GET(prefixPath+"/favicon.ico", func(c *gin.Context) {
		data, err := templates.GetFile("favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Data(http.StatusOK, "image/x-icon", data)
	})

	router.Any(prefixPath+"/remove_session", ctrl.RemoveSession)
	router.GET(prefixPath+"/ws", ctrl.Websocket)

	log.Info("command -> " + config.Command)
	log.Info("workdir -> " + config.Workdir)
	addr := fmt.Sprintf("http://%v:%v%v", config.Host, config.Port, prefixPath)
	log.Info("please visit " + addr)

	return router
}
