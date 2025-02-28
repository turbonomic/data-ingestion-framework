package mediationcontainer

import (
	"time"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// DiscoveryResponseSender handles sending messages for DiscoveryResponse to
// a specific channel.
type DiscoveryResponseSender struct {
	chunkSendDelayMillis int
	numObjectsPerChunk   int
}

// Send an input DiscoveryResponse to a specific channel. It will block till the channel is ready to receive.
// Before being sent, the DiscoveryResponse will be processed by rearranging it into multiple
// chunks with each chunk contains DTOs of one type.
// For example, if the input DiscoveryDesponse contains DTOs for Entity and Notification,
// two DiscoveryResponse instances (chunks) for Entity and Notification will be sent to the channel.
func (d *DiscoveryResponseSender) Send(discoveryResponse *proto.DiscoveryResponse, msgID int32, probeMsgChan chan *proto.MediationClientMessage) {
	chunks := d.chunkDiscoveryResponse(discoveryResponse)
	glog.V(3).Infof("Total chunks to be sent: %d", len(chunks))
	// Send chunked discovery response for each DTO type
	numChunkSend := 0
	for _, chunk := range chunks {
		clientMsg := NewClientMessageBuilder(msgID).SetDiscoveryResponse(chunk).Create()
		// Wait for a preset delay, if configured
		if d.chunkSendDelayMillis > 0 {
			glog.V(3).Infof("Sleeping %v milliseconds before sending next chunk", d.chunkSendDelayMillis)
			time.Sleep(time.Millisecond * time.Duration(d.chunkSendDelayMillis))
		}
		// Send the response on the callback channel to send to the server
		probeMsgChan <- clientMsg // This will block till the channel is ready to receive
		numChunkSend++
		glog.V(3).Infof("chunk #%d sent", numChunkSend)
	}
}

func chunkData[T any](items []T, chunkSize int) (chunks [][]T) {
	glog.V(3).Infof("Generating chunks of %v DTOs", chunkSize)
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// Send discovery response. The message will be chunked so that for each DTO type in DiscoveryResponse, one chunk
// will be generated and sent. This is required to get the DTOs passed at server side.
func (d *DiscoveryResponseSender) chunkDiscoveryResponse(dr *proto.DiscoveryResponse) []*proto.DiscoveryResponse {
	chunks := []*proto.DiscoveryResponse{}

	if len(dr.EntityDTO) > 0 {
		entityChunks := chunkData(dr.EntityDTO, d.numObjectsPerChunk)
		for idx, entityChunk := range entityChunks {
			glog.V(3).Infof("chunk #%v consists of %d EntityDTOs", idx+1, len(entityChunk))
			chunk := &proto.DiscoveryResponse{
				EntityDTO: entityChunk,
			}
			chunks = append(chunks, chunk)
		}
	}

	if len(dr.ErrorDTO) > 0 {
		glog.V(3).Infof("chunk with %d ErrorDTO", len(dr.ErrorDTO))
		chunk := &proto.DiscoveryResponse{
			ErrorDTO: dr.ErrorDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DiscoveredGroup) > 0 {
		glog.V(3).Infof("chunk with %d DiscoveredGroup", len(dr.DiscoveredGroup))
		chunk := &proto.DiscoveryResponse{
			DiscoveredGroup: dr.DiscoveredGroup,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.EntityProfile) > 0 {
		glog.V(3).Infof("chunk with %d EntityProfile", len(dr.EntityProfile))
		chunk := &proto.DiscoveryResponse{
			EntityProfile: dr.EntityProfile,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DeploymentProfile) > 0 {
		glog.V(3).Infof("chunk with %d DeploymentProfile", len(dr.DeploymentProfile))
		chunk := &proto.DiscoveryResponse{
			DeploymentProfile: dr.DeploymentProfile,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.Notification) > 0 {
		glog.V(3).Infof("chunk with %d Notification", len(dr.Notification))
		chunk := &proto.DiscoveryResponse{
			Notification: dr.Notification,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.MetadataDTO) > 0 {
		glog.V(3).Infof("chunk with %d MetadataDTO", len(dr.MetadataDTO))
		chunk := &proto.DiscoveryResponse{
			MetadataDTO: dr.MetadataDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.DerivedTarget) > 0 {
		glog.V(3).Infof("chunk with %d DerivedTarget", len(dr.DerivedTarget))
		chunk := &proto.DiscoveryResponse{
			DerivedTarget: dr.DerivedTarget,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.NonMarketEntityDTO) > 0 {
		glog.V(3).Infof("chunk with %d NonMarketEntityDTO", len(dr.NonMarketEntityDTO))
		chunk := &proto.DiscoveryResponse{
			NonMarketEntityDTO: dr.NonMarketEntityDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.FlowDTO) > 0 {
		glog.V(3).Infof("chunk with %d FlowDTO", len(dr.FlowDTO))
		chunk := &proto.DiscoveryResponse{
			FlowDTO: dr.FlowDTO,
		}
		chunks = append(chunks, chunk)
	}

	if len(dr.ActionPolicies) > 0 {
		glog.V(3).Infof("chunk with %d ActionPolicies", len(dr.ActionPolicies))
		chunk := &proto.DiscoveryResponse{
			ActionPolicies: dr.ActionPolicies,
		}
		chunks = append(chunks, chunk)
	}

	// Send an empty response to signal completion of discovery
	chunks = append(chunks, &proto.DiscoveryResponse{})

	return chunks
}
