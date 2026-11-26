package workflow

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"aether/event-fabric/internal/store"
)

type fakeRepo struct { calls int; err error }
func (f *fakeRepo) Incident(context.Context,string)(store.Incident,error){return store.Incident{},f.err}
func (f *fakeRepo) Transition(context.Context,string,int64,string,string)(store.Incident,error){f.calls++;return store.Incident{},f.err}

type testGate struct{}
func (testGate) Allow(_ context.Context,role,action string)(bool,error){return role=="operator" && action=="acknowledge",nil}

func TestOperatorGate(t *testing.T){
	const token="test-only-operator-token-123456"
	for _,tc:=range []struct{auth,body string; code,calls int; err error}{
		{"",`{"expected_version":0,"action":"acknowledge"}`,401,0,nil},
		{token,`{"action":"acknowledge"}`,400,0,nil},
		{token,`{"expected_version":0,"action":"acknowledge"}`,200,1,nil},
		{token,`{"expected_version":0,"action":"approve"}`,403,0,nil},
		{token,`{"expected_version":0,"action":"acknowledge"}`,409,1,store.ErrVersion},
	}{
		f:=&fakeRepo{err:tc.err}
		h,err:=Handler(f,token,"test-only-approver-token-654321",testGate{})
		if err!=nil{t.Fatal(err)}
		req:=httptest.NewRequest("POST","/api/v1/incidents/test/transitions",strings.NewReader(tc.body))
		if tc.auth!=""{req.Header.Set("Authorization","Bearer "+tc.auth)}
		w:=httptest.NewRecorder()
		h.ServeHTTP(w,req)
		if w.Code!=tc.code || f.calls!=tc.calls{t.Fatalf("status=%d calls=%d want %d/%d",w.Code,f.calls,tc.code,tc.calls)}
	}
}

type unavailableGate struct{}
func (unavailableGate) Allow(context.Context,string,string)(bool,error){return false,errors.New("OPA offline")}

func TestUnavailablePolicyCannotMutate(t *testing.T){
	const operator="test-only-operator-token-123456"
	const approver="test-only-approver-token-654321"
	f:=&fakeRepo{}
	h,err:=Handler(f,operator,approver,unavailableGate{})
	if err!=nil{t.Fatal(err)}
	req:=httptest.NewRequest("POST","/api/v1/incidents/test/transitions",strings.NewReader(`{"expected_version":0,"action":"acknowledge"}`))
	req.Header.Set("Authorization","Bearer "+operator)
	w:=httptest.NewRecorder()
	h.ServeHTTP(w,req)
	if w.Code!=503 || f.calls!=0{t.Fatalf("status=%d calls=%d",w.Code,f.calls)}
}

func TestIdenticalCredentialsRejected(t *testing.T){
	const token="test-only-operator-token-123456"
	if _,err:=Handler(&fakeRepo{},token,token,testGate{});err==nil{t.Fatal("identical tokens accepted")}
}
