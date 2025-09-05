package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// EventService handles event listening for asset-transfer-events chaincode
type EventService struct {
	network *client.Network
}

// NewEventService creates a new event service instance
func NewEventService(gateway *client.Gateway) *EventService {
	//get network from gateWay
	network := gateway.GetNetwork("mychannel")

	return &EventService{
		network: network,
	}
}

// StartEventListening starts listening for chaincode events
func (es *EventService) StartEventListening(ctx context.Context) error {
	fmt.Println("🎧 Starting event listener...")
	
	events, err := es.network.ChaincodeEvents(ctx, "basic")
	if err != nil {
		return fmt.Errorf("failed to start event listening: %w", err)
	}

	go func() {
		for event := range events {
			asset := formatJSON(event.Payload)
			fmt.Printf("\n<-- Chaincode event received: %s - %s\n", event.EventName, asset)
		}
	}()

	return nil
}

// ListenForSpecificEvents listens for specific event types
func (es *EventService) ListenForSpecificEvents(ctx context.Context, eventNames []string) error {
	events, err := es.network.ChaincodeEvents(ctx, "basic", client.WithStartBlock(0))
	if err != nil {
		return fmt.Errorf("failed to start specific event listening: %w", err)
	}

	go func() {
		fmt.Printf("🔍 Listening for specific events: %v\n", eventNames)
		
		for event := range events {
			for _, name := range eventNames {
				if event.EventName == name {
					asset := formatJSON(event.Payload)
					fmt.Printf("\n<-- Specific event received: %s - %s\n", event.EventName, asset)
					break
				}
			}
		}
	}()

	return nil
}

// formatJSON formats JSON data with proper indentation
func formatJSON(data []byte) string {
	//bytes.Buffer：可增长的缓冲区，性能更好
	var result bytes.Buffer
	//&result写入目标，data源数据，""前缀，"  "缩进2空格
	if err := json.Indent(&result, data, "", "  "); err != nil {
		//这个是复制字节数据到新字符串
		return string(data)
	}
	//这个是bytes.Buffer转换为字符串
	return result.String()
}