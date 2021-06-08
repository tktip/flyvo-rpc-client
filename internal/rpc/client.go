package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	tipRPC "github.com/tktip/flyvo-api/pkg/rpc"
	"github.com/tktip/flyvo-rpc-client/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Flyvo struct {
	RootAddress string `yaml:"address"`
}

type Client struct {
	RpcServerAddress  string `yaml:"serverAddress"`
	RpcCertFile       string `yaml:"certFile"`
	FlyvoApiEndpoints Flyvo  `yaml:"flyvo"`

	PollFrequency   time.Duration  `yaml:"pollFrequency"`
	BadConnectSleep time.Duration  `yaml:"connFailSleep"`
	ConnTimeout     *time.Duration `yaml:"connTimeout"`

	ConnOK     bool
	ctx        context.Context
	grpcConn   *grpc.ClientConn
	tipClient  tipRPC.TipFlyvoClient
	done       bool
	inFlightWg sync.WaitGroup
	httpClient *http.Client
}

func (c *Client) Run(ctx context.Context) {
	if c.RpcServerAddress == "" {
		c.RpcServerAddress = defaultAddress
		log.Logger.Warnf("No address provided, defaulting to '%s'", defaultAddress)
	}

	if c.ConnTimeout == nil {
		log.Logger.Warn("No timeout provided, defaulting to 15s")
		t := time.Second * 15
		c.ConnTimeout = &t
	}

	var err error
	opts := []grpc.DialOption{}
	if c.RpcCertFile != "" {
		var creds credentials.TransportCredentials

		// Create the client TLS credentials
		creds, err = credentials.NewClientTLSFromFile(c.RpcCertFile, "")
		if err != nil {
			log.Logger.Fatalf("could not load tls cert: %s", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
		log.Logger.Infof("TLS certificate registered")
	} else {

		log.Logger.Infof("Running without TLS (insecure)")
		opts = append(opts, grpc.WithInsecure())
	}

	// Set up a tipClient to the server.
	c.grpcConn, err = grpc.Dial(c.RpcServerAddress, opts...)
	if err != nil {
		log.Logger.Fatalf("did not connect: %v", err)
	}
	c.tipClient = tipRPC.NewTipFlyvoClient(c.grpcConn)
	c.ctx = ctx
	c.pollServerForGenericRequests()
}

//Contacts TIP every 5 seconds and asks TIP to create an event.
//stalls
func (c *Client) pollServerForGenericRequests() {
	go func() {
		c.inFlightWg.Add(1)

		for !c.done {
			ctx, cancel := context.WithTimeout(context.Background(), *c.ConnTimeout)

			//This runs the function ProcessRequests in flyvo-api.
			//It returns a stream object through which requests are sent and received.

			log.Logger.Debugf("Trying to connect to TIP (timeout %s)", *c.ConnTimeout)
			pollConnection, err := c.tipClient.ProcessRequests(ctx)

			//If the connection attempt failed, no point in doing anything.
			if err != nil {
				c.ConnOK = false
				log.Logger.Errorf("Failed to connect to TIP: %v", err)
				log.Logger.Debugf("Sleeping for %s", c.BadConnectSleep)
				time.Sleep(c.BadConnectSleep)
				cancel()
				continue
			}

			if !c.ConnOK {
				c.ConnOK = true
				log.Logger.Info("Successfully connected to TIP again.")
			}

			//While the stream is open, grab incoming data
			for {
				log.Logger.Debug("Retrieving")
				request, err := pollConnection.Recv()

				//io.EOF means that the connection was closed at the other end (flyvo-api).
				if err == io.EOF {
					log.Logger.Debug("poll EOF")
					break
				} else if err != nil {
					log.Logger.Errorf("Failed to receive generic request from TIP: %s", err.Error())
					break
				}

				log.Logger.Debugf("Received this: path[%v], headers[%v], body[%s]",
					request.Path,
					request.Headers,
					request.Body,
				)

				//Do some processing of the received request
				response, err := c.handleGenericRequest(*request)
				if err != nil {
					log.Logger.Errorf("Failed to process request with msgID %s: %s", request.MsgID, err.Error())
				}

				//Then respond to flyvo-api with the result of processing.
				err = pollConnection.Send(&response)
				if err != nil {
					log.Logger.Errorf("Failed to respond to request with msgID %s: %s", request.MsgID, err.Error())
				}
			}

			//Send EOF to flyvo-api, indicating that we're done.
			err = pollConnection.CloseSend()
			if err != nil {
				log.Logger.Errorf("Error during close send: %s", err.Error())
			}

			//cancel lingering context since we're done pre-timeout
			cancel()
			log.Logger.Debug("Done looking for requests")
			log.Logger.Debugf("Sleeping for %s", c.PollFrequency)
			time.Sleep(c.PollFrequency)
		}
		log.Logger.Debug("I'm done")
		c.inFlightWg.Done()
	}()
	select {
	case <-c.ctx.Done():
		c.done = true
		c.grpcConn.Close()
		c.inFlightWg.Wait()
	}
}

func (c *Client) handleGenericRequest(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	req, _ := json.Marshal(request)
	log.Logger.Debugf("Generic request: %s", req)
	switch request.Path {
	case tipRPC.PathGetAbsences:
		response, err = c.handleGetAbsenceForPeriod(request)
	case tipRPC.PathRegisterAbsences:
		response, err = c.handlePushUnauthorizedAbsence(request)
	case tipRPC.PathRegisterSickLeave:
		response, err = c.handleRegisterSickLeave(request)
	case tipRPC.PathGetSickLeaves:
		response, err = c.handleGetSickLeavesLastYear(request)
	case tipRPC.PathGetTeacherCourses:
		response, err = c.handleGetTodaysCoursesForTeacher(request)
	case tipRPC.PathAbsenceToSickLeave:
		response, err = c.handleAbsenceToSickLeave(request)
	default:
		log.Logger.Debugf("Unknown path '%s'", request.Path)
		response = tipRPC.Generic{
			Body:   []byte("unknown path"),
			Status: http.StatusBadRequest,
		}
		err = ErrorBadPath
	}

	response.MsgID = request.MsgID
	resp, _ := json.Marshal(response)
	log.Logger.Debugf("Response from FLYVO: %s", resp)
	return response, err
}

//SendGeneric - send a generic request
func (c *Client) SendGeneric(ctx context.Context, message tipRPC.Generic) (*tipRPC.Generic, error) {
	log.Logger.Debug("Sending generic message")
	if c.done {
		return nil, ErrorShuttingDown
	}
	c.inFlightWg.Add(1)
	defer c.inFlightWg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.tipClient.HandleGeneric(ctx, &message)
}

func (c *Client) PutEvent(ctx context.Context, message *tipRPC.Event) (*tipRPC.Generic, error) {
	eventjson, _ := json.Marshal(message)
	log.Logger.Debugf("PUT RPC: %s", eventjson)

	return c.tipClient.UpdateEvent(ctx, message)
}

func (c *Client) DeleteEvent(ctx context.Context, eventId string) (*tipRPC.Generic, error) {
	log.Logger.Debugf("DELETE RPC: %s", eventId)

	return c.tipClient.DeleteEvent(ctx, &tipRPC.String{Value: eventId})
}
func (c *Client) PostEvent(ctx context.Context, message *tipRPC.Event) (*tipRPC.Generic, error) {

	eventjson, _ := json.Marshal(message)
	log.Logger.Debugf("POST RPC: %s", eventjson)

	return c.tipClient.PublishEvent(ctx, message)
}
