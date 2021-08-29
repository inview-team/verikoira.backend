package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	ginzap "github.com/akath19/gin-zap"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/inview-team/verikoira.backend/internal/config"
	"github.com/inview-team/verikoira.backend/internal/db"
	"github.com/inview-team/verikoira.backend/internal/rmq"
	"go.uber.org/zap"
)

const addr = "amqp://rmq:5672"

type KoiraAPI struct {
	http        *http.Server
	publisher   *rmq.Publisher
	mongoClient *db.Storage
}

type SearchQuery struct {
	Payload string `json:"payload"`
}

type QueryToSend struct {
	TaskID  string `json:"task_id"`
	Payload string `json:"payload"`
}

func New(conf *config.Settings, ctx context.Context) (*KoiraAPI, error) {
	mongo, err := db.New()
	if err != nil {
		return nil, err
	}
	k := &KoiraAPI{
		http: &http.Server{
			Addr: net.JoinHostPort(conf.Host, conf.Port),
		},
		publisher:   rmq.NewPublisher(addr, "tasks"),
		mongoClient: mongo,
	}
	k.http.Handler = k.setupRouter()

	zap.L().Debug("connecting to RMQ", zap.String("address", conf.Rmq.Address))

	return k, nil
}

func (k *KoiraAPI) Run() {
	errs := make(chan error, 1)

	defer func() {
		if err := k.http.Close(); err != nil {
			zap.L().Error("server stopped with error", zap.Error(err))
		}
	}()

	err := k.publisher.Reconnect()
	if err != nil {
		zap.L().Error("failed to connect to RabbitMQ", zap.Error(err))
		return
	}

	go func() {
		zap.L().Info("server started")
		errs <- k.http.ListenAndServe()
	}()

	err = <-errs
	if err != nil {
		zap.L().Error("server exited with error", zap.Error(err))
	}
}

func (k *KoiraAPI) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())
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

	to_send := QueryToSend{
		TaskID:  uuid.New().String(),
		Payload: query.Payload,
	}
	res, err := json.Marshal(to_send)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to send query"})
		zap.L().Error("failed to marshal JSON", zap.Error(err))
		return
	}

	err = k.publisher.Send(res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to perform search request"})
		zap.L().Error("failed to perform search request", zap.Error(err))
		return
	}

	/*	resChan, err := k.consumer.Reconnect(context.TODO())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to connect to RabbitMQ"})
			zap.L().Error("failed to connect to RabbitMQ", zap.Error(err))
			return
		}
		result := <-resChan
		data := result.Body
		c.JSON(http.StatusOK, gin.H{"result": string(data)})
		zap.L().Debug("received data", zap.String("result", string(data)))
		err = result.Ack(false)
		if err != nil {
			zap.L().Error("failed to ack message", zap.Error(err))
		}*/
}
