package lab
import (
 "crypto/tls"
 "testing"
 "time"
)
func TestTLSFailsClosed(t *testing.T){
 if _,err:=(TLSFiles{CA:"missing"}).Server();err==nil{t.Fatal("partial server TLS configuration accepted")}
 if _,err:=(TLSFiles{CA:"missing"}).Client();err==nil{t.Fatal("partial client TLS configuration accepted")}
 if _,err:=NewAdapterWithTLS("http://fixture:8083",time.Second,&tls.Config{MinVersion:tls.VersionTLS13});err==nil{t.Fatal("TLS downgrade accepted")}
}
