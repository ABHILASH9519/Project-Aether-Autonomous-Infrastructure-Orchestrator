package lab

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// ReconcileConfiguration invokes a deployment-owned, fixed command. No event
// data or request-controlled arguments reach the shell.
func ReconcileConfiguration(ctx context.Context) error {
	return exec.CommandContext(ctx,"/bin/bash","/lab/reconcile.sh").Run()
}

func VerifyConfiguration() error {
	const want="mode=healthy\nreplicas=1\n"
	for _,path:=range []string{"/tmp/aether-ansible/workload.conf","/tmp/aether-puppet/workload.conf"} {
		b,err:=os.ReadFile(path)
		if err!=nil || string(b)!=want{return errors.New("configuration verification failed")}
	}
	return nil
}
