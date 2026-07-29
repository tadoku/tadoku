package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type imageUpdater struct {
	Status struct {
		LastCheckedAt string `json:"lastCheckedAt"`
		Conditions    []struct {
			Type    string `json:"type"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"conditions"`
	} `json:"status"`
}

type kubeClient struct {
	client *http.Client
	url    string
	token  string
}

type imageUpdaterReader interface {
	get(context.Context) (*imageUpdater, error)
}

func newKubeClient(namespace, name string) (*kubeClient, error) {
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}
	caBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read Kubernetes CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("parse Kubernetes CA")
	}
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if host == "" || port == "" {
		return nil, errors.New("Kubernetes service environment is missing")
	}
	apiURL := fmt.Sprintf(
		"https://%s:%s/apis/argocd-image-updater.argoproj.io/v1alpha1/namespaces/%s/imageupdaters/%s",
		host, port, url.PathEscape(namespace), url.PathEscape(name),
	)
	return &kubeClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					RootCAs:    pool,
				},
			},
		},
		url:   apiURL,
		token: strings.TrimSpace(string(tokenBytes)),
	}, nil
}

func (k *kubeClient) get(ctx context.Context) (*imageUpdater, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+k.token)
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Kubernetes API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var resource imageUpdater
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}
	return &resource, nil
}
