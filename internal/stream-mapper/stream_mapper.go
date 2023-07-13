package stream_mapper

import (
	"errors"
	pbDp "github.com/media-streaming-mesh/msm-cp/api/v1alpha1/msm_dp"
	"github.com/media-streaming-mesh/msm-k8s/pkg/model"
	nodeMapper "github.com/media-streaming-mesh/msm-k8s/pkg/node_mapper"
	"github.com/media-streaming-mesh/msm-k8s/pkg/stream_api"
	"github.com/media-streaming-mesh/msm-k8s/pkg/transport"
	"github.com/sirupsen/logrus"
	"sync"
)

var streamId StreamId

// StreamId type uint32 auto increment
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
		m.logger.Errorf("Failed to setup GRPC client, error %s\n", err)
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

	//_, err := m.streamAPI.GetStreams()
	//if err != nil {
	//	m.logger.Errorf("unable to get streams %v", err)
	//}

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
			m.logger.Errorf("Process stream failed %v", err)
		}
		m.waitForData()
	}()
}

func (m *StreamMapper) processStream(data model.StreamData) error {
	var serverProxyIP string
	var clientProxyIP string
	var serverDpGrpcClient transport.Client
	var clientDpGrpcClient transport.Client

	m.logger.Infof("Processing stream %v", data)

	if len(data.ServerPorts) == 0 || len(data.ClientPorts) == 0 {
		return errors.New("[Stream Mapper] empty server/client port")
	}

	// Check if client/server on same node
	isOnSameNode := nodeMapper.IsOnSameNode(data.StubIp, data.ServerIp)
	m.logger.Infof("server %v and client %v - same node is %v", data.ServerIp, data.StubIp, isOnSameNode)

	// Get proxy ip
	clientProxyIP, err := nodeMapper.MapNode(data.StubIp)
	if err != nil {
		return err
	}
	m.logger.Infof("client msm-proxy ip %v", clientProxyIP)
	if !isOnSameNode {
		serverProxyIP, err = nodeMapper.MapNode(data.ServerIp)
		if err != nil {
			return err
		}
		m.logger.Infof("server msm-proxy ip %v", serverProxyIP)
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
			m.logger.Debugf("stream Id %v", streamId)

			// Create new stream
			stream = model.Stream{
				StreamId: streamId,
				ProxyMap: make(map[string]model.Proxy),
			}

			if isOnSameNode {
				// add the stream-mapper to the proxy
				streamData, result := clientDpGrpcClient.CreateStream(streamId, pbDp.Encap_RTP_UDP, data.ServerIp, data.ServerPorts[0])
				streamDataCopy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.logger.Debugf("GRPC create stream %v result %v", streamDataCopy, result)
			} else {
				// add the stream-mapper to the server proxy
				streamData, result := serverDpGrpcClient.CreateStream(streamId, pbDp.Encap_RTP_UDP, data.ServerIp, data.ServerPorts[0])
				streamDataCopy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.logger.Debugf("GRPC create server stream %v result %v", streamDataCopy, result)

				// add the stream-mapper to the client proxy
				streamData, result = clientDpGrpcClient.CreateStream(streamId, pbDp.Encap_RTP_UDP, serverProxyIP, data.ServerPorts[0])
				streamDataClientProxy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.logger.Debugf("GRPC create client stream %v result %v", streamDataClientProxy, result)

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
				streamData, result := clientDpGrpcClient.CreateStream(stream.StreamId, pbDp.Encap_RTP_UDP, serverProxyIP, data.ServerPorts[0])
				streamDataCopy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.logger.Debugf("GRPC create client stream %v result %v", streamDataCopy, result)

				// add the client proxy to the proxy map
				stream.ProxyMap[clientProxyIP] = model.Proxy{
					ProxyIp:     clientProxyIP,
					StreamState: model.Create,
				}
			}
		}

		// Send Create Endpoint to proxy
		clientProxy := stream.ProxyMap[clientProxyIP]
		m.logger.Debugf("Client proxy %v total clients %v", clientProxy, len(clientProxy.Clients))

		if !isOnSameNode && clientProxy.StreamState < model.Play {
			endpoint, result := serverDpGrpcClient.CreateEndpoint(stream.StreamId, pbDp.Encap_RTP_UDP, clientProxyIP, 8050)
			endpointCopy := &pbDp.Endpoint{
				Ip:         endpoint.Ip,
				Port:       endpoint.Port,
				QuicStream: endpoint.QuicStream,
				Encap:      endpoint.Encap,
			}
			m.logger.Debugf("GRPC created proxy-proxy ep %v result %v", endpointCopy, result)
		}

		endpoint, result := clientDpGrpcClient.CreateEndpoint(stream.StreamId, pbDp.Encap_RTP_UDP, data.ClientIp, data.ClientPorts[0])
		endpointCopy := &pbDp.Endpoint{
			Ip:         endpoint.Ip,
			Port:       endpoint.Port,
			QuicStream: endpoint.QuicStream,
			Encap:      endpoint.Encap,
		}
		m.logger.Debugf("GRPC create client ep %v result %v", endpointCopy, result)

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
		m.logger.Debugf("Stream %v", stream)
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
		m.logger.Debugf("Client proxy PLAY proxy clients %v total clients %v", clientProxy, len(clientProxy.Clients))
		for _, c := range clientProxy.Clients {
			if c.ClientIp == data.ClientIp && c.Port == data.ClientPorts[0] {
				endpoint, result := clientDpGrpcClient.UpdateEndpoint(stream.StreamId, data.ClientIp, data.ClientPorts[0])
				endpointCopy := &pbDp.Endpoint{
					Ip:         endpoint.Ip,
					Port:       endpoint.Port,
					QuicStream: endpoint.QuicStream,
					Encap:      endpoint.Encap,
				}
				m.logger.Debugf("GRPC update ep %v result %v", endpointCopy, result)

				if !isOnSameNode && clientProxy.StreamState < model.Play {
					endpoint2, result := serverDpGrpcClient.UpdateEndpoint(stream.StreamId, clientProxyIP, 8050)
					stream.ProxyMap[clientProxyIP] = model.Proxy{
						ProxyIp:     clientProxy.ProxyIp,
						StreamState: model.Play,
						Clients:     clientProxy.Clients,
					}
					endpoint2Copy := &pbDp.Endpoint{
						Ip:         endpoint2.Ip,
						Port:       endpoint2.Port,
						QuicStream: endpoint2.QuicStream,
						Encap:      endpoint2.Encap,
					}
					m.logger.Debugf("GRPC update proxy ep %v result %v", endpoint2Copy, result)
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
		m.logger.Debugf("Client proxy TEARDOWN %v proxy clients %v total clients %v", clientProxy, len(clientProxy.Clients), m.getClientCount(data.ServerIp))

		for i, c := range clientProxy.Clients {
			if c.ClientIp == data.ClientIp && c.Port == data.ClientPorts[0] {
				endpoint, result := clientDpGrpcClient.DeleteEndpoint(stream.StreamId, data.ClientIp, data.ClientPorts[0])
				endpointCopy := &pbDp.Endpoint{
					Ip:         endpoint.Ip,
					Port:       endpoint.Port,
					QuicStream: endpoint.QuicStream,
					Encap:      endpoint.Encap,
				}
				m.logger.Debugf("GRPC delete ep %v %v", endpointCopy, result)
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
			endpoint2Copy := &pbDp.Endpoint{
				Ip:         endpoint2.Ip,
				Port:       endpoint2.Port,
				QuicStream: endpoint2.QuicStream,
				Encap:      endpoint2.Encap,
			}
			m.logger.Debugf("GRPC delete ep %v %v", endpoint2Copy, result)
		}

		// Delete stream if all clients terminate
		if m.getClientCount(data.ServerIp) == 0 {
			if isOnSameNode {
				streamData, result := clientDpGrpcClient.DeleteStream(stream.StreamId, data.ServerIp, 8050)
				streamDataCopy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.streamMap.Delete(data.ServerIp)
				m.logger.Debugf("GRPC delete stream %v %v", streamDataCopy, result)
			} else {
				streamData, result := serverDpGrpcClient.DeleteStream(stream.StreamId, data.ServerIp, 8050)
				streamDataCopy := &pbDp.StreamData{
					Id:        streamData.Id,
					Operation: streamData.Operation,
					Protocol:  streamData.Protocol,
					Endpoint:  streamData.Endpoint,
					Enable:    streamData.Enable,
				}
				m.streamMap.Delete(data.ServerIp)
				m.logger.Debugf("GRPC delete stream %v %v", streamDataCopy, result)

				streamData2, result := clientDpGrpcClient.DeleteStream(stream.StreamId, serverProxyIP, 8050)
				streamDataCopy2 := &pbDp.StreamData{
					Id:        streamData2.Id,
					Operation: streamData2.Operation,
					Protocol:  streamData2.Protocol,
					Endpoint:  streamData2.Endpoint,
					Enable:    streamData2.Enable,
				}
				m.logger.Debugf("GRPC delete stream %v %v", streamDataCopy2, result)
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
