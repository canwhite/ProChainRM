package network

import (
	"crypto/x509" // X.509证书解析，用于TLS连接验证
	"fmt"         // 格式化输出，用于错误处理
	"log"         // 日志输出
	"os"          // 操作系统接口，用于读取证书文件

	"github.com/hyperledger/fabric-gateway/pkg/identity" // Fabric网关身份验证包
	"google.golang.org/grpc"                             // gRPC客户端通信库
	"google.golang.org/grpc/credentials"                 // gRPC凭证管理
)

func NewGrpcConnection() (*grpc.ClientConn, error) {
	// 获取Fabric证书路径
	certPath := os.Getenv("FABRIC_CERT_PATH")
	if certPath == "" {
		certPath = "../test-network/organizations/peerOrganizations/org1.example.com" // 默认路径
	}

	// 获取Fabric连接地址
	peerHost := os.Getenv("FABRIC_PEER_HOST")
	peerPort := os.Getenv("FABRIC_PEER_PORT")

	// 智能判断运行环境
	if peerHost == "" {
		// 检测是否在Docker容器中运行（通过检查/proc/1/cgroup或其他方式）
		if _, err := os.Stat("/.dockerenv"); err == nil {
			// 在Docker容器中，使用host.docker.internal
			peerHost = "host.docker.internal"
			log.Printf("检测到Docker环境，使用 host.docker.internal")
		} else {
			// 本地开发环境，检测是否MacOS或Linux
			// 在MacOS上，localhost可以工作
			peerHost = "localhost"
			log.Printf("检测到本地环境，使用 localhost")
		}
	}

	if peerPort == "" {
		peerPort = "7051" // 默认端口
	}

	//pem是最原始的证书，没有经过解析的证书，由start和end组成
	tlsCertificatePEM, err := os.ReadFile(fmt.Sprintf("%s/tlsca/tlsca.org1.example.com-cert.pem", certPath))

	if err != nil {
		return nil, fmt.Errorf("failed to read TLS certificate: %w", err)
	}

	//将pem格式解析为x509证书对象
	tlsCertificate, err := identity.CertificateFromPEM(tlsCertificatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TLS certificate: %w", err)
	}

	//创建证书池，得到通信证书
	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCertificate)
	//credentials是凭证，用于验证服务器证书
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	//注意返回的是grpc的client对象，供gateway使用
	peerAddress := fmt.Sprintf("%s:%s", peerHost, peerPort)
	log.Printf("连接Fabric网络: %s", peerAddress)
	return grpc.NewClient(peerAddress, grpc.WithTransportCredentials(transportCredentials))

}

// NewIdentity用于生成Fabric网络所需的X.509身份
func NewIdentity() *identity.X509Identity {
	// 获取Fabric证书路径
	certPath := os.Getenv("FABRIC_CERT_PATH")
	if certPath == "" {
		certPath = "../test-network/organizations/peerOrganizations/org1.example.com" // 默认路径
	}

	//先读pem
	certificatePEM, err := os.ReadFile(fmt.Sprintf("%s/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem", certPath))
	if err != nil {

		panic(fmt.Errorf("failed to read certificate: %w", err))
	}

	//将pem格式解析为x509证书对象,
	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(fmt.Errorf("failed to parse certificate: %w", err))
	}
	//创建身份对象
	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		panic(fmt.Errorf("failed to create identity: %w", err))
	}

	return id
}

func NewSign() identity.Sign {
	// 获取Fabric证书路径
	certPath := os.Getenv("FABRIC_CERT_PATH")
	if certPath == "" {
		certPath = "../test-network/organizations/peerOrganizations/org1.example.com" // 默认路径
	}

	//还是先拿pem
	privateKeyPEM, err := os.ReadFile(fmt.Sprintf("%s/users/User1@org1.example.com/msp/keystore/priv_sk", certPath))
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}

	//解析pem，搞到x509证书对象
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(fmt.Errorf("failed to parse private key: %w", err))
	}

	//创建签名函数
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(fmt.Errorf("failed to create sign function: %w", err))
	}

	return sign
}

/**
反正都需要一个x509证书;
then一个通过pool new一个NewClientTLSFromCert
一个new一个NewX509Identity
一个new一个NewPrivateKeySign
*/
