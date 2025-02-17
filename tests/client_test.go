package transport

import (
	pb_dp "github.com/media-streaming-mesh/msm-cp/api/v1alpha1/msm_dp"
	"github.com/media-streaming-mesh/msm-k8s/pkg/transport"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCreateStream(t *testing.T) {
	logger := logrus.New()

	grpcClient, err := transport.SetupClient("172.18.0.2")
	require.NoError(t, err)
	if err != nil {
		logger.Debugf("Failed to connect to server, error %s\n", err)
	}
	dpGrpcClient := transport.Client{
		Log:        logger,
		GrpcClient: grpcClient,
	}

	t.Log("FIXME")
	stream, result := dpGrpcClient.CreateStream(1, pb_dp.Encap_RTP_UDP, "192.168.82.20", 8050)
	streamDataCopy := &pb_dp.StreamData{
		Id:        stream.Id,
		Operation: stream.Operation,
		Protocol:  stream.Protocol,
		Endpoint:  stream.Endpoint,
		Enable:    stream.Enable,
	}
	logger.Debugf("Create stream-mapper %v %v", streamDataCopy, result)
}
