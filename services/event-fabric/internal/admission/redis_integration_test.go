package admission

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisBudget(t *testing.T) {
	addr:=os.Getenv("AETHER_TEST_REDIS_ADDR")
	if addr=="" { t.Skip("requires disposable Redis") }
	client:=redis.NewClient(&redis.Options{Addr:addr,DB:15})
	defer client.Close()
	ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	const key="aether:ingest:budget:v1"
	if err:=client.Del(ctx,key).Err();err!=nil {t.Fatal(err)}
	defer client.Del(ctx,key)
	g:=RedisGate{Client:client,Limit:2}
	for _,want:=range []bool{true,true,false}{
		got,err:=g.Allow(ctx)
		if err!=nil || got!=want{t.Fatalf("allow=%v err=%v want=%v",got,err,want)}
	}
	ttl,err:=client.PTTL(ctx,key).Result()
	if err!=nil || ttl<=0 || ttl>time.Minute{t.Fatalf("ttl=%v err=%v",ttl,err)}
}
