package metering

import (
	"context"
	"fmt"
	"os"

	openmeter "github.com/asimovo-platform/metering-client/client"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MeteringClient struct {
	om  *openmeter.ClientWithResponses
	ctx context.Context
}

func NewMeteringClient(context context.Context) (*MeteringClient, error) {
	server := os.Getenv("METERING_SERVER")

	api_key := os.Getenv("METERING_API_KEY")

	if server == "" || api_key == "" {
		return nil, fmt.Errorf("metering server or api key not set")
	}

	logger := log.FromContext(context)
	om, err := openmeter.NewAuthClientWithResponses(server, api_key)
	if err != nil {
		return nil, err
	}

	logger.Info("connected to openmeter")

	return &MeteringClient{
		om:  om,
		ctx: context,
	}, nil
}

func (client *MeteringClient) SendEvent(event cloudevents.Event) error {
	logger := log.FromContext(client.ctx)

	logger.Info(fmt.Sprintf("Sending event to openmeter: %s", event.ID()))

	resp, err := client.om.IngestEventWithResponse(client.ctx, event)
	if err != nil {
		logger.Error(err, fmt.Sprintf("unable to send event to openmeter: %s", err.Error()))
	}

	logger.Info(fmt.Sprintf("openmeter resp status: %s\n", resp.HTTPResponse.Status))
	logger.Info(fmt.Sprintf("openmeter resp body: %s\n", resp.Body))

	return nil
}
