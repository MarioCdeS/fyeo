package secretservice

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/secretmanager/v1"

	"github.com/MarioCdeS/fyeo/internal/errors"
)

const nameTemplate = "projects/%s/secrets/%s/versions/%s"

type SecretService struct {
	projectID string
	service   *secretmanager.Service
}

func New(projectID string, ctx context.Context) (*SecretService, error) {
	service, err := secretmanager.NewService(ctx)

	if err != nil {
		return nil, errors.NewWithCause("failed to initialise the GCP Secret Manager service", err)
	}

	return &SecretService{projectID, service}, nil
}

func (s *SecretService) Fetch(key, version string, ctx context.Context) (string, error) {
	name := fmt.Sprintf(nameTemplate, s.projectID, key, version)
	res, err := s.service.Projects.Secrets.Versions.Access(name).Context(ctx).Do()

	if err != nil {
		msg := fmt.Sprintf("failed to fetch secret %s (version %s)", key, version)
		return "", errors.NewWithCause(msg, err)
	}

	dec, err := base64.StdEncoding.DecodeString(res.Payload.Data)

	if err != nil {
		msg := fmt.Sprintf("failed to decode the value for secret %s (version %s)", key, version)
		return "", errors.NewWithCause(msg, err)
	}

	return string(dec), nil
}
