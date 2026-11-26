package lab

import (
	"context"
	"crypto/sha256"
	"fmt"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Adapter struct {
	base string
	client *http.Client
}

func NewAdapter(raw string)(*Adapter,error){
	u,err:=url.Parse(raw)
	if err!=nil || u.Host=="" || (u.Scheme!="http" && u.Scheme!="https") || u.User!=nil || u.RawQuery!="" || u.Fragment!=""{return nil,errors.New("invalid fixture URL")}
	return &Adapter{strings.TrimRight(raw,"/"),&http.Client{Timeout:2*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}},nil
}

// Repair writes the fixed desired content through the fixture API and
// independently reads it back. The target base URL is deployment-owned.
func (a *Adapter) Repair(ctx context.Context,id string)error{
	endpoint:=a.base+"/v1/fixtures/"+fmt.Sprintf("%x",sha256.Sum256([]byte(id)))
	req,err:=http.NewRequestWithContext(ctx,http.MethodPut,endpoint,nil)
	if err!=nil{return err}
	resp,err:=a.client.Do(req)
	if err!=nil{return err}
	resp.Body.Close()
	if resp.StatusCode!=204{return errors.New("fixture repair rejected")}
	req,err=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil)
	if err!=nil{return err}
	resp,err=a.client.Do(req)
	if err!=nil{return err}
	defer resp.Body.Close()
	b,err:=io.ReadAll(io.LimitReader(resp.Body,1024))
	if err!=nil || resp.StatusCode!=200 || string(b)!=desired{return errors.New("fixture verification failed")}
	return nil
}
