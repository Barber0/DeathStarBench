package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
)

type ParamAgent struct {
	cliSet  *kubernetes.Clientset
	ipRank  int
	podName string
	ns      string
	svcName string

	paramMap map[string]map[string]interface{}
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
	if err = paramAgent.acquireParams(); err != nil {
		return nil, err
	}
	if err = paramAgent.setKernel(); err != nil {
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
	epCli := p.cliSet.CoreV1().Endpoints(p.ns)
	epObj, err := epCli.Get(ctx, p.svcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	podIp := os.Getenv(EnvMyPodIp)
	var ipArr []int64
	for _, subset := range epObj.Subsets {
		tmpIpArr := make([]int64, 0, len(subset.Addresses))
		var hasMyIp bool
		for _, addr := range subset.Addresses {
			hasMyIp = hasMyIp || addr.IP == podIp
			tmpIpArr = append(tmpIpArr, parseIp2Int(addr.IP))
		}
		if hasMyIp {
			ipArr = tmpIpArr
			break
		}
	}

	sort.Slice(ipArr, func(i, j int) bool {
		return ipArr[i] < ipArr[j]
	})

	podIpNum := parseIp2Int(podIp)
	for i, tmpIpNum := range ipArr {
		if podIpNum == tmpIpNum {
			p.ipRank = i
			break
		}
	}
	return nil
}

func (p *ParamAgent) acquireParams() error {
	paramCenterAddr := os.Getenv(EnvParamCenterAddr)
	paramCenterPort := os.Getenv(EnvParamCenterPort)
	if paramCenterAddr == "" || paramCenterPort == "" {
		return fmt.Errorf("invalid param center destination, addr: %v, port: %v",
			paramCenterAddr,
			paramCenterPort,
		)
	}

	pCenterAddr := "http://" + net.JoinHostPort(
		paramCenterAddr,
		paramCenterPort,
	)

	paramAcquireUrl := path.Join(pCenterAddr, PCenterPathAcquireParam)

	resp, err := http.DefaultClient.Get(paramAcquireUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBodyBts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var paramMap map[string]map[string]interface{}
	if err = json.Unmarshal(respBodyBts, &paramMap); err != nil {
		return err
	}

	p.paramMap = paramMap
	return nil
}

func (p *ParamAgent) setKernel() error {
	km, ok := p.paramMap[ParamTypeKernel]
	if !ok {
		return nil
	}
	for key, val := range km {
		cmd := exec.Command(
			"sysctl",
			"-w",
			fmt.Sprintf("%s=%v", key, val),
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	cmd := exec.Command("sysctl", "-p")
	return cmd.Run()
}

func parseIp2Int(podIp string) int64 {
	return big.NewInt(0).SetBytes(net.ParseIP(podIp).To4()).Int64()
}
