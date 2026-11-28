package lab

import (
 "crypto/tls"
 "crypto/x509"
 "errors"
 "os"
)

type TLSFiles struct{ CA,Cert,Key string }

func (f TLSFiles) Enabled() bool{return f.CA!=""||f.Cert!=""||f.Key!=""}

func (f TLSFiles) material()(*x509.CertPool,tls.Certificate,error){
 var empty tls.Certificate
 if f.CA==""||f.Cert==""||f.Key==""{return nil,empty,errors.New("all fixture TLS files are required")}
 pem,err:=os.ReadFile(f.CA);if err!=nil{return nil,empty,err}
 pool:=x509.NewCertPool()
 if !pool.AppendCertsFromPEM(pem){return nil,empty,errors.New("invalid fixture CA")}
 cert,err:=tls.LoadX509KeyPair(f.Cert,f.Key)
 return pool,cert,err
}
func (f TLSFiles) Server()(*tls.Config,error){
 pool,cert,err:=f.material();if err!=nil{return nil,err}
 return &tls.Config{MinVersion:tls.VersionTLS13,Certificates:[]tls.Certificate{cert},ClientCAs:pool,ClientAuth:tls.RequireAndVerifyClientCert,
  VerifyConnection:func(c tls.ConnectionState)error{
   if len(c.PeerCertificates)==0{return errors.New("missing executor identity")}
   for _,u:=range c.PeerCertificates[0].URIs{if u.String()=="spiffe://aether.local/executor"{return nil}}
   return errors.New("unexpected fixture client identity")
  }},nil
}
func(f TLSFiles) Client()(*tls.Config,error){
 pool,cert,err:=f.material();if err!=nil{return nil,err}
 return &tls.Config{MinVersion:tls.VersionTLS13,RootCAs:pool,Certificates:[]tls.Certificate{cert}},nil
}
func TLSFromEnv()TLSFiles{return TLSFiles{CA:os.Getenv("LAB_TLS_CA"),Cert:os.Getenv("LAB_TLS_CERT"),Key:os.Getenv("LAB_TLS_KEY")}}
