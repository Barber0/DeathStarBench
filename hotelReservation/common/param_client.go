package common

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"sort"
	"time"
)

type ParamAgent struct {
	cliSet  *kubernetes.Clientset
	IpRank  int
	podName string
	ns      string
	svcName string

	ConfMap map[string]interface{}
}

func NewParamAgent(svcName string) (*ParamAgent, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cliSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	paramAgent := &ParamAgent{
		cliSet:  cliSet,
		svcName: svcName,
	}
	if err = paramAgent.selfAware(context.TODO()); err != nil {
		return nil, err
	}
	if err = paramAgent.setKernel(); err != nil {
		return nil, err
	}
	if err = paramAgent.loadConfig(); err != nil {
		return nil, err
	}
	return paramAgent, nil
}

func (p *ParamAgent) selfAware(ctx context.Context) error {
	p.ns = os.Getenv(EnvMyPodNs)
	if p.ns == "" {
		p.ns = corev1.NamespaceDefault
	}
	p.podName = os.Getenv(EnvMyPodName)

	podIp := os.Getenv(EnvMyPodIp)
	var ipArr []int64

	fmt.Println("podIP: ", podIp)

SelfAwareRetry:
	for stopRetry := false; !stopRetry; {
		epCli := p.cliSet.CoreV1().Endpoints(p.ns)
		epObj, err := epCli.Get(ctx, p.svcName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, subset := range epObj.Subsets {
			tmpIpArr := make([]int64, 0, len(subset.Addresses))
			var hasMyIp bool
			for _, addr := range subset.Addresses {
				fmt.Printf("hasMyIp: %v, ip: %v\n", hasMyIp, addr.IP)
				hasMyIp = hasMyIp || addr.IP == podIp
				tmpIpArr = append(tmpIpArr, parseIp2Int(addr.IP))
			}
			if hasMyIp {
				ipArr = tmpIpArr
				break SelfAwareRetry
			}
		}
		time.Sleep(3 * time.Second)
	}

	sort.Slice(ipArr, func(i, j int) bool {
		return ipArr[i] < ipArr[j]
	})

	podIpNum := parseIp2Int(podIp)
	for i, tmpIpNum := range ipArr {
		if podIpNum == tmpIpNum {
			p.IpRank = i
			break
		}
	}
	return nil
}

func (p *ParamAgent) setKernel() error {

	kernelMapEnv := fmt.Sprintf(ParamEnvTplKernel, p.svcName, p.IpRank)
	kernelMapStr := os.Getenv(kernelMapEnv)
	if kernelMapStr == "" {
		log.Println("kernel params not found")
		return nil
	}

	var kernelMap map[string]interface{}
	if err := json.Unmarshal([]byte(kernelMapStr), &kernelMap); err != nil {
		return err
	}

	for key, val := range kernelMap {
		if err := SetSysctl(key, val.(string)); err != nil {
			return err
		}
	}
	cmd := exec.Command("sysctl", "-p")
	return cmd.Run()
}

func (p *ParamAgent) loadConfig() error {
	kernelMapEnv := fmt.Sprintf(ParamEnvTplConfig, p.svcName, p.IpRank)
	kernelMapStr := os.Getenv(kernelMapEnv)
	if kernelMapStr == "" {
		log.Println("config params not found")
		return nil
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(kernelMapStr), &configMap); err != nil {
		return err
	}

	p.ConfMap = configMap
	return nil
}

func parseIp2Int(podIp string) int64 {
	return big.NewInt(0).SetBytes(net.ParseIP(podIp).To4()).Int64()
}
