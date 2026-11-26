package lab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRepairAndReplay(t *testing.T){
	dir:=t.TempDir()
	server:=httptest.NewServer(Handler(dir))
	defer server.Close()
	a,err:=NewAdapter(server.URL)
	if err!=nil{t.Fatal(err)}
	for i:=0;i<2;i++{if err=a.Repair(context.Background(),"synthetic/../id");err!=nil{t.Fatal(err)}}
	entries,err:=os.ReadDir(dir)
	if err!=nil || len(entries)!=1{t.Fatalf("fixture count=%d err=%v",len(entries),err)}
}
func TestVerificationFailure(t *testing.T){
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.Method=="PUT"{w.WriteHeader(204);return}
		_,_=w.Write([]byte("wrong contents"))
	}))
	defer s.Close()
	a,err:=NewAdapter(s.URL)
	if err!=nil{t.Fatal(err)}
	if a.Repair(context.Background(),"id")==nil{t.Fatal("invalid verification accepted")}
}
