package rpc

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	model "github.com/tktip/flyvo-api/pkg/flyvo"
	tipRPC "github.com/tktip/flyvo-api/pkg/rpc"
	"github.com/tktip/flyvo-rpc-client/internal/log"
)

const cTypeJson = "application/json"

const (
	defaultAddress                  = "localhost:50051"
	convertAbsenceToSickLeave       = "/absence"
	getCoursesEndpoint              = "/getoverview/"
	registerUnauthorizedAbsence     = "/absence"
	getUnauthorizedAbsencesEndpoint = "/getinvalidabsenceforperson"
	registerSickLeave               = "/selfcertification"
	getSickLeaves                   = "/getselfcertificationoverview"
)

func (c *Client) doHTTPToFlyVo(method string, url string, body io.Reader) ([]byte, int, error) {
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	log.Logger.Debugf("Sending request to '%s'", url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, -1, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return bod, resp.StatusCode, err
}

func (c *Client) handlePushUnauthorizedAbsence(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	url := c.FlyvoApiEndpoints.RootAddress + registerUnauthorizedAbsence
	cont, status, err := c.doHTTPToFlyVo(http.MethodPost, url, bytes.NewReader(request.Body))
	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Headers: map[string]string{"path": request.Path},
		MsgID:   request.MsgID,
		Status:  int32(status),
		Body:    cont,
	}, nil
}

func (c *Client) handleGetAbsenceForPeriod(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	url := c.FlyvoApiEndpoints.RootAddress + getUnauthorizedAbsencesEndpoint
	cont, status, err := c.doHTTPToFlyVo(http.MethodGet, url, bytes.NewReader(request.Body))
	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Headers: map[string]string{"path": request.Path},
		MsgID:   request.MsgID,
		Status:  int32(status),
		Body:    cont,
	}, nil
}

func (c *Client) handleGetSickLeavesLastYear(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	gsl := model.GetSickLeavesRequest{}
	err = json.Unmarshal(request.Body, &gsl)
	if err != nil {
		return tipRPC.Generic{
			Body:   []byte(err.Error()),
			Status: http.StatusUnprocessableEntity,
		}, err
	}

	url := c.FlyvoApiEndpoints.RootAddress + getSickLeaves + "/" + gsl.VismaID + "/" + gsl.ToDate
	cont, status, err := c.doHTTPToFlyVo(http.MethodGet, url, nil)
	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Headers: map[string]string{"path": request.Path},
		MsgID:   request.MsgID,
		Status:  int32(status),
		Body:    cont,
	}, nil
}

func (c *Client) handleRegisterSickLeave(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	url := c.FlyvoApiEndpoints.RootAddress + registerSickLeave
	cont, status, err := c.doHTTPToFlyVo(http.MethodPost, url, bytes.NewReader(request.Body))
	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Headers: map[string]string{"path": request.Path},
		MsgID:   request.MsgID,
		Status:  int32(status),
		Body:    cont,
	}, nil
}

func (c *Client) handleGetTodaysCoursesForTeacher(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	gcr := model.GetCoursesRequest{}
	err = json.Unmarshal(request.Body, &gcr)
	if err != nil {
		return tipRPC.Generic{
			Body:   []byte(err.Error()),
			Status: http.StatusUnprocessableEntity,
		}, err
	}

	from := gcr.FromDate.Format("02012006")
	to := gcr.ToDate.Format("02012006")

	url := c.FlyvoApiEndpoints.RootAddress + getCoursesEndpoint + from + "/" + to
	cont, status, err := c.doHTTPToFlyVo(http.MethodGet, url, nil)

	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Path:   request.Path,
		MsgID:  request.MsgID,
		Status: int32(status),
		Body:   cont,
	}, nil
}

func (c *Client) handleAbsenceToSickLeave(request tipRPC.Generic) (response tipRPC.Generic, err error) {
	url := c.FlyvoApiEndpoints.RootAddress + convertAbsenceToSickLeave
	cont, status, err := c.doHTTPToFlyVo(http.MethodPost, url, bytes.NewReader(request.Body))
	if err != nil {
		return tipRPC.Generic{Body: []byte(err.Error()), Status: http.StatusInternalServerError}, err
	}

	return tipRPC.Generic{
		Headers: map[string]string{"path": request.Path},
		MsgID:   request.MsgID,
		Status:  int32(status),
		Body:    cont,
	}, nil
}
