package cloud

import (
 "context"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "sync"
 "testing"
 "time"

 "aether/event-fabric/internal/store"
)
func TestLocalStackPublish(t *testing.T){
 for _,corrupt:=range []bool{false,true}{
  var mu sync.Mutex
  var exists bool
  var saved []byte
  s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
   mu.Lock();defer mu.Unlock()
   body,_:=io.ReadAll(r.Body)
   if !strings.HasPrefix(r.Header.Get("Authorization"),"AWS4-HMAC-SHA256 Credential=test/20261128/us-east-1/s3/aws4_request,")||r.Header.Get("X-Amz-Content-Sha256")!=digest(body){t.Error("missing local request signature")}
   if r.URL.Path=="/"+bucket{
    if r.Method=="HEAD"{if !exists{w.WriteHeader(404)};return}
    exists=true;return
   }
   if r.Method=="PUT"{saved=body;return}
   if corrupt{_,_=w.Write([]byte("wrong"));return}
   _,_=w.Write(saved)
  }))
  a:=NewLocalStack();a.base=s.URL;a.now=func()time.Time{return time.Date(2026,11,28,0,0,0,0,time.UTC)}
  body:=[]byte("{\"status\":\"succeeded\"}\n")
  e:=store.Evidence{Key:"jobs/"+strings.Repeat("a",64)+".json",Body:body,SHA256:digest(body)}
  err:=a.Publish(context.Background(),e)
  if (err!=nil)!=corrupt{t.Fatalf("corrupt=%v err=%v",corrupt,err)}
  if !corrupt{if err:=a.Publish(context.Background(),e);err!=nil{t.Fatal(err)}}
  s.Close()
 }
}
func TestRejectMalformedEvidence(t *testing.T){
 a:=NewLocalStack()
 if err:=a.Publish(context.Background(),store.Evidence{Key:"../outside",Body:[]byte("x"),SHA256:digest([]byte("x"))});err==nil{t.Fatal("unsafe key accepted")}
}
