package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/google/cadvisor/utils/oomparser"
	logger "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"k8s.io/client-go/util/homedir"
)

// see https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/stats/cri_stats_provider.go#L900 for regexp structure context
var (
	log               *logger.Logger
	containerRegexpFS = regexp.MustCompile(`^\/kubepods\/(?:burstable\/)?pod([a-z0-9-]+)`)
	containerRegexpSD = regexp.MustCompile(`^\/kubepods.slice\/(?:kubepods-burstable.slice\/)kubepods-burstable-pod([a-z0-9_]+)`)
)

type oomEvent struct {
	PodID  string
	Parsed *oomparser.OomInstance
}

func main() {
	log = logger.New()
	log.SetReportCaller(true)
	log.SetLevel(logger.DebugLevel)
	log.SetFormatter(&logger.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	})

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	nodename := os.Getenv("NODENAME")
	if len(nodename) == 0 {
		log.Fatal("Missing environment variable: NODENAME")
	}

	client, err := initKubeClient(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	log.Print("Starting application")
	startTime := time.Now()

	oomLog, err := oomparser.New()
	if err != nil {
		panic(err.Error())
	}
	outStream := make(chan *oomparser.OomInstance, 10)
	go oomLog.StreamOoms(outStream)

	for event := range outStream {
		log.Debugf("raw oom event: %v", event)

		parsedContainerFS := containerRegexpFS.FindStringSubmatch(event.VictimContainerName)
		parsedContainerSD := containerRegexpSD.FindStringSubmatch(event.VictimContainerName)
		parsedContainer := []string{}

		if parsedContainerSD != nil {
			log.Debugf("using systemd cgroup path")
			// need to format _ ---> - first
			parsedContainerSD[1] = strings.ReplaceAll(parsedContainerSD[1], "_", "-")
			parsedContainer = parsedContainerSD
		} else {
			log.Debugf("not using systemd cgroup path")
			parsedContainer = parsedContainerFS
		}

		if parsedContainer != nil {
			if event.TimeOfDeath.Before(startTime) {
				log.Infof("historic oom, skipping: %s", parsedContainer[1])
				continue
			}
			oom := &oomEvent{
				PodID:  parsedContainer[1],
				Parsed: event,
			}
			processOomEvent(client, nodename, oom)
		}
	}
	log.Errorf("Unexpectedly stopped receiving OOM notifications")
}

func processOomEvent(client *kubernetes.Clientset, nodename string, event *oomEvent) {
	log.WithFields(logger.Fields{
		"pod_id":  event.PodID,
		"process": event.Parsed.ProcessName,
		"pid":     event.Parsed.Pid,
	}).Info("Parsed OOM event")

	pod, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodename),
	})
	if err != nil {
		panic(err.Error())
	}

	for _, p := range pod.Items {
		if string(p.UID) == event.PodID {
			emitEvent(client, event, p)
			break
		}
	}
}

func emitEvent(client *kubernetes.Clientset, event *oomEvent, pod v1.Pod) {

	ref, err := reference.GetReference(scheme.Scheme, &pod)
	if err != nil {
		panic(err.Error())
	}

	log.Println("this is my ref kind", ref.Kind, " and my ref labels", pod.Labels["app"])

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Debugf)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events(pod.Namespace)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "oomie"})
	msg := fmt.Sprintf("System OOM encountered, victim process: %s, pid: %d", event.Parsed.ProcessName, event.Parsed.Pid)
	recorder.AnnotatedEventf(ref, pod.Labels, v1.EventTypeWarning, "OOM", msg)
}

func initKubeClient(kubeconfig *string) (*kubernetes.Clientset, error) {
	// this returns a config object which configures both the token and TLS
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Info("unable to load in-cluster configuration, using KUBE_CONFIG location")
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
		kubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return kubeClient, nil
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return kubeClient, nil
}
