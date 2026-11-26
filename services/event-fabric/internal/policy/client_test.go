package policy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecision(t *testing.T){
	for _,tc:=range []struct{body string; status int; allowed,failed bool}{
		{`{"result":true}`,200,true,false},
		{`{"result":false}`,200,false,false},
		{`{}`,200,false,true},
		{`{"result":"true"}`,200,false,true},
		{`{"result":true} {}`,200,false,true},
		{`{"result":true}`,503,false,true},
	}{
		s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
			if r.URL.Path!="/v1/data/aether/workflow/allow" || r.Method!="POST"{t.Error("unexpected policy route")}
			w.WriteHeader(tc.status)
			_,_=w.Write([]byte(tc.body))
		}))
		c,err:=New(s.URL)
		if err!=nil{t.Fatal(err)}
		got,err:=c.Allow(context.Background(),"operator","acknowledge")
		s.Close()
		if got!=tc.allowed || (err!=nil)!=tc.failed{t.Fatalf("%+v: %v %v",tc,got,err)}
	}
}
