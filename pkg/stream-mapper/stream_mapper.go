package stream_mapper

import (
	"errors"
	pb_dp "github.com/media-streaming-mesh/msm-cp/api/v1alpha1/msm_dp"
	"github.com/media-streaming-mesh/msm-cp/pkg/model"
	node_mapper "github.com/media-streaming-mesh/msm-cp/pkg/node-mapper"
	"github.com/media-streaming-mesh/msm-cp/pkg/stream_api"
	"github.com/media-streaming-mesh/msm-cp/pkg/transport"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

var log *logrus.Logger

var streamId StreamId

// Stream id type uint32 auto increment
type StreamId struct {
	sync.Mutex
	id uint32
}

type StreamMapper struct {
	logger    *logrus.Logger
	dpClients map[string]transport.Client
	streamMap *sync.Map
	streamAPI *stream_api.StreamAPI
	dataChan  chan model.StreamData
}

func NewStreamMapper(logger *logrus.Logger, streamMap *sync.Map) *StreamMapper {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	setLogLvl(log)
	setLogType(log)
	return &StreamMapper{
		logger:    logger,
		dpClients: make(map[string]transport.Client),
		streamMap: streamMap,
		streamAPI: stream_api.NewStreamAPI(logger),
		dataChan:  make(chan model.StreamData, 1),
	}
}

func (m *StreamMapper) ConnectClient(ip string) {
	grpcClient, err := transport.SetupClient(ip)
	if err != nil {
		log.Errorf("Failed to setup GRPC client, error %s\n", err)
	}
	dpClient := transport.Client{
		Log:        m.logger,
		GrpcClient: grpcClient,
	}

	m.dpClients[ip] = dpClient
}

func (m *StreamMapper) DisconnectClient(ip string) {
	dpClient := m.dpClients[ip]
	dpClient.Close()
	delete(m.dpClients, ip)
}

func (m *StreamMapper) WatchStream() {
	//TODO: process previous cached streams
	_, err := m.streamAPI.GetStreams()
	if err != nil {
		log.Errorf("unable to get streams %v", err)
	}

	//TODO: remove test code
	//m.streamAPI.DeleteStreams()

	m.waitForData()
	m.streamAPI.WatchStreams(m.dataChan)
}

func (m *StreamMapper) waitForData() {
	go func() {
		streamData := <-m.dataChan
		err := m.processStream(streamData)
		if err != nil {
			log.Errorf("Process stream failed %v", err)
		}
		m.waitForData()
	}()
}

func (m *StreamMapper) processStream(data model.StreamData) error {
	var serverProxyIP string
	var clientProxyIP string
	var serverDpGrpcClient transport.Client
	var clientDpGrpcClient transport.Client

	log.Infof("Processing stream %v", data)

	if len(data.ServerPorts) == 0 || len(data.ClientPorts) == 0 {
		return errors.New("[Stream Mapper] empty server/client port")
	}

	// Check if client/server on same node
	isOnSameNode := node_mapper.IsOnSameNode(data.StubIp, data.ServerIp)
	log.Infof("server %v and client %v - same node is %v", data.ServerIp, data.StubIp, isOnSameNode)

	// Get proxy ip
	clientProxyIP, err := node_mapper.MapNode(data.StubIp)
	if err != nil {
		return err
	}
	log.Infof("client msm-proxy ip %v", clientProxyIP)
	if !isOnSameNode {
		serverProxyIP, err = node_mapper.MapNode(data.ServerIp)
		if err != nil {
			return err
		}
		log.Infof("server msm-proxy ip %v", serverProxyIP)
	}

	// Create GRPC connection
	clientDpGrpcClient = m.dpClients[clientProxyIP]

	if !isOnSameNode {
		serverDpGrpcClient = m.dpClients[serverProxyIP]
	}

	// Send data to proxy
	if data.StreamState == model.Create {
		var stream model.Stream
		savedStream, _ := m.streamMap.Load(data.ServerIp)

		// Send Create Stream to proxy
		if savedStream == nil {
			streamId := GetStreamID()
			log.Debugf("stream Id %v", streamId)

			// Create new stream
			stream = model.Stream{
				StreamId: streamId,
				ProxyMap: make(map[string]model.Proxy),
			}

			if isOnSameNode {
				// add the stream-mapper to the proxy
				streamData, result := clientDpGrpcClient.CreateStream(streamId, pb_dp.Encap_RTP_UDP, data.ServerIp, data.ServerPorts[0])
				log.Debugf("GRPC create stream %v result %v", streamData, *result)
			} else {
				// add the stream-mapper to the server proxy
				streamData, result := serverDpGrpcClient.CreateStream(streamId, pb_dp.Encap_RTP_UDP, data.ServerIp, data.ServerPorts[0])
				log.Debugf("GRPC create server stream %v result %v", streamData, *result)

				// add the stream-mapper to the client proxy
				streamData, result = clientDpGrpcClient.CreateStream(streamId, pb_dp.Encap_RTP_UDP, serverProxyIP, data.ServerPorts[0])
				log.Debugf("GRPC create client stream %v result %v", streamData, *result)

				// add the server proxy to the proxy map
				stream.ProxyMap[serverProxyIP] = model.Proxy{
					ProxyIp:     serverProxyIP,
					StreamState: model.Create,
				}
			}

			// add the client proxy to the proxy map
			stream.ProxyMap[clientProxyIP] = model.Proxy{
				ProxyIp:     clientProxyIP,
				StreamState: model.Create,
			}

			// now store the RTSP stream-mapper
			m.streamMap.Store(data.ServerIp, stream)
		} else {
			stream = savedStream.(model.Stream)
			if _, exists := stream.ProxyMap[clientProxyIP]; !exists {
				// add the stream-mapper to the client proxy
				streamData, result := clientDpGrpcClient.CreateStream(stream.StreamId, pb_dp.Encap_RTP_UDP, serverProxyIP, data.ServerPorts[0])
				log.Debugf("GRPC create client stream %v result %v", streamData, *result)

				// add the client proxy to the proxy map
				stream.ProxyMap[clientProxyIP] = model.Proxy{
					ProxyIp:     clientProxyIP,
					StreamState: model.Create,
				}
			}
		}

		// Send Create Endpoint to proxy
		clientProxy := stream.ProxyMap[clientProxyIP]
		log.Debugf("Client proxy %v total clients %v", clientProxy, len(clientProxy.Clients))

		if !isOnSameNode && clientProxy.StreamState < model.Play {
			endpoint, result := serverDpGrpcClient.CreateEndpoint(stream.StreamId, pb_dp.Encap_RTP_UDP, clientProxyIP, 8050)
			log.Debugf("GRPC created proxy-proxy ep %v result %v", endpoint, *result)
		}

		endpoint, result := clientDpGrpcClient.CreateEndpoint(stream.StreamId, pb_dp.Encap_RTP_UDP, data.ClientIp, data.ClientPorts[0])
		log.Debugf("GRPC create client ep %v result %v", endpoint, result)

		// Update streamMap data
		clientProxy.Clients = append(clientProxy.Clients, model.Client{
			ClientIp: data.ClientIp,
			Port:     data.ClientPorts[0],
		})
		stream.ProxyMap[clientProxyIP] = model.Proxy{
			ProxyIp:     clientProxyIP,
			StreamState: clientProxy.StreamState,
			Clients:     clientProxy.Clients,
		}
		log.Debugf("Stream %v", stream)
	}

	if data.StreamState == model.Play {
		savedStream, _ := m.streamMap.Load(data.ServerIp)
		if savedStream == nil {
			return errors.New("[Stream Mapper] Can't find stream")
		}

		stream := savedStream.(model.Stream)
		clientProxy, exists := stream.ProxyMap[clientProxyIP]
		if !exists {
			return errors.New("[Stream Mapper] Can't find client proxy")
		}
		log.Debugf("Client proxy PLAY proxy clients %v total clients %v", clientProxy, len(clientProxy.Clients))
		for _, c := range clientProxy.Clients {
			if c.ClientIp == data.ClientIp && c.Port == data.ClientPorts[0] {
				endpoint, result := clientDpGrpcClient.UpdateEndpoint(stream.StreamId, data.ClientIp, data.ClientPorts[0])
				log.Debugf("GRPC update ep %v result %v", endpoint, result)

				if !isOnSameNode && clientProxy.StreamState < model.Play {
					endpoint2, result := serverDpGrpcClient.UpdateEndpoint(stream.StreamId, clientProxyIP, 8050)
					stream.ProxyMap[clientProxyIP] = model.Proxy{
						ProxyIp:     clientProxy.ProxyIp,
						StreamState: model.Play,
						Clients:     clientProxy.Clients,
					}
					log.Debugf("GRPC update proxy ep %v result %v", endpoint2, result)
				}
				break
			}
		}
	}

	if data.StreamState == model.Teardown {
		savedStream, _ := m.streamMap.Load(data.ServerIp)
		if savedStream == nil {
			return errors.New("[Stream Mapper] Can't find stream")
		}

		stream := savedStream.(model.Stream)
		clientProxy, exists := stream.ProxyMap[clientProxyIP]
		if !exists {
			return errors.New("[Stream Mapper] Can't find client proxy")
		}
		log.Debugf("Client proxy TEARDOWN %v proxy clients %v total clients %v", clientProxy, len(clientProxy.Clients), m.getClientCount(data.ServerIp))

		for i, c := range clientProxy.Clients {
			if c.ClientIp == data.ClientIp && c.Port == data.ClientPorts[0] {
				endpoint, result := clientDpGrpcClient.DeleteEndpoint(stream.StreamId, data.ClientIp, data.ClientPorts[0])
				log.Debugf("GRPC delete ep %v %v", endpoint, result)
				clientProxy.Clients = append(clientProxy.Clients[:i], clientProxy.Clients[i+1:]...)
				break
			}
		}

		stream.ProxyMap[clientProxyIP] = model.Proxy{
			ProxyIp:     clientProxy.ProxyIp,
			StreamState: clientProxy.StreamState,
			Clients:     clientProxy.Clients,
		}

		// End proxy-proxy connection
		if !isOnSameNode && clientProxy.StreamState < model.Teardown && len(clientProxy.Clients) == 0 {
			endpoint2, result := serverDpGrpcClient.DeleteEndpoint(stream.StreamId, clientProxyIP, 8050)
			delete(stream.ProxyMap, clientProxyIP)
			log.Debugf("GRPC delete ep %v %v", endpoint2, result)
		}

		// Delete stream if all clients terminate
		if m.getClientCount(data.ServerIp) == 0 {
			if isOnSameNode {
				streamData, result := clientDpGrpcClient.DeleteStream(stream.StreamId, data.ServerIp, 8050)
				m.streamMap.Delete(data.ServerIp)
				log.Debugf("GRPC delete stream %v %v", streamData, result)
			} else {
				streamData, result := serverDpGrpcClient.DeleteStream(stream.StreamId, data.ServerIp, 8050)
				m.streamMap.Delete(data.ServerIp)
				log.Debugf("GRPC delete stream %v %v", streamData, result)

				streamData2, result := clientDpGrpcClient.DeleteStream(stream.StreamId, serverProxyIP, 8050)
				log.Debugf("GRPC delete stream %v %v", streamData2, result)
			}
		}
	}
	return nil
}

func (m *StreamMapper) getClientCount(serverEp string) int {
	count := 0
	data, _ := m.streamMap.Load(serverEp)
	if data != nil {
		for _, v := range data.(model.Stream).ProxyMap {
			count += len(v.Clients)
		}
	}
	return count
}

func (si *StreamId) ID() (id uint32) {
	si.Lock()
	defer si.Unlock()
	id = si.id
	si.id++
	return
}

func GetStreamID() uint32 {
	return streamId.ID()
}

// sets the log level of the logger
func setLogLvl(l *logrus.Logger) {
	logLevel := os.Getenv("LOG_LEVEL")

	switch logLevel {
	case "DEBUG":
		l.SetLevel(logrus.DebugLevel)
	case "WARN":
		l.SetLevel(logrus.WarnLevel)
	case "INFO":
		l.SetLevel(logrus.InfoLevel)
	case "ERROR":
		l.SetLevel(logrus.ErrorLevel)
	case "TRACE":
		l.SetLevel(logrus.TraceLevel)
	case "FATAL":
		l.SetLevel(logrus.FatalLevel)
	default:
		l.SetLevel(logrus.DebugLevel)
	}
}

// sets the log type of the logger
func setLogType(l *logrus.Logger) {
	logType := os.Getenv("LOG_TYPE")

	switch strings.ToLower(logType) {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: true,
		})
	default:
		l.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}
