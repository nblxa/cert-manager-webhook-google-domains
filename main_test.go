package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
	"gopkg.in/yaml.v3"
)

var (
	domain = os.Getenv("TEST_DOMAIN_NAME")
	apiKey = os.Getenv("TEST_SECRET")

	configFile         = "testdata/googledomains/config.json"
	secretYamlFilePath = "testdata/googledomains/cert-manager-google-domains-secret.yml"
	secretName         = "cert-manager-google-domains-secret"
	secretKeyName      = "api-key"
)

type SecretYaml struct {
	ApiVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
	SecretType string `yaml:"type" json:"type"`
	Metadata   struct {
		Name string `yaml:"name"`
	}
	Data struct {
		ApiKey string `yaml:"api-key"`
	}
}

func TestRunsSuite(t *testing.T) {
	slogger := zapLogger.Sugar()

	secretYaml := SecretYaml{}
	secretYaml.ApiVersion = "v1"
	secretYaml.Kind = "Secret"
	secretYaml.SecretType = "Opaque"
	secretYaml.Metadata.Name = secretName
	secretYaml.Data.ApiKey = apiKey

	secretYamlFile, err := yaml.Marshal(&secretYaml)
	if err != nil {
		slogger.Error(err)
	}
	_ = ioutil.WriteFile(secretYamlFilePath, secretYamlFile, 0644)

	providerConfig := googledomainsDNSProviderConfig{
		"https://acmedns.googleapis.com/v1",
		domain,
		secretName,
		secretKeyName,
	}
	file, _ := json.MarshalIndent(providerConfig, "", " ")
	_ = ioutil.WriteFile(configFile, file, 0644)

	fixture := dns.NewFixture(&googledomainsDNSProviderSolver{},
		dns.SetResolvedZone(domain),
		dns.SetResolvedFQDN(GetRandomString(8)+"."+domain+"."),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/googledomains"),
		dns.SetBinariesPath("_test/kubebuilder/bin"),
		dns.SetStrict(false),
	)

	fixture.RunConformance(t)

	_ = os.Remove(configFile)
	_ = os.Remove(secretYamlFilePath)
}

func GetRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}