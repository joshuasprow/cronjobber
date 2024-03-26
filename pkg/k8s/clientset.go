package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	if kubeconfig == "" {
		godotenv.Load()

		kc := os.Getenv("KUBECONFIG")

		if kc == "" {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("get user home dir: %w", err)
			}

			kc = filepath.Join(homedir, ".kube", "config")
		}

		kubeconfig = kc
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
