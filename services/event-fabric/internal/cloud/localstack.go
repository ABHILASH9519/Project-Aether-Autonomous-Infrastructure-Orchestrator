// Package cloud implements a deliberately emulator-only S3 adapter.
// Its fixed dummy credentials and endpoint must never be used for AWS.
package cloud

import (
 "bytes"
 "context"
 "crypto/hmac"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "net/http"
 "regexp"
 "strings"
 "time"

 "aether/event-fabric/internal/store"
)

const bucket="aether-lab-evidence"
var keyPattern=regexp.MustCompile(`^jobs/[a-f0-9]{64}\.json$`)

type LocalStack struct{base string; client *http.Client; now func()time.Time}
func NewLocalStack()*LocalStack{
 return &LocalStack{base:"http://localstack:4566",client:&http.Client{Timeout:4*time.Second,Transport:&http.Transport{MaxIdleConns:4,IdleConnTimeout:30*time.Second},CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}},now:time.Now}
}
func digest(b []byte)string{h:=sha256.Sum256(b);return hex.EncodeToString(h[:])}
func mac(key []byte,value string)[]byte{h:=hmac.New(sha256.New,key);_,_=h.Write([]byte(value));return h.Sum(nil)}

// sign is limited to path-style S3 requests against the local emulator:
// no query, session token, user-selected endpoint or real AWS credentials.
func sign(r *http.Request,body []byte,now time.Time){
 date:=now.UTC().Format("20060102")
 stamp:=now.UTC().Format("20060102T150405Z")
 hash:=digest(body)
 r.Header.Set("X-Amz-Date",stamp)
 r.Header.Set("X-Amz-Content-Sha256",hash)
 signed:="host;x-amz-content-sha256;x-amz-date"
 canonical:=strings.Join([]string{r.Method,r.URL.EscapedPath(),"","host:"+r.URL.Host+"\nx-amz-content-sha256:"+hash+"\nx-amz-date:"+stamp+"\n",signed,hash},"\n")
 scope:=date+"/us-east-1/s3/aws4_request"
 toSign:="AWS4-HMAC-SHA256\n"+stamp+"\n"+scope+"\n"+digest([]byte(canonical))
 key:=mac(mac(mac(mac([]byte("AWS4test"),date),"us-east-1"),"s3"),"aws4_request")
 r.Header.Set("Authorization","AWS4-HMAC-SHA256 Credential=test/"+scope+", SignedHeaders="+signed+", Signature="+hex.EncodeToString(mac(key,toSign)))
}
func (a *LocalStack) request(ctx context.Context,method,path string,body []byte)(*http.Response,error){
 req,err:=http.NewRequestWithContext(ctx,method,a.base+path,bytes.NewReader(body))
 if err!=nil{return nil,err}
 if len(body)>0{req.Header.Set("Content-Type","application/json")}
 sign(req,body,a.now())
 return a.client.Do(req)
}
func closeResponse(r *http.Response){_,_=io.Copy(io.Discard,io.LimitReader(r.Body,4096));_ = r.Body.Close()}
func (a *LocalStack) ensureBucket(ctx context.Context)error{
 path:="/"+bucket
 r,err:=a.request(ctx,http.MethodHead,path,nil)
 if err!=nil{return err}
 status:=r.StatusCode;closeResponse(r)
 if status==200{return nil}
 if status!=404{return fmt.Errorf("local bucket lookup HTTP %d",status)}
 r,err=a.request(ctx,http.MethodPut,path,nil)
 if err!=nil{return err}
 status=r.StatusCode;closeResponse(r)
 if status==200||status==204{return nil}
 if status==409{
  r,err=a.request(ctx,http.MethodHead,path,nil)
  if err!=nil{return err}
  status=r.StatusCode;closeResponse(r)
  if status==200{return nil}
 }
 return fmt.Errorf("local bucket creation HTTP %d",status)
}
// Publish writes a fixed synthetic job object and verifies its bytes.
func (a *LocalStack) Publish(ctx context.Context,e store.Evidence)error{
 if !keyPattern.MatchString(e.Key)||len(e.Body)==0||len(e.Body)>64*1024||digest(e.Body)!=e.SHA256{return errors.New("invalid evidence")}
 ctx,cancel:=context.WithTimeout(ctx,12*time.Second);defer cancel()
 if err:=a.ensureBucket(ctx);err!=nil{return err}
 path:="/"+bucket+"/"+e.Key
 r,err:=a.request(ctx,http.MethodPut,path,e.Body)
 if err!=nil{return err}
 status:=r.StatusCode;closeResponse(r)
 if status!=200&&status!=204{return fmt.Errorf("local evidence upload HTTP %d",status)}
 r,err=a.request(ctx,http.MethodGet,path,nil)
 if err!=nil{return err}
 defer r.Body.Close()
 if r.StatusCode!=200{return fmt.Errorf("local evidence readback HTTP %d",r.StatusCode)}
 b,err:=io.ReadAll(io.LimitReader(r.Body,int64(len(e.Body)+1)))
 if err!=nil{return err}
 if !bytes.Equal(b,e.Body){return errors.New("local evidence readback mismatch")}
 return nil
}
