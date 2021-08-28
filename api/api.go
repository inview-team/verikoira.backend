package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	ginzap "github.com/akath19/gin-zap"
	"github.com/gin-gonic/gin"
	"github.com/inview-team/verikoira.backend/internal/config"
	"github.com/inview-team/verikoira.backend/internal/db"
	"go.uber.org/zap"
)

type KoiraAPI struct {
	http        *http.Server
	mongoClient *db.Storage
}

func New(conf *config.Settings, ctx context.Context) (*KoiraAPI, error) {
	mongo, err := db.New(&conf.Database, ctx)
	if err != nil {
		return nil, err
	}

	k := &KoiraAPI{
		http: &http.Server{
			Addr: net.JoinHostPort(conf.Host, conf.Port),
		},
		mongoClient: mongo,
	}
	k.http.Handler = k.setupRouter()

	return k, nil
}

func (k *KoiraAPI) Run() {
	errs := make(chan error, 1)

	defer func() {
		if err := k.http.Close(); err != nil {
			zap.L().Error("server stopped with error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("server started")
		errs <- k.http.ListenAndServe()
	}()

	err := <-errs
	if err != nil {
		zap.L().Error("server exited with error", zap.Error(err))
	}
}

func (k *KoiraAPI) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginzap.Logger(3*time.Second, zap.L()))

	r.POST("/api/search", k.search)
	r.GET("/api/query")
	r.GET("/api/query/:uuid")

	return r
}

func (k *KoiraAPI) search(c *gin.Context) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to read request body"})
		zap.L().Error("failed to read request body", zap.Error(err))
		return
	}

	query := SearchQuery{}
	err = json.Unmarshal(bodyBytes, &query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to parse request"})
		zap.L().Error("failed to parse request", zap.Error(err))
		return
	}
}
