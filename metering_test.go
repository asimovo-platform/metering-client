package metering

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

func TestMeteringClient_SendEvent(t *testing.T) {
	t.Run("1. Test metering client using correct config", func(t *testing.T) {
		os.Setenv("METERING_SERVER", "https://openmeter.cloud")
		os.Setenv("METERING_API_KEY", "om_3NpyD5J4WX0lXPN55LMLIfxDKd81RcDj.63th27OqK-MarYeXarelvaz1dDiohPs8s6t4914T1Pc ")

		client, err := NewMeteringClient(context.Background())
		if err != nil {
			t.Errorf("unable to create metering client: %s", err)
			return
		}

		if client == nil {
			t.Errorf("metering client is nil for some reason")
			return
		}

		e := cloudevents.New()
		e.SetID(uuid.New().String())
		e.SetType("workstation-runtime")
		e.SetSubject("metering client test")
		e.SetTime(time.Now())
		e.SetSource("asimovo-workload-operator")
		e.SetData("application/json", map[string]string{
			"name":      "metering-client name",
			"namespace": "metering-client namespace",
			"phase":     "metering-client phase",
			"config":    "metering-client config",
		})

		err = client.SendEvent(e)
		if err != nil {
			t.Errorf(fmt.Sprintf("failed openmeter sending event test: %s", err))
			return
		}

		t.Log("successfully sent event to openmeter")
	})

	t.Run("2. Test metering client using empty config", func(t *testing.T) {
		os.Setenv("METERING_SERVER", "")
		os.Setenv("METERING_API_KEY", "")

		_, err := NewMeteringClient(context.Background())
		if err == nil {
			t.Errorf("error should be returned")
			return
		}

		if err.Error() != "metering configurations are not set" {
			t.Errorf("incorrect error returned: %s", err.Error())
			return
		}

		t.Logf("successfully returned error: %s", err.Error())
	})

	t.Run("3. Test metering client using incorrect config", func(t *testing.T) {
		os.Setenv("METERING_SERVER", "https://localhost:9001")
		os.Setenv("METERING_API_KEY", "qwerty123")
		os.Setenv("METER_ID", "test")

		_, err := NewMeteringClient(context.Background())
		if err == nil {
			t.Errorf("error should be returned")
			return
		}

		t.Logf("successfully returned error: %s", err.Error())
	})
}
