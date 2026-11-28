// Command aetherctl is a local-only operator demo for the synthetic lab.
package main

import (
 "bytes"
 "context"
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "io"
 "net/http"
 "os"
 "time"
)

const ingestURL="http://127.0.0.1:8081"
const workflowURL="http://127.0.0.1:8082"

type incident struct{
 State string `json:"state"`
 Version int64 `json:"version"`
 EventCount int64 `json:"event_count"`
}
type client struct{ http *http.Client; operator,approver string }
func (c client) request(ctx context.Context,method,url,token string,body,out any,want int)error{
 var b io.Reader
 if body!=nil{
  raw,err:=json.Marshal(body);if err!=nil{return err};b=bytes.NewReader(raw)
 }
 req,err:=http.NewRequestWithContext(ctx,method,url,b);if err!=nil{return err}
 if body!=nil{req.Header.Set("Content-Type","application/json")}
 if token!=""{req.Header.Set("Authorization","Bearer "+token)}
 res,err:=c.http.Do(req);if err!=nil{return err};defer res.Body.Close()
 if res.StatusCode!=want{return fmt.Errorf("%s returned HTTP %d (expected %d)",method,res.StatusCode,want)}
 if out!=nil{return json.NewDecoder(io.LimitReader(res.Body,1<<20)).Decode(out)}
 _,err=io.Copy(io.Discard,io.LimitReader(res.Body,4096));return err
}
func (c client) wait(ctx context.Context,url,state string,version int64)error{
 for {
  var i incident
  if err:=c.request(ctx,"GET",url,c.operator,nil,&i,200);err==nil{
   if i.State==state&&i.Version==version&&i.EventCount==1{return nil}
   if i.State=="failed"{return errors.New("synthetic remediation failed")}
   if i.Version>version{return errors.New("unexpected incident version")}
  }
  select{case <-ctx.Done():return ctx.Err();case <-time.After(500*time.Millisecond):}
 }
}
func (c client) demo(ctx context.Context,output io.Writer)error{
 var nonce [12]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{return err}
 id:="demo-"+hex.EncodeToString(nonce[:])
 url:=workflowURL+"/api/v1/incidents/"+id
 fmt.Fprintf(output,"AETHER | synthetic incident %s\n",id)
 envelope:=map[string]any{"event_id":id,"event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":time.Now().UTC(),"source":"aetherctl-demo","correlation_id":id,"payload":map[string]any{}}
 if err:=c.request(ctx,"POST",ingestURL+"/api/v1/events","",envelope,nil,202);err!=nil{return err}
 if err:=c.wait(ctx,url,"open",0);err!=nil{return err}
 fmt.Fprintln(output,"[PASS] Kafka ingestion and PostgreSQL projection")
 if err:=c.request(ctx,"POST",url+"/transitions",c.operator,map[string]any{"expected_version":0,"action":"execute_lab"},nil,409);err!=nil{return err}
 fmt.Fprintln(output,"[PASS] Execution blocked before approval")
 for _,step:=range []struct{v int64;action,state,role string}{
  {0,"acknowledge","acknowledged","operator"},
  {1,"request_approval","awaiting_approval","operator"},
  {2,"approve","approved","approver"},
  {3,"execute_lab","executing","operator"},
 }{
  token:=c.operator;if step.role=="approver"{token=c.approver}
  var i incident
  if err:=c.request(ctx,"POST",url+"/transitions",token,map[string]any{"expected_version":step.v,"action":step.action},&i,200);err!=nil{return err}
  if i.State!=step.state||i.Version!=step.v+1{return errors.New("unexpected transition response")}
  fmt.Fprintf(output,"[PASS] %-8s %-18s -> %s (v%d)\n",step.role,step.action,i.State,i.Version)
 }
 if err:=c.wait(ctx,url,"resolved",6);err!=nil{return err}
 var job struct{Status string `json:"status"`; Evidence string `json:"evidence"`;Attempts int `json:"attempts"`}
 if err:=c.request(ctx,"GET",url+"/job",c.operator,nil,&job,200);err!=nil{return err}
 if job.Status!="succeeded"||job.Evidence!="fixture content verified"||job.Attempts<1||job.Attempts>3{return errors.New("unexpected execution evidence")}
 var audit struct{Entries []struct{Version int64 `json:"version"`;Action string `json:"action"`;Role string `json:"actor_role"`;State string `json:"next_state"`} `json:"entries"`}
 if err:=c.request(ctx,"GET",url+"/audit",c.operator,nil,&audit,200);err!=nil{return err}
 roles:=[]string{"operator","operator","approver","operator","worker","worker"}
 if len(audit.Entries)!=len(roles){return errors.New("incomplete audit")}
 for n,e:=range audit.Entries{
  if e.Version!=int64(n+1)||e.Role!=roles[n]{return errors.New("unexpected audit ordering or role")}
 }
 fmt.Fprintf(output,"[PASS] RESOLVED | attempts=%d | %s\n",job.Attempts,job.Evidence)
 fmt.Fprintln(output,"AUDIT  VERSION  ROLE      ACTION              NEXT STATE")
 for _,e:=range audit.Entries{fmt.Fprintf(output,"       %-8d %-9s %-19s %s\n",e.Version,e.Role,e.Action,e.State)}
 return nil
}
func main(){
 if len(os.Args)<2||os.Args[1]!="demo"{fmt.Fprintln(os.Stderr,"usage: aetherctl demo -approve-lab");os.Exit(2)}
 flags:=flag.NewFlagSet("demo",flag.ExitOnError)
 approve:=flags.Bool("approve-lab",false,"explicitly authorize synthetic fixture repair using both lab roles")
 _=flags.Parse(os.Args[2:])
 operator,approver:=os.Getenv("AETHER_OPERATOR_TOKEN"),os.Getenv("AETHER_APPROVER_TOKEN")
 if !*approve||len(operator)<24||len(approver)<24||operator==approver||flags.NArg()!=0{fmt.Fprintln(os.Stderr,"demo requires -approve-lab and distinct local role tokens");os.Exit(2)}
 c:=client{http:&http.Client{Timeout:8*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}},operator:operator,approver:approver}
 ctx,cancel:=context.WithTimeout(context.Background(),120*time.Second);defer cancel()
 if err:=c.demo(ctx,os.Stdout);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
}
