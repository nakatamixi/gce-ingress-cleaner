package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"encoding/json"
)

type TargetIngressAnnotations struct {
	Backends            string `json:"ingress.kubernetes.io/backends"`
	ForwardingRule      string `json:"ingress.kubernetes.io/forwarding-rule"`
	HTTPSForwardingRule string `json:"ingress.kubernetes.io/https-forwarding-rule"`
	HTTPTargetProxy     string `json:"ingress.kubernetes.io/target-proxy"`
	HTTPSTargetProxy    string `json:"ingress.kubernetes.io/https-target-proxy"`
	URLMap              string `json:"ingress.kubernetes.io/url-map"`
}

func main() {
	var (
		ingress   string
		namespace string
	)
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&ingress, "i", "", "ingress name")
	flags.StringVar(&namespace, "n", "", "namespace of ingress")
	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return
	}
	a, err := readManifest(ingress, namespace)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("clean command is")
	fmt.Printf("kubectl delete ing %s -n %s\n", ingress, namespace)
	printfIfNotEmpty("gcloud compute forwarding-rules delete %s --global -q\n", a.ForwardingRule)
	printfIfNotEmpty("gcloud compute forwarding-rules delete %s --global -q\n", a.HTTPSForwardingRule)
	printfIfNotEmpty("gcloud compute target-http-proxies delete %s -q\n", a.HTTPTargetProxy)
	printfIfNotEmpty("gcloud compute target-https-proxies delete %s -q\n", a.HTTPSTargetProxy)
	printfIfNotEmpty("gcloud compute url-maps delete %s -q\n", a.URLMap)
	backends, err := getNotNodePortSvcBackends(a.Backends)
	if err != nil {
		log.Fatal(err)
	}
	for _, be := range backends {
		printfIfNotEmpty("gcloud compute backend-services delete %s --global -q\n", be)
		printfIfNotEmpty("gcloud compute health-checks delete %s -q\n", be)
	}

}

func readManifest(ingress, namespace string) (*TargetIngressAnnotations, error) {
	var m struct {
		Metadata struct {
			Annotations TargetIngressAnnotations
		}
	}
	out, err := exec.Command("kubectl", "get", "ing", ingress, "-n", namespace, "-o", "json").Output()
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(out, &m)
	if err != nil {
		return nil, err
	}
	return &m.Metadata.Annotations, nil

}

func printfIfNotEmpty(format, s string) {
	if s != "" {
		fmt.Printf(format, s)
	}
}

func getNotNodePortSvcBackends(be string) ([]string, error) {
	var m map[string]interface{}
	var r []string
	if err := json.Unmarshal([]byte(be), &m); err != nil {
		return r, err
	}

	for k, _ := range m {
		// k8s-be- is shared NodePort. do not delete it.
		if strings.HasPrefix(k, "k8s-be-") {
			continue
		}
		r = append(r, k)
	}
	return r, nil
}
