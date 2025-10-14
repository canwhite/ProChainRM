package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// RSACrypto RSA加密解密器
type RSACrypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSACrypto 创建RSA加密解密器
func NewRSACrypto() (*RSACrypto, error) {
	// 从环境变量或文件读取私钥和公钥
	privateKeyPEM := getPrivateKey()
	publicKeyPEM := getPublicKey()
	
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("未找到RSA私钥配置")
	}
	
	// 解析私钥
	// 这里要拿一个block，是因为PEM格式的密钥文件本质上是ASCII文本，需要通过pem.Decode将其解码出block结构体。
	// block.Bytes字段才是真正的DER编码的密钥字节内容，只有这样才能后续用x509等标准库进行密钥解析（ParsePKCS1PrivateKey等）。
	// 如果直接用PEM字符串解析，会失败，因此必须先decode成block，然后拿里面的Bytes出来继续操作。

	// 公钥其实同理，也必须通过pem.Decode获取block，然后用block.Bytes做解析。
	// 这里只是结构不同：私钥用ParsePKCS1PrivateKey, 公钥根据其格式用ParsePKIXPublicKey或ParsePKCS1PublicKey解析。
	// 本质上都是先拿block，然后用block.Bytes进一步解析。
	privateKeyBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("私钥PEM格式错误")
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}
	
	// 解析公钥
	var publicKey *rsa.PublicKey
	if publicKeyPEM != "" {
		log.Printf("🔍 开始解析公钥，PEM长度: %d", len(publicKeyPEM))
		publicKeyBlock, _ := pem.Decode([]byte(publicKeyPEM))
		if publicKeyBlock == nil {
			log.Printf("警告: 公钥PEM解码失败")
		} else {
			log.Printf("🔍 PEM解码成功，类型: %s，数据长度: %d", publicKeyBlock.Type, len(publicKeyBlock.Bytes))
			pub, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
			if err != nil {
				log.Printf("警告: PKIX公钥解析失败，尝试PKCS1格式: %v", err)
				// 尝试用PKCS1格式解析
				pubKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
				if err != nil {
					log.Printf("警告: PKCS1公钥解析也失败: %v", err)
				} else {
					publicKey = pubKey
					log.Printf("🔍 PKCS1解析成功，N长度: %d bits，E: %d", publicKey.N.BitLen(), publicKey.E)
				}
			} else {
				publicKey = pub.(*rsa.PublicKey)
				log.Printf("🔍 PKIX解析成功，N长度: %d bits，E: %d", publicKey.N.BitLen(), publicKey.E)
			}
		}
	}
	
	// 如果没有单独的公钥文件，使用私钥中的公钥
	if publicKey == nil {
		log.Printf("信息: 使用私钥中的公钥")
		publicKey = &privateKey.PublicKey
		log.Printf("信息: 私钥中的公钥 - N length: %d bits, E: %d", publicKey.N.BitLen(), publicKey.E)
	}
	
	// 最终检查
	if publicKey == nil {
		return nil, fmt.Errorf("无法获取公钥")
	}
	
	return &RSACrypto{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Decrypt 解密数据
func (r *RSACrypto) Decrypt(encryptedBase64 string) (string, error) {
	if r.privateKey == nil {
		return "", fmt.Errorf("私钥未初始化")
	}
	
	// Base64解码
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}
	
	// 使用RSA-OAEP解密，使用SHA-1哈希与Next.js保持一致
	decrypted, err := rsa.DecryptOAEP(sha1.New(), nil, r.privateKey, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}
	
	return string(decrypted), nil
}

// Encrypt 加密数据
func (r *RSACrypto) Encrypt(data string) (string, error) {
	log.Printf("🔐 Encrypt方法被调用，r.publicKey指针: %p", r.publicKey)
	if r.publicKey == nil {
		return "", fmt.Errorf("公钥未初始化")
	}
	
	log.Printf("🔐 开始加密数据，长度: %d", len(data))
	// 使用RSA-OAEP加密，使用SHA-1哈希与Next.js保持一致
	// nil 参数会导致空指针异常，必须使用 crypto/rand.Reader
	encrypted, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, r.publicKey, []byte(data), nil)
	if err != nil {
		return "", fmt.Errorf("RSA加密失败: %v", err)
	}
	
	log.Printf("🔐 加密成功，结果长度: %d", len(encrypted))
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// getPrivateKey 获取私钥
func getPrivateKey() string {
	// 优先从环境变量获取
	if key := os.Getenv("RSA_PRIVATE_KEY"); key != "" {
		return key
	}
	
	// 尝试从文件读取
	if data, err := os.ReadFile("security/rsa_private_key.pem"); err == nil {
		return string(data)
	}
	
	// 尝试从备选路径读取
	if data, err := os.ReadFile("../security/rsa_private_key.pem"); err == nil {
		return string(data)
	}
	
	log.Println("警告: 未找到RSA私钥，请配置RSA_PRIVATE_KEY环境变量或security/rsa_private_key.pem文件")
	return ""
}

// getPublicKey 获取公钥
func getPublicKey() string {
	// 优先从环境变量获取
	if key := os.Getenv("RSA_PUBLIC_KEY"); key != "" {
		return key
	}
	
	// 尝试从文件读取
	if data, err := os.ReadFile("security/rsa_public_key.pem"); err == nil {
		log.Printf("✅ 从 security/rsa_public_key.pem 读取公钥成功，大小: %d bytes", len(data))
		return string(data)
	}
	
	// 尝试从备选路径读取
	if data, err := os.ReadFile("../security/rsa_public_key.pem"); err == nil {
		log.Printf("✅ 从 ../security/rsa_public_key.pem 读取公钥成功，大小: %d bytes", len(data))
		return string(data)
	}
	
	log.Printf("❌ 未找到公钥文件")
	return ""
}

// 全局RSA加密解密器实例
var globalRSACrypto *RSACrypto

// InitRSACrypto 初始化全局RSA加密解密器
func InitRSACrypto() error {
	if globalRSACrypto != nil {
		log.Printf("信息: RSA加密解解密器已经初始化，跳过")
		return nil
	}
	
	var err error
	globalRSACrypto, err = NewRSACrypto()
	if err != nil {
		log.Printf("❌ RSA加密解密器初始化失败: %v", err)
		return err
	}
	
	log.Printf("✅ RSA加密解密器初始化成功")
	return err
}

// DecryptWithRSA 使用全局RSA解密器解密数据
func DecryptWithRSA(encryptedBase64 string) (string, error) {
	if globalRSACrypto == nil {
		return "", fmt.Errorf("RSA加密解密器未初始化")
	}
	return globalRSACrypto.Decrypt(encryptedBase64)
}

// EncryptWithRSA 使用全局RSA加密器加密数据
func EncryptWithRSA(data string) (string, error) {
	if globalRSACrypto == nil {
		return "", fmt.Errorf("RSA加密解密器未初始化")
	}
	return globalRSACrypto.Encrypt(data)
}