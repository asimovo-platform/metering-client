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
		return nil, fmt.Errorf("metering configurations are not set")
	}

	logger := log.FromContext(context)
	om, err := openmeter.NewAuthClientWithResponses(server, api_key)
	if err != nil {
		return nil, err
	}

	// test connection
	res, err := om.ListMeters(context)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to openmeter: %s", err.Error())
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("connection check failed - status: %d", res.StatusCode)
	}

	// config should be ok
	logger.Info(fmt.Sprintf("Connected to openmeter: %s", server))

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
		return err
	}
	if resp.HTTPResponse.StatusCode != 204 {
		return fmt.Errorf("cannot connect to openmeter: %s", resp.HTTPResponse.Status)
	}

	logger.Info(fmt.Sprintf("openmeter resp status: %s\n", resp.HTTPResponse.Status))
	logger.Info(fmt.Sprintf("openmeter resp body: %s\n", resp.Body))

	return nil
}
