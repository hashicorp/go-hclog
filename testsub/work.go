package testsub

import "github.com/hashicorp/go-hclog"

func DoWork(log hclog.Logger) {
	log.Info("this is test", "who", "programmer", "why", "testing is fun")
}
