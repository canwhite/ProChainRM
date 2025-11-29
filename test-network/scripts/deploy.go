package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)


func main(){
	fmt.Println("Starting network deployment with single shell session...")

	// 使用一个shell脚本执行所有命令，避免重复cd和source
	script := `
# 先切换到test-network目录并保持在其中
cd test-network

echo "=== Step 1: Stopping previous network ==="
./network.sh down

echo ""
echo "=== Step 2: Starting network ==="
./network.sh up

echo ""
echo "=== Step 3: Creating channel ==="
./network.sh createChannel

echo ""
echo "=== Step 4: Setting environment and deploying chaincode ==="
source set-env.sh
./network.sh deployCC -ccn novel-basic -ccp ../novel-resource-events -ccl go -ccv 1.0 -cci InitLedger -ccep 'OR("Org1MSP.member","Org2MSP.member")'

echo ""
echo "=== Step 5: Waiting for chaincode to be ready ==="
sleep 10

echo ""
echo "=== Step 6: Querying chaincode ==="
peer chaincode query -C mychannel -n novel-basic -c '{"function":"GetAllNovels","Args":[]}'

echo ""
echo "=== All tasks completed ==="
`

	// 创建单个shell命令来执行整个脚本
	cmd := exec.Command("sh", "-c", script)
	// 输出重定向，这就像把子程序的"嘴巴"连接到你的屏幕
	cmd.Stdout = os.Stdout  // 将整个过程输出显示到屏幕
	cmd.Stderr = os.Stderr  // 将错误信息显示到屏幕

	if err := cmd.Run(); err != nil {
		log.Printf("Script execution failed: %v", err)
	} else {
		fmt.Println("Script execution completed successfully")
	}
}