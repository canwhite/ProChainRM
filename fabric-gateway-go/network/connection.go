package network // 网络包 - 处理Hyperledger Fabric网络的连接和身份验证

import (
	"crypto/x509" // X.509证书解析，用于TLS连接验证
	"fmt"         // 格式化输出，用于错误处理
	"os"          // 操作系统接口，用于读取证书文件

	"github.com/hyperledger/fabric-gateway/pkg/identity" // Fabric网关身份验证包
	"google.golang.org/grpc"                           // gRPC客户端通信库
	"google.golang.org/grpc/credentials"                // gRPC凭证管理
)

// NewGrpcConnection 创建与Fabric网关的gRPC客户端连接
// 因为需要先有证书才能建立连接，所以我们先获取证书
func NewGrpcConnection() (*grpc.ClientConn, error) {
	// 加载peer0.org1.example.com的TLS证书，该证书用于建立与Fabric节点的安全连接
	// PEM（Privacy Enhanced Mail）是一种用于存储和传输加密数据的文本编码格式，
	// 常用于保存证书（如X.509证书）、私钥、公钥等。PEM格式以ASCII编码，
	// 内容被包裹在"-----BEGIN ...-----"和"-----END ...-----"之间。例如：
	// -----BEGIN CERTIFICATE-----
	// （Base64编码的证书内容）
	// -----END CERTIFICATE-----
	// 在Hyperledger Fabric中，TLS证书和身份证书通常以PEM格式存储和分发。
	tlsCertificatePEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS certificate: %w", err)
	}

	// 将PEM格式的证书内容解析为X.509证书对象
	// x509是一种国际标准的公钥基础设施（PKI）证书格式，常用于SSL/TLS通信中的身份认证和加密。x509证书通常采用ASN.1格式编码，包含公钥、持有者信息、签发者信息、有效期、用途、数字签名等内容。Hyperledger Fabric等区块链系统广泛使用x509证书来标识和认证网络中的各类实体（如用户、节点、组织），以保证通信安全和身份可信。
	tlsCertificate, err := identity.CertificateFromPEM(tlsCertificatePEM)
	if err != nil {
		// 如果证书解析失败，返回错误信息
		return nil, fmt.Errorf("failed to parse TLS certificate: %w", err)
	}

	// 创建新的证书池，用于存储可信证书
	certPool := x509.NewCertPool()
	// 将TLS证书添加到可信证书池
	certPool.AddCert(tlsCertificate)
	// 创建客户端TLS凭证，使用证书池验证服务器证书，"peer0.org1.example.com"是预期的服务器名称，用于SNI验证
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	// 建立到Fabric网关的gRPC连接
	// 连接到本地7051端口（peer0.org1.example.com的标准端口）
	// dns:///前缀支持DNS服务发现
	return grpc.NewClient("dns:///localhost:7051", grpc.WithTransportCredentials(transportCredentials))
}


// NewIdentity 为网关连接创建基于X.509证书的客户端身份
// 该身份用于在Fabric网络中进行身份验证
func NewIdentity() *identity.X509Identity {
	// 加载User1@org1.example.com的客户端证书
	// 该证书用于证明客户端应用的身份
	// 证书位于MSP（成员服务提供者）目录结构中
	certificatePEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem")
	if err != nil {
		// 如果读取证书失败，程序panic
		// 因为身份创建对应用启动至关重要
		panic(fmt.Errorf("failed to read certificate: %w", err))
	}

	// 将PEM格式的客户端证书解析为X.509证书对象
	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		// 如果证书解析失败，程序panic
		panic(fmt.Errorf("failed to parse certificate: %w", err))
	}

	// 创建X.509身份对象，使用解析的证书
	// "Org1MSP"是Fabric网络中组织1的成员服务提供者ID
	// 此身份用于验证交易和查询
	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		// 如果身份创建失败，程序panic
		panic(fmt.Errorf("failed to create identity: %w", err))
	}

	// 返回创建好的身份对象，供Fabric网关使用
	return id
}

// NewSign 创建用于数字签名的函数
// 该函数使用私钥为消息摘要生成数字签名，确保交易完整性和不可抵赖性
func NewSign() identity.Sign {
	// 加载与User1@org1.example.com关联的私钥
	// 该私钥用于为交易创建数字签名
	// 私钥安全存储在MSP密钥库目录中
	privateKeyPEM, err := os.ReadFile("../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk")
	if err != nil {
		// 如果读取私钥失败，程序panic
		// 因为签名能力对交易提交至关重要
		panic(fmt.Errorf("failed to read private key: %w", err))
	}

	// 将PEM格式的私钥内容解析为私钥对象
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		// 如果私钥解析失败，程序panic
		panic(fmt.Errorf("failed to parse private key: %w", err))
	}

	// 创建签名函数，使用私钥进行加密签名
	// 此函数用于签署交易提案和消息
	// 签名证明了与X.509证书对应的私钥所有权
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		// 如果创建签名函数失败，程序panic
		panic(fmt.Errorf("failed to create sign function: %w", err))
	}

	// 返回签名函数，供Fabric网关操作使用
	return sign
}