package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)


type EventService struct {
	network * client.Network
}


func NewEventService(gateway * client.Gateway) * EventService {
	//先获取network再获取event
	network := gateway.GetNetwork("mychannel")
	return &EventService{
		network: network,
	}
}

func (es * EventService) StartEventListening(ctx context.Context) error {
	fmt.Println("🎧 Starting event listener...")
	events,err := es.network.ChaincodeEvents(ctx, "novel-basic")
	if err != nil {	
		// 是的，%v是Go语言fmt包中最通用的格式化动词，几乎所有类型都可以用%v来输出其默认格式。
		// 例如：字符串、数字、结构体、切片、map、error等类型都可以用%v打印出来。
		// 但%w只能用于fmt.Errorf，并且只能用于error类型的包装，不能用于其他类型。
		return fmt.Errorf("failed to start event listening: %w", err)
	}
	//监听数据
	go func() {
		for event := range events {
			//多路复用器
			select {
				case <-ctx.Done():
					return
				default:
					novelOrUserCredit := formatJSON(event.Payload)
					fmt.Printf("\n<-- Chaincode event received: %s - %s\n", event.EventName, novelOrUserCredit)
			}
		}
	}()
	
	return nil
}


//监听特定事件
func (es * EventService) ListenForSpecificEvents(ctx context.Context, eventNames []string) error {
	events, err := es.network.ChaincodeEvents(ctx, "novel-basic", client.WithStartBlock(0))
	if err != nil {
		return fmt.Errorf("failed to start specific event listening: %w", err)
	}
	go func() {
		fmt.Printf("🔍 Listening for specific events: %v\n", eventNames)
		for {
			select {
			case event, ok := <-events:
				if !ok {
					fmt.Println("All events processed, closing listener")
					return // channel关闭，所有事件处理完
				}

				// 检查是否是指定事件
				for _, name := range eventNames {
					if event.EventName == name {
						novelOrUserCredit := formatJSON(event.Payload)
						fmt.Printf("\n<-- Event from block %d: %s - %s\n",
							event.BlockNumber, event.EventName, novelOrUserCredit)
						break
					}
				}

			case <-ctx.Done():
				fmt.Println("Context cancelled, stopping listener")
				return
			}
		}
	}()
	return nil
}

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
