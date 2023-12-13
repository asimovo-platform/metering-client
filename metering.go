package metering

import (
	"context"
	"fmt"

	openmeter "github.com/asimovo-platform/metering-client/client"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MeteringClient struct {
	om  *openmeter.ClientWithResponses
	ctx context.Context
}

func NewMeteringClient(context context.Context) (*MeteringClient, error) {
	logger := log.FromContext(context)

	// TODO: Use environment variable
	om, err := openmeter.NewAuthClientWithResponses("https://openmeter.cloud", "om_3NpyD5J4WX0lXPN55LMLIfxDKd81RcDj.63th27OqK-MarYeXarelvaz1dDiohPs8s6t4914T1Pc")
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
