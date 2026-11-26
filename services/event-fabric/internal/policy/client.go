// Package policy implements a fail-closed client for the OPA decision API.
package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Gate interface {
	Allow(context.Context, string, string) (bool, error)
}

type Client struct {
	endpoint string
	http *http.Client
}

func New(raw string) (*Client,error) {
	u,err:=url.Parse(raw)
	if err!=nil || u.Host=="" || (u.Scheme!="http" && u.Scheme!="https") || u.User!=nil || u.RawQuery!="" || u.Fragment!="" {
		return nil,errors.New("invalid OPA URL")
	}
	return &Client{
		endpoint:strings.TrimRight(raw,"/")+"/v1/data/aether/workflow/allow",
		http:&http.Client{Timeout:2*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}},
	},nil
}

func (c *Client) Allow(ctx context.Context,role,action string)(bool,error) {
	body,err:=json.Marshal(map[string]any{"input":map[string]string{"role":role,"action":action,"mode":"simulation"}})
	if err!=nil{return false,err}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.endpoint,bytes.NewReader(body))
	if err!=nil{return false,err}
	req.Header.Set("Content-Type","application/json")
	resp,err:=c.http.Do(req)
	if err!=nil{return false,err}
	defer resp.Body.Close()
	if resp.StatusCode!=http.StatusOK{return false,errors.New("policy service returned non-200")}
	var decision struct{ Result *bool `json:"result"` }
	d:=json.NewDecoder(io.LimitReader(resp.Body,8192))
	if d.Decode(&decision)!=nil || decision.Result==nil || d.Decode(new(any))!=io.EOF{
		return false,errors.New("invalid or undefined policy decision")
	}
	return *decision.Result,nil
}
