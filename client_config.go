package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls" 
	"encoding/base64"
	"fmt"
	"hash"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

// OneNET 平台配置 (使用提供的常量)
const (
	ProductID   = ""                                   // 产品ID
	AccessKey   = "" // 产品的 Access Key
	
	// Broker URL (使用 MQTTS)
	BrokerURL   = "ssl://5S34OM4Rc6.mqttstls.acc.cmcconenet.cn:8883" 
	
	// Token 算法配置
	AuthVersion    = "2018-10-31" 
	AuthMethod     = "sha1" 
	KeepAlive      = 60 * time.Second
	ExpiryDuration = 1 * time.Hour 
)

// OneNET_Sign 计算 OneNET 平台要求的签名 (Sign)
func OneNET_Sign(key, stringForSignature, method string) (string, error) {
	// 1. Key 参与计算前应先进行 base64decode
	rawKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("base64 decode key failed: %w", err)
	}

	// 2. 选择 HMAC 算法的 Hash 构造函数
	var hmacNew func() hash.Hash
	switch strings.ToLower(method) {
	case "md5":
		hmacNew = md5.New
	case "sha1":
		hmacNew = sha1.New
	case "sha256":
		hmacNew = sha256.New
	default:
		return "", fmt.Errorf("unsupported signature method: %s", method)
	}

	// 3. 计算 HMAC 签名
	h := hmac.New(hmacNew, rawKey)
	if _, err := h.Write([]byte(stringForSignature)); err != nil {
		return "", fmt.Errorf("hmac write failed: %w", err)
	}

	// 4. 对签名结果进行 Base64 编码
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// getOneNETToken 构造完整的 Token 字符串 (Password)
func getOneNETToken(productID, deviceName, accessKey, method, version string) (string, error) {
	// 设备级别鉴权 res 格式
	res := fmt.Sprintf("products/%s/devices/%s", productID, deviceName)
	et := time.Now().Add(ExpiryDuration).Unix()

	// 1. 构造 StringForSignature (et, method, res, version 顺序，以 '\n' 分隔)
	stringForSignature := fmt.Sprintf("%d\n%s\n%s\n%s", et, method, res, version)

	// 2. 计算签名 sign
	sign, err := OneNET_Sign(accessKey, stringForSignature, method)
	if err != nil {
		return "", err
	}

	// 3. 构造 Token Map
	tokenParams := map[string]string{
		"version": version,
		"res":     res,
		"et":      fmt.Sprintf("%d", et),
		"method":  method,
		"sign":    sign,
	}

	// 4. 排序、URL 编码并连接
	var keys []string
	for k := range tokenParams {
		keys = append(keys, k)
	}
	sort.Strings(keys) 

	var encodedParams []string
	for _, k := range keys {
		encodedValue := url.QueryEscape(tokenParams[k])
		encodedParams = append(encodedParams, fmt.Sprintf("%s=%s", k, encodedValue))
	}

	return strings.Join(encodedParams, "&"), nil
}

// getConnectOptions 构造 MQTT 连接选项
func getConnectOptions(deviceName string) *mqtt.ClientOptions {
	// --- 1. 构造认证 Token ---
	token, err := getOneNETToken(ProductID, deviceName, AccessKey, AuthMethod, AuthVersion)
	if err != nil {
		log.Fatalf("生成 OneNET Token 失败: %v", err)
	}

	// --- 2. 构造 MQTT Options ---
	opts := mqtt.NewClientOptions().AddBroker(BrokerURL)
	
	// MQTT 认证配置
	opts.SetClientID(deviceName) 
	opts.SetUsername(ProductID)   
	opts.SetPassword(token)    

	opts.SetKeepAlive(KeepAlive)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(true)
    
    // --- 3. 禁用 TLS 证书校验 (解决 x509 错误，仅用于测试) ---
    if strings.HasPrefix(BrokerURL, "ssl") {
        tlsConfig := &tls.Config{
            // 警告：跳过证书验证，生产环境不推荐
            InsecureSkipVerify: true, 
            ClientAuth:         tls.NoClientCert,
        }
        opts.SetTLSConfig(tlsConfig)
        log.Printf("[%s] 警告: MQTTS 证书验证已禁用 (InsecureSkipVerify=true)。", deviceName)
    }
    
    // 连接丢失处理
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("[%s] MQTT 连接丢失: %v", deviceName, err)
	})

	return opts
}