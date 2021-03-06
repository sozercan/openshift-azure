package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"

	"github.com/satori/uuid"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/client-go/tools/clientcmd/api/v1"

	acsapi "github.com/openshift/openshift-azure/pkg/api"
	"github.com/openshift/openshift-azure/pkg/tls"
)

func selectNodeImage(cs *acsapi.ContainerService, c *Config) {
	c.ImagePublisher = "redhat"
	c.ImageOffer = "osa-preview"
	c.ImageVersion = "latest"

	switch os.Getenv("DEPLOY_OS") {
	case "":
		c.ImageSKU = "osa_" + strings.Replace(cs.Properties.OrchestratorProfile.OrchestratorVersion[1:], ".", "", -1)
	case "centos7":
		c.ImageSKU = "origin_" + strings.Replace(cs.Properties.OrchestratorProfile.OrchestratorVersion[1:], ".", "", -1)
	}

	c.ImageResourceGroup = os.Getenv("IMAGE_RESOURCEGROUP")
	c.ImageResourceName = os.Getenv("IMAGE_RESOURCENAME")
}

func image(imageConfigFormat, component, version string) string {
	image := strings.Replace(imageConfigFormat, "${component}", component, -1)
	return strings.Replace(image, "${version}", version, -1)
}

func selectContainerImagesOrigin(cs *acsapi.ContainerService, c *Config) {
	c.ImageConfigFormat = os.Getenv("OREG_URL")
	if c.ImageConfigFormat == "" {
		c.ImageConfigFormat = "docker.io/openshift/origin-${component}:${version}"
	}

	switch cs.Properties.OrchestratorProfile.OrchestratorVersion {
	case "v3.10":
		v := "v3.10.0"
		c.ControlPlaneImage = image(c.ImageConfigFormat, "control-plane", v)
		c.NodeImage = image(c.ImageConfigFormat, "node", v)
		c.ServiceCatalogImage = image(c.ImageConfigFormat, "service-catalog", v)
		c.TemplateServiceBrokerImage = image(c.ImageConfigFormat, "template-service-broker", v)
		c.RegistryImage = image(c.ImageConfigFormat, "docker-registry", v)
		c.RouterImage = image(c.ImageConfigFormat, "haproxy-router", v)
		c.WebConsoleImage = image(c.ImageConfigFormat, "web-console", v)

		c.MasterEtcdImage = "quay.io/coreos/etcd:v3.2.15"

		c.OAuthProxyImage = "docker.io/openshift/oauth-proxy:v1.0.0"
		c.PrometheusImage = "docker.io/openshift/prometheus:v2.2.1"
		c.PrometheusAlertBufferImage = "docker.io/openshift/prometheus-alert-buffer:v0.0.2"
		c.PrometheusAlertManagerImage = "docker.io/openshift/prometheus-alertmanager:v0.14.0"
		c.PrometheusNodeExporterImage = "docker.io/openshift/prometheus-node-exporter:v0.15.2"

		c.AnsibleServiceBrokerImage = "docker.io/ansibleplaybookbundle/origin-ansible-service-broker:latest"

		c.RegistryConsoleImage = "docker.io/cockpit/kubernetes:latest"

		c.AzureCLIImage = "docker.io/microsoft/azure-cli:latest"

		c.SyncImage = "quay.io/openshift-on-azure/sync:latest"
	}
}

func selectContainerImagesOSA(cs *acsapi.ContainerService, c *Config) {
	c.ImageConfigFormat = os.Getenv("OREG_URL")
	if c.ImageConfigFormat == "" {
		c.ImageConfigFormat = "registry.access.redhat.com/openshift3/ose-${component}:${version}"
	}

	switch cs.Properties.OrchestratorProfile.OrchestratorVersion {
	//TODO: confirm minor version after release
	case "v3.10":
		v := "v3.10"
		c.ControlPlaneImage = image(c.ImageConfigFormat, "control-plane", v)
		c.NodeImage = image(c.ImageConfigFormat, "node", v)
		c.ServiceCatalogImage = image(c.ImageConfigFormat, "service-catalog", v)
		c.AnsibleServiceBrokerImage = image(c.ImageConfigFormat, "ansible-service-broker", v)
		c.TemplateServiceBrokerImage = image(c.ImageConfigFormat, "template-service-broker", v)
		c.RegistryImage = image(c.ImageConfigFormat, "docker-registry", v)
		c.RouterImage = image(c.ImageConfigFormat, "haproxy-router", v)
		c.WebConsoleImage = image(c.ImageConfigFormat, "web-console", v)

		c.MasterEtcdImage = "registry.access.redhat.com/rhel7/etcd:3.2.22"

		c.OAuthProxyImage = "registry.access.redhat.com/openshift3/oauth-proxy:" + v
		c.PrometheusImage = "registry.access.redhat.com/openshift3/prometheus:" + v
		c.PrometheusAlertBufferImage = "registry.access.redhat.com/openshift3/prometheus-alert-buffer:" + v
		c.PrometheusAlertManagerImage = "registry.access.redhat.com/openshift3/prometheus-alertmanager:" + v
		c.PrometheusNodeExporterImage = "registry.access.redhat.com/openshift3/prometheus-node-exporter:" + v

		c.RegistryConsoleImage = "registry.access.redhat.com/openshift3/registry-console:" + v

		c.AzureCLIImage = "docker.io/microsoft/azure-cli:latest" //TODO: create mapping for OSA release to any other image we use

		c.SyncImage = "quay.io/openshift-on-azure/sync:latest"
	}
}

func selectContainerImages(cs *acsapi.ContainerService, c *Config) {
	switch os.Getenv("DEPLOY_OS") {
	case "":
		selectContainerImagesOSA(cs, c)
	case "centos7":
		selectContainerImagesOrigin(cs, c)
	}
}

func internalAPIServerHostname(cs *acsapi.ContainerService) string {
	return cs.Properties.AzProfile.ResourceGroup + "." + cs.Location + ".cloudapp.azure.com"
}

func Generate(cs *acsapi.ContainerService, c *Config) (err error) {
	c.Version = versionLatest

	// TODO: Need unique name, potentially derivative from PublicHostname
	c.RouterLBCName = fmt.Sprintf("router-%s", strings.Split(cs.Properties.OrchestratorProfile.OpenShiftConfig.PublicHostname, ".")[1])
	selectNodeImage(cs, c)

	selectContainerImages(cs, c)

	// Generate CAs
	cas := []struct {
		cn   string
		key  **rsa.PrivateKey
		cert **x509.Certificate
	}{
		{
			cn:   "etcd-signer",
			key:  &c.EtcdCaKey,
			cert: &c.EtcdCaCert,
		},
		{
			cn:   "openshift-signer",
			key:  &c.CaKey,
			cert: &c.CaCert,
		},
		{
			cn:   "openshift-frontproxy-signer",
			key:  &c.FrontProxyCaKey,
			cert: &c.FrontProxyCaCert,
		},
		{
			cn:   "openshift-service-serving-signer",
			key:  &c.ServiceSigningCaKey,
			cert: &c.ServiceSigningCaCert,
		},
		{
			cn:   "service-catalog-signer",
			key:  &c.ServiceCatalogCaKey,
			cert: &c.ServiceCatalogCaCert,
		},
	}
	for _, ca := range cas {
		if *ca.key != nil && *ca.cert != nil {
			continue
		}
		if *ca.key, *ca.cert, err = tls.NewCA(ca.cn); err != nil {
			return
		}
	}

	certs := []struct {
		cn           string
		organization []string
		dnsNames     []string
		ipAddresses  []net.IP
		extKeyUsage  []x509.ExtKeyUsage
		signingKey   *rsa.PrivateKey
		signingCert  *x509.Certificate
		key          **rsa.PrivateKey
		cert         **x509.Certificate
	}{
		// Generate etcd certs
		{
			cn:          "etcd-server",
			dnsNames:    []string{"master-000000", "master-000001", "master-000002"},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			signingKey:  c.EtcdCaKey,
			signingCert: c.EtcdCaCert,
			key:         &c.EtcdServerKey,
			cert:        &c.EtcdServerCert,
		},
		{
			cn:          "etcd-peer",
			dnsNames:    []string{"master-000000", "master-000001", "master-000002"},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			signingKey:  c.EtcdCaKey,
			signingCert: c.EtcdCaCert,
			key:         &c.EtcdPeerKey,
			cert:        &c.EtcdPeerCert,
		},
		{
			cn:          "etcd-client",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			signingKey:  c.EtcdCaKey,
			signingCert: c.EtcdCaCert,
			key:         &c.EtcdClientKey,
			cert:        &c.EtcdClientCert,
		},
		// Generate openshift master certs
		{
			cn:           "system:admin",
			organization: []string{"system:cluster-admins", "system:masters"},
			extKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:          &c.AdminKey,
			cert:         &c.AdminCert,
		},
		{
			cn:          "aggregator-front-proxy",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			signingKey:  c.FrontProxyCaKey,
			signingCert: c.FrontProxyCaCert,
			key:         &c.AggregatorFrontProxyKey,
			cert:        &c.AggregatorFrontProxyCert,
		},
		{
			cn:           "system:openshift-node-admin",
			organization: []string{"system:node-admins"},
			extKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:          &c.MasterKubeletClientKey,
			cert:         &c.MasterKubeletClientCert,
		},
		{
			cn:          "system:master-proxy",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:         &c.MasterProxyClientKey,
			cert:        &c.MasterProxyClientCert,
		},
		{
			cn: cs.Properties.OrchestratorProfile.OpenShiftConfig.PublicHostname,
			dnsNames: []string{
				internalAPIServerHostname(cs),
				cs.Properties.OrchestratorProfile.OpenShiftConfig.PublicHostname,
				"master-000000",
				"master-000001",
				"master-000002",
				"kubernetes",
				"kubernetes.default",
				"kubernetes.default.svc",
				"kubernetes.default.svc.cluster.local",
			},
			ipAddresses: []net.IP{net.ParseIP("172.30.0.1")},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			key:         &c.MasterServerKey,
			cert:        &c.MasterServerCert,
		},
		{
			cn:           "system:openshift-master",
			organization: []string{"system:cluster-admins", "system:masters"},
			extKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:          &c.OpenShiftMasterKey,
			cert:         &c.OpenShiftMasterCert,
		},
		{
			cn: "servicecatalog-api",
			dnsNames: []string{
				"servicecatalog-api",
				"apiserver.kube-service-catalog.svc", // TODO: unclear how safe this is
			},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			signingKey:  c.ServiceCatalogCaKey,
			signingCert: c.ServiceCatalogCaCert,
			key:         &c.ServiceCatalogServerKey,
			cert:        &c.ServiceCatalogServerCert,
		},
		{
			cn:          "system:serviceaccount:kube-service-catalog:service-catalog-apiserver",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:         &c.ServiceCatalogAPIClientKey,
			cert:        &c.ServiceCatalogAPIClientCert,
		},
		{
			cn:          "system:serviceaccount:openshift-infra:bootstrap-autoapprover",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:         &c.BootstrapAutoapproverKey,
			cert:        &c.BootstrapAutoapproverCert,
		},
		{
			cn:          "system:serviceaccount:openshift-infra:node-bootstrapper",
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			key:         &c.NodeBootstrapKey,
			cert:        &c.NodeBootstrapCert,
		},
		{
			cn: cs.Properties.OrchestratorProfile.OpenShiftConfig.RouterProfiles[0].PublicSubdomain,
			dnsNames: []string{
				cs.Properties.OrchestratorProfile.OpenShiftConfig.RouterProfiles[0].PublicSubdomain,
				"*." + cs.Properties.OrchestratorProfile.OpenShiftConfig.RouterProfiles[0].PublicSubdomain,
			},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			key:         &c.RouterKey,
			cert:        &c.RouterCert,
		},
		{
			cn: "docker-registry-default." + cs.Properties.OrchestratorProfile.OpenShiftConfig.RouterProfiles[0].PublicSubdomain,
			dnsNames: []string{
				"docker-registry-default." + cs.Properties.OrchestratorProfile.OpenShiftConfig.RouterProfiles[0].PublicSubdomain,
				"docker-registry.default.svc",
				"docker-registry.default.svc.cluster.local",
			},
			extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			key:         &c.RegistryKey,
			cert:        &c.RegistryCert,
		},
	}
	for _, cert := range certs {
		if cert.signingKey == nil && cert.signingCert == nil {
			cert.signingKey, cert.signingCert = c.CaKey, c.CaCert
		}
		if *cert.key != nil && *cert.cert != nil &&
			(*cert.cert).CheckSignatureFrom(cert.signingCert) == nil {
			continue
		}
		if *cert.key, *cert.cert, err = tls.NewCert(cert.cn, cert.organization, cert.dnsNames, cert.ipAddresses, cert.extKeyUsage, cert.signingKey, cert.signingCert); err != nil {
			return
		}
	}

	secrets := []struct {
		secret *[]byte
		n      int
	}{
		{
			secret: &c.SessionSecretAuth,
			n:      24,
		},
		{
			secret: &c.SessionSecretEnc,
			n:      24,
		},
		{
			secret: &c.RegistryHTTPSecret,
		},
		{
			secret: &c.AlertManagerProxySessionSecret,
		},
		{
			secret: &c.AlertsProxySessionSecret,
		},
		{
			secret: &c.PrometheusProxySessionSecret,
		},
	}
	for _, s := range secrets {
		if len(*s.secret) != 0 {
			continue
		}
		if s.n == 0 {
			s.n = 32
		}
		if *s.secret, err = randomBytes(s.n); err != nil {
			return
		}
	}

	kubeconfigs := []struct {
		clientKey  *rsa.PrivateKey
		clientCert *x509.Certificate
		endpoint   string
		username   string
		namespace  string
		kubeconfig **v1.Config
	}{
		{
			clientKey:  c.OpenShiftMasterKey,
			clientCert: c.OpenShiftMasterCert,
			endpoint:   internalAPIServerHostname(cs),
			username:   "system:openshift-master",
			kubeconfig: &c.MasterKubeconfig,
		},
		{
			clientKey:  c.BootstrapAutoapproverKey,
			clientCert: c.BootstrapAutoapproverCert,
			endpoint:   internalAPIServerHostname(cs),
			username:   "system:serviceaccount:openshift-infra:bootstrap-autoapprover",
			namespace:  "openshift-infra",
			kubeconfig: &c.BootstrapAutoapproverKubeconfig,
		},
		{
			clientKey:  c.AdminKey,
			clientCert: c.AdminCert,
			endpoint:   internalAPIServerHostname(cs),
			username:   "system:admin",
			kubeconfig: &c.AdminKubeconfig,
		},
		{
			clientKey:  c.NodeBootstrapKey,
			clientCert: c.NodeBootstrapCert,
			endpoint:   internalAPIServerHostname(cs),
			username:   "system:serviceaccount:openshift-infra:node-bootstrapper",
			kubeconfig: &c.NodeBootstrapKubeconfig,
		},
		{
			clientKey:  c.AdminKey,
			clientCert: c.AdminCert,
			// sync kubeconfig has the same capabilities as admin kubeconfig, only difference
			// is the use of HCP internal DNS to avoid waiting for the Azure loadbalancer to
			// come up in order to start creating cluster objects.
			endpoint:   "master-000000",
			username:   "system:admin",
			kubeconfig: &c.SyncKubeconfig,
		},
	}
	for _, kc := range kubeconfigs {
		if kc.namespace == "" {
			kc.namespace = "default"
		}
		if *kc.kubeconfig, err = makeKubeConfig(kc.clientKey, kc.clientCert, c.CaCert, kc.endpoint, kc.username, kc.namespace); err != nil {
			return
		}
	}

	if c.ServiceAccountKey == nil {
		if c.ServiceAccountKey, err = tls.NewPrivateKey(); err != nil {
			return
		}
	}

	if len(c.HtPasswd) == 0 {
		if c.HtPasswd, err = makeHtPasswd("demo", "demo"); err != nil {
			return
		}
	}

	if c.SSHKey == nil {
		if c.SSHKey, err = tls.NewPrivateKey(); err != nil {
			return
		}
	}

	if len(c.RegistryStorageAccount) == 0 {
		if c.RegistryStorageAccount, err = randomStorageAccountName(); err != nil {
			return
		}
	}

	if len(c.RegistryConsoleOAuthSecret) == 0 {
		if pass, err := randomString(64); err != nil {
			c.RegistryConsoleOAuthSecret = fmt.Sprintf("user%s", pass)
			return nil
		}
	}

	if len(c.RouterStatsPassword) == 0 {
		if c.RouterStatsPassword, err = randomString(10); err != nil {
			return
		}
	}

	if uuid.Equal(c.ServiceCatalogClusterID, uuid.Nil) {
		if c.ServiceCatalogClusterID, err = uuid.NewV4(); err != nil {
			return
		}
	}

	c.RunSyncLocal = os.Getenv("RUN_SYNC_LOCAL")

	c.TenantID = cs.Properties.AzProfile.TenantID
	c.SubscriptionID = cs.Properties.AzProfile.SubscriptionID
	c.ResourceGroup = cs.Properties.AzProfile.ResourceGroup

	return
}

func makeKubeConfig(clientKey *rsa.PrivateKey, clientCert, caCert *x509.Certificate, endpoint, username, namespace string) (*v1.Config, error) {
	clustername := strings.Replace(endpoint, ".", "-", -1)
	authinfoname := username + "/" + clustername
	contextname := namespace + "/" + clustername + "/" + username

	caCertBytes, err := tls.CertAsBytes(caCert)
	if err != nil {
		return nil, err
	}
	clientCertBytes, err := tls.CertAsBytes(clientCert)
	if err != nil {
		return nil, err
	}
	clientKeyBytes, err := tls.PrivateKeyAsBytes(clientKey)
	if err != nil {
		return nil, err
	}

	return &v1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []v1.NamedCluster{
			{
				Name: clustername,
				Cluster: v1.Cluster{
					Server: "https://" + endpoint,
					CertificateAuthorityData: caCertBytes,
				},
			},
		},
		AuthInfos: []v1.NamedAuthInfo{
			{
				Name: authinfoname,
				AuthInfo: v1.AuthInfo{
					ClientCertificateData: clientCertBytes,
					ClientKeyData:         clientKeyBytes,
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: contextname,
				Context: v1.Context{
					Cluster:   clustername,
					Namespace: namespace,
					AuthInfo:  authinfoname,
				},
			},
		},
		CurrentContext: contextname,
	}, nil
}

func makeHtPasswd(username, password string) ([]byte, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return append([]byte(username+":"), b...), nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadAtLeast(rand.Reader, b, n); err != nil {
		return nil, err
	}
	return b, nil
}

func randomStorageAccountName() (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 24)
	for i := range b {
		o, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		b[i] = letterBytes[o.Int64()]
	}

	return string(b), nil
}

func randomString(length int) (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		o, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		b[i] = letterBytes[o.Int64()]
	}

	return string(b), nil
}
