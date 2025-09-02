package network

import (
	"crypto/x509" // X.509证书解析，用于TLS连接验证
	"fmt"         // 格式化输出，用于错误处理
	"os"          // 操作系统接口，用于读取证书文件

	"github.com/hyperledger/fabric-gateway/pkg/identity" // Fabric网关身份验证包
	"google.golang.org/grpc"                           // gRPC客户端通信库
	"google.golang.org/grpc/credentials"                // gRPC凭证管理
)

func NewGrpcConnection() (*grpc.ClientConn, error) {
	//pem是最原始的证书，没有经过解析的证书，由start和end组成
	tlsCertificatePEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem")

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
	return grpc.NewClient("dns:///localhost:7051", grpc.WithTransportCredentials(transportCredentials))

}

// NewIdentity用于生成Fabric网络所需的X.509身份
func NewIdentity() *identity.X509Identity {
	//先读pem
	certificatePEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem")
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
	//还是先拿pem
	privateKeyPEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk")
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

