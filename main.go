package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

// 设备列表 (使用提供的常量)
const (
	DeviceName1 = "866560088910415"
	DeviceName2 = "866560088599200"
)

func main() {
	log.Printf("==== OneNET Go 多设备模拟器启动 ====")
	log.Printf("产品ID: %s", ProductID)

	deviceNames := []string{DeviceName1, DeviceName2}
	
	var wg sync.WaitGroup // 用于等待所有设备协程结束
    
    // 初始化停止信号通道，用于通知所有设备协程退出
    stopSig := make(chan struct{})
	
	// 为每个设备启动一个独立的 Go 协程
	for _, name := range deviceNames {
		wg.Add(1)
		// 调用 runDeviceWithStop，并传入 wait group 和停止信号通道
		go runDeviceWithStop(name, &wg, stopSig) 
	}
    
    // --- 优雅退出机制 ---
    
    // 1. 设置信号监听
    quit := make(chan os.Signal, 1)
    // 监听 Ctrl+C (SIGINT) 和 kill (SIGTERM) 信号
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    // 2. 阻塞直到接收到信号
    sig := <-quit
    log.Printf("==== 收到系统信号: %v, 正在执行优雅退出... ====", sig)
    
    // 3. 关闭全局停止信号通道，通知所有设备协程退出
    close(stopSig)

    // 4. 等待所有设备协程完成退出
    log.Printf("等待所有设备断开 MQTT 连接...")
    
    waitTimeout := 5 * time.Second
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Printf("所有设备已成功断开连接。")
    case <-time.After(waitTimeout):
        log.Printf("⚠️ 警告: 超时 %v，强制退出程序。", waitTimeout)
    }

	log.Printf("==== OneNET Go 多设备模拟器退出完成 ====")
}

// runDeviceWithStop 负责单个设备的连接和主循环，支持优雅停止
func runDeviceWithStop(name string, wg *sync.WaitGroup, stop <-chan struct{}) {
    // 确保无论如何都通知 WaitGroup 退出
    defer wg.Done() 

    // 1. 获取 MQTT 连接配置
    opts := getConnectOptions(name)
    
    // 创建 Device 实例 (包含本地状态和静态属性)
    dev := initDeviceState(name) 

    // 设置连接成功回调：所有业务逻辑都在连接成功后执行
    opts.SetOnConnectHandler(func(client mqtt.Client) {
        log.Printf("[%s] MQTT 连接成功!", name)
        
        // 1. 设置 Client
        dev.Client = client 
        
        // 2. 订阅该设备专属的命令 Topic
        dev.subscribeForCommands() 

        // 3. 启动设备模拟的主循环 (包含动态上报和周期管理)
        go dev.startDeviceSimulation() 
    })
    
    // 设置连接丢失回调 
    opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
        log.Printf("[%s] MQTT 连接丢失: %v. 尝试重连...", name, err)
    })

    // 2. 创建并连接客户端
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Printf("[%s] 连接 Broker 失败: %v. 设备退出。", name, token.Error())
        return // 连接失败，退出协程
    }
    
    // 3. 阻塞协程，等待停止信号
    select {
    case <-stop:
        // 收到停止信号，执行优雅断开
        log.Printf("[%s] 正在断开 MQTT 连接...", name)
        // Disconnect(250) 允许 250ms 完成正在发送/接收的数据包
        client.Disconnect(250) 
        log.Printf("[%s] MQTT 连接已断开。", name)
    }
}