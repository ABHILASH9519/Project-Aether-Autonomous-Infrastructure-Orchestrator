package main

import ("testing";"time")

func TestPercentile(t *testing.T){
	s:=[]time.Duration{time.Millisecond,2*time.Millisecond,3*time.Millisecond,4*time.Millisecond,5*time.Millisecond}
	for _,tc:=range []struct{p int; want float64}{{50,3},{95,5},{99,5},{100,5}}{
		if got:=percentile(s,tc.p);got!=tc.want{t.Fatalf("p%d got %f want %f",tc.p,got,tc.want)}
	}
}
