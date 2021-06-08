package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tipRPC "github.com/tktip/flyvo-api/pkg/rpc"
	"github.com/tktip/flyvo-rpc-client/internal/log"
	"github.com/tktip/flyvo-rpc-client/internal/rpc"
)

type Server struct {
	Port           string        `yaml:"port"`
	RequestTimeout time.Duration `yaml:"timeout"`
	RpcClient      *rpc.Client   `yaml:"rpc"`
}
type ActivityRequest struct {
	Activity tipRPC.Event `json:"activity"`
}

// generic proxies generic request to TIP
// @Summary proxies generic request to TIP
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "Result provided by tip"
// @Failure 422 {string} string "On bad request body"
// @Failure 500 {string} string "On rpc error"
// @Router /generic [POST]
func (s *Server) SendGenericRequest(c *gin.Context) {
	g := tipRPC.Generic{}
	err := c.BindJSON(&g)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	}

	if g.Path == "" {
		c.String(http.StatusBadRequest, "no path provided in request body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, err := s.RpcClient.SendGeneric(ctx, tipRPC.Generic{
		MsgID:   g.MsgID,
		Headers: g.Headers,
		Body:    []byte(g.Body),
		Path:    g.Path,
	})

	if err != nil {
		log.Logger.Errorf("Error on generic rpc call: %s", err.Error())
		c.String(http.StatusInternalServerError, "Error on generic rpc call: "+err.Error())
		return
	}

	log.Logger.Debugf("Received rpc response from server: %+v", resp)
	c.Writer.WriteHeader(int(resp.Status))
	for h, v := range resp.Headers {
		c.Header(h, v)
	}

	c.Writer.Write(resp.Body)
}

// generic proxies create event request to TIP
// @Summary proxies create event request to TIP
// @Accept application/json
// @Produce application/json
// @Param body body rpc.Event true "event"
// @Success 200 {string} string "Body is yet to be defined"
// @Failure 422 {string} string "On bad request body"
// @Failure 500 {string} string "On rpc error"
// @Router /events [POST]
func (s *Server) PostEvent(c *gin.Context) {
	//bod, _ := ioutil.ReadAll(c.Request.Body)
	//c.Writer.WriteHeader(http.StatusOK)
	//log.Logger.Debugf("Body: %s", bod)
	//return

	actReq := ActivityRequest{}
	err := c.BindJSON(&actReq)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	}

	eventjson, _ := json.Marshal(actReq)
	log.Logger.Debugf("Post event: %s", eventjson)

	ctx, cancel := context.WithTimeout(context.Background(), s.RequestTimeout)
	defer cancel()

	response, err := s.RpcClient.PostEvent(ctx, &actReq.Activity)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		log.Logger.Errorf("Failed to post event: %s", err.Error())
		return
	}

	respJson, _ := json.Marshal(response)
	log.Logger.Debugf("POST Response: %s", respJson)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(response.Body)
}

// generic proxies patch event request to TIP
// @Summary proxies patch event request to TIP
// @Accept application/json
// @Produce application/json
// @Param body body rpc.Event true "event"
// @Success 200 {string} string "Body is yet to be defined"
// @Failure 422 {string} string "On bad request body"
// @Failure 500 {string} string "On rpc error"
// @Router /events [PUT]
func (s *Server) PutEvent(c *gin.Context) {

	actReq := ActivityRequest{}
	err := c.BindJSON(&actReq)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, err.Error())
		return
	}

	eventjson, _ := json.Marshal(actReq)
	log.Logger.Debugf("Put event: %s", eventjson)

	ctx, cancel := context.WithTimeout(context.Background(), s.RequestTimeout)
	defer cancel()

	response, err := s.RpcClient.PutEvent(ctx, &actReq.Activity)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to post event: "+err.Error())
		log.Logger.Errorf("Failed to post event: %s", err.Error())
		return
	}

	respJson, _ := json.Marshal(response)
	log.Logger.Debugf("PUT Response: %s", respJson)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(response.Body)
}

// generic proxies create event request to TIP
// @Summary proxies create event request to TIP
// @Accept application/json
// @Produce application/json
// @Param id path string true "event to be deleted"
// @Success 200 {string} string "Body is yet to be defined"
// @Failure 422 {string} string "On bad request body"
// @Failure 500 {string} string "On rpc error"
// @Router /events/id [DELETE]
func (s *Server) DeleteEvent(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.String(http.StatusBadRequest, "no id provided")
		return
	}

	log.Logger.Debugf("Delete event: %s", id)

	ctx, cancel := context.WithTimeout(context.Background(), s.RequestTimeout)
	defer cancel()

	response, err := s.RpcClient.DeleteEvent(ctx, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete event: "+err.Error())
		log.Logger.Errorf("Failed to post event: %s", err.Error())
		return
	}

	respJson, _ := json.Marshal(response)
	log.Logger.Debugf("DELETE Response: %s", respJson)

	c.Writer.WriteHeader(int(response.Status))
	c.Writer.Write(response.Body)
}

func (s *Server) PingRPCServer(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), s.RequestTimeout)
	defer cancel()
	response, err := s.RpcClient.SendGeneric(ctx, tipRPC.Generic{
		Path: "ping",
	})

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
	c.Writer.Write(response.Body)
}

func (s *Server) ConnAlive(c *gin.Context) {

	if !s.RpcClient.ConnOK {
		c.String(http.StatusInternalServerError, "not connected to RPC server")
	} else {
		c.String(http.StatusOK, "connected to RPC server")
	}
}

func (s *Server) Run(ctx context.Context) error {
	if s.RequestTimeout < time.Second {
		s.RequestTimeout = time.Second * 1
		log.Logger.Warnf("RequestTimeout was not set or set to less than 1 sec. Set to 1 second")
	}

	rpcCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go s.RpcClient.Run(rpcCtx)

	g := gin.New()
	log.Logger.Infof("Starting gin at port %s", s.Port)
	g.GET("/ping", s.PingRPCServer)
	g.GET("/alive", s.ConnAlive)
	g.POST("/generic", s.SendGenericRequest)
	g.POST("/events", s.PostEvent)
	g.PUT("/events", s.PutEvent)
	g.DELETE("/events/:id", s.DeleteEvent)
	return g.Run(":" + s.Port)
}
