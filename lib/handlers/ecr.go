// Event handler for ECR auth
package handlers

import (
	"encoding/json"
	"log"

	"github.com/liangrog/kctlr-docker-auth/lib/aws"
	lc "github.com/liangrog/kctlr-docker-auth/lib/core"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Handling namespace event
type EcrHandler struct {
	// kubernetes client
	client kubernetes.Interface

	// Namespaces exclude
	ExcludeNamespaces []string

	// Name of the secret
	// Default [prefix]-ecr-<iam username>
	SecretName string
}

// Create a new ECR handler
func NewEcrHandler(client kubernetes.Interface, nsExclud []string, secretName string) *EcrHandler {
	/*user, err := aws.GetIamUser()
	if err != nil {
		log.Fatal("Failure to get IAM user for ECR handler")
	}
	*/
	var name string
	if secretName != "" {
		name = secretName
	} else {
		name = "ecr"
	}

	return &EcrHandler{
		client:            client,
		ExcludeNamespaces: nsExclud,
		SecretName:        name,
	}

}

// If skip namespace processing
func (h *EcrHandler) ifSkip(namespace string) bool {
	found, _ := lc.InArray(namespace, h.ExcludeNamespaces)
	return found
}

// docker config
type dockerConfig struct {
	Auths map[string]registryAuth `json:"auths,omitempty"`
}

// docker config
type registryAuth struct {
	Auth  string `json:"auth"`
	Email string `json:"email"`
}

// Construct ECR Secret
func (h *EcrHandler) buildEcrSecret(namespace string) (*v1.Secret, error) {
	result, err := aws.GetEcrAuths()
	if err != nil {
		return nil, err
	}

	auths := map[string]registryAuth{}
	for _, data := range result {
		endpoint := *data.ProxyEndpoint
		auths[endpoint] = registryAuth{
			Auth:  *data.AuthorizationToken,
			Email: "none",
		}
	}

	dcJson, err := json.Marshal(dockerConfig{Auths: auths})
	if err != nil {
		return nil, err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.SecretName,
			Namespace: namespace,
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: map[string][]byte{".dockerconfigjson": dcJson},
	}

	return secret, nil
}

// Handling update and create
// Create event will be trigger during initiation
// hence no different to update
func (h *EcrHandler) Upsert(obj interface{}) error {
	ns := obj.(*v1.Namespace)

	// Don't process if excluded
	if skip := h.ifSkip(ns.GetName()); skip {
		// If exist, delete from excluded namespace
		if s, err := lc.GetSecret(h.client, ns.GetName(), h.SecretName, metav1.GetOptions{}); s.Name != "" {
			err = lc.DeleteSecret(h.client, ns.GetName(), h.SecretName, &metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Failed to delete existing secret %s in namespace %s", h.SecretName, ns.GetName())
			}
			log.Printf("Deleted existing secret %s in namespace %s", h.SecretName, ns.GetName())
		}

		log.Printf("Ignoring excluded namespace %s in update", ns.GetName())
		return nil
	}

	secret, err := h.buildEcrSecret(ns.GetName())
	if err != nil {
		return err
	}

	// Check if it's a new secret
	s, err := lc.GetSecret(h.client, ns.GetName(), h.SecretName, metav1.GetOptions{})
	if s.Name == "" {
		// Create
		_, err = lc.CreateSecret(h.client, ns.GetName(), secret)
		if err != nil {
			log.Printf("Failed to create ECR secret %s for namespace %s", h.SecretName, ns.GetName())
			return err
		}

		log.Printf("Successfully created ECR secret %s for namespace %s", h.SecretName, ns.GetName())
	} else {
		// Update
		_, err = lc.UpdateSecret(h.client, ns.GetName(), secret)
		if err != nil {
			log.Printf("Failed to update ECR secret %s for namespace %s", h.SecretName, ns.GetName())
			return err
		}

		log.Printf("Successfully updated ECR secret %s for namespace %s", h.SecretName, ns.GetName())
	}

	return nil
}

// Handling namespace add/update event
func (h *EcrHandler) Create(obj interface{}) error {
	return h.Upsert(obj)
}

// Handler namespace update
func (h *EcrHandler) Update(old, new interface{}) error {
	return h.Upsert(new)
}

// Handling namespace delete event
// Don't need to delete as NS is delete anyway
func (h *EcrHandler) Delete(obj interface{}) error {
	return nil
}
