package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Device 结构体封装了每个设备的 MQTT 客户端、名称及本地状态
type Device struct {
	Name   string
	Client mqtt.Client

	// --- 本地属性状态 (对应物模型 int32 类型) ---
	OUT      int32
	Interval int32
	Relay    int32

	// --- 静态属性缓存 (只读属性只需计算一次) ---
	// 注意：这里缓存的是包装后的属性，用于 property/post
	StaticProps map[string]interface{}

	// --- 控制通道：用于发送信号 (如周期更新) 给 Runner ---
	controlChan chan int32
}

// ======================================================================
// Topic 模板定义
// ======================================================================
const (
	// 产品ID: 5S34OM4Rc6

	// ======================================================================
	// 🚀 发布 (设备 -> 云端)
	// ======================================================================
	PropertyPostTopicTemplate     = "$sys/5S34OM4Rc6/{device-name}/thing/property/post"      // 发布: 直连设备上报属性
	EventPostTopicTemplate        = "$sys/5S34OM4Rc6/{device-name}/thing/event/post"         // 发布: 直连设备上报事件
	PropertySetReplyTopicTemplate = "$sys/5S34OM4Rc6/{device-name}/thing/property/set_reply" // 发布: 直连设备属性设置响应
	PropertyGetReplyTopicTemplate = "$sys/5S34OM4Rc6/{device-name}/thing/property/get_reply" // 发布: 直连设备回复平台获取设备属性
	PackPostTopicTemplate         = "$sys/5S34OM4Rc6/{device-name}/thing/pack/post"          // 发布: 直连设备或子设备批量上报属性或事件

	// ======================================================================
	// ⬇️ 订阅 (云端 -> 设备)
	// ======================================================================
	PropertyPostReplyTopicTemplate = "$sys/5S34OM4Rc6/{device-name}/thing/property/post/reply" // 订阅: 直连设备上报属性响应
	EventPostReplyTopicTemplate    = "$sys/5S34OM4Rc6/{device-name}/thing/event/post/reply"    // 订阅: 直连设备上报事件响应
	PackPostReplyTopicTemplate     = "$sys/5S34OM4Rc6/{device-name}/thing/pack/post/reply"     // 订阅: 平台回复"设备批量上报属性或事件"
	PropertySetTopicTemplate       = "$sys/5S34OM4Rc6/{device-name}/thing/property/set"        // 订阅: 设置直连设备属性
	PropertyGetTopicTemplate       = "$sys/5S34OM4Rc6/{device-name}/thing/property/get"        // 订阅: 平台获取直连设备的属性
)

// EventReportFormat 定义事件上报格式
type EventReportFormat int

const (
	// 格式 1: 直接格式 (事件参数扁平，不包装) - event/post
	FormatDirect EventReportFormat = iota

	// 格式 2: 包装格式 (事件参数使用 {"value": ...} 包装) - event/post
	FormatWrapped

	// 格式 3: 批量格式 (使用 pack/post topic) - 推荐
	FormatBatch
)

// CurrentEventFormat 用于在不同格式之间切换测试
// 保持 FormatBatch 作为当前模式
var CurrentEventFormat = FormatDirect

// initDeviceState 初始化设备状态
func initDeviceState(deviceName string) *Device {
	rand.Seed(time.Now().UnixNano())
	dev := &Device{
		Name:        deviceName,
		OUT:         0,
		Relay:       0,
		Interval:    10,
		controlChan: make(chan int32, 1),
	}
	// 缓存包装后的静态属性 (用于 property/post)
	dev.StaticProps = dev.generateStaticProperties()
	return dev
}

// getTopic 根据设备名和模板获取最终的 Topic 字符串
func getTopic(deviceName string, template string) string {
	s := strings.ReplaceAll(template, "5S34OM4Rc6", ProductID)
	return strings.ReplaceAll(s, "{device-name}", deviceName)
}

// wrapValue 辅助函数：将值包装成 {"value": data} 标准格式 (用于 property/post)
func wrapValue(data interface{}) map[string]interface{} {
	return map[string]interface{}{"value": data}
}

// ======================================================================
// 属性数据生成 (Raw vs Wrapped)
// ======================================================================

// generateRawStaticProperties 模拟生成静态/只读属性数据 (返回原始值，用于 property/get_reply)
func (d *Device) generateRawStaticProperties() map[string]interface{} {
	// 原始属性值，不进行 wrapValue 包装
	properties := map[string]interface{}{
		"imsi": fmt.Sprintf("46000%d", rand.Int31n(999999999)),
		"cell_info": []string{
			fmt.Sprintf("LAC:%d", 1024+rand.Int31n(10)),
			fmt.Sprintf("CID:%d", 2048+rand.Int31n(100)),
		},
		"macs": []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"},

		// 结构体和数组的原始结构 (int32 类型保持 int32)
		"$OneNET_LBS": []map[string]interface{}{
			{
				"mnc": int32(1), "mcc": int32(460), "lac": int32(1024), "cid": int32(2048),
				"networkType": int32(2), "ss": int32(80 + rand.Int31n(20)),
				"signalLength": int32(100), "ta": int32(1), "flag": int32(1),
			},
		},
		"$OneNET_LBS_WIFI": map[string]interface{}{
			"imsi":     fmt.Sprintf("WIFI_IMSI_%d", rand.Int31n(1000)),
			"serverip": "", "macs": "", "mmac": "", "smac": "", "idfa": "",
		},
	}
	return properties
}

// generateRawDynamicProperties 模拟生成动态属性数据 (返回原始值，用于 property/get_reply)
func (d *Device) generateRawDynamicProperties() map[string]interface{} {
	// 原始属性值，不进行 wrapValue 包装 (int32 类型保持 int32)
	properties := map[string]interface{}{
		"temperature": rand.Int31n(40) + 10,
		"csq":         rand.Int31n(31),
		"OUT":         d.OUT,
		"relay":       d.Relay,
		"interval":    d.Interval,
	}
	return properties
}

// generateStaticProperties 模拟生成静态/只读属性数据 (用于 property/post，需要包装)
func (d *Device) generateStaticProperties() map[string]interface{} {
	rawProps := d.generateRawStaticProperties()
	wrappedProps := make(map[string]interface{})
	for k, v := range rawProps {
		wrappedProps[k] = wrapValue(v)
	}
	return wrappedProps
}

// generateDynamicProperties 模拟生成动态属性数据 (用于 property/post，需要包装)
func (d *Device) generateDynamicProperties() map[string]interface{} {
	rawProps := d.generateRawDynamicProperties()
	wrappedProps := make(map[string]interface{})
	for k, v := range rawProps {
		wrappedProps[k] = wrapValue(v)
	}
	return wrappedProps
}

// ======================================================================
// 上报逻辑
// ======================================================================

// postDeviceProperty 模拟设备上报属性
// 🚀 发布到: $sys/5S34OM4Rc6/{device-name}/thing/property/post
func (d *Device) postDeviceProperty(isFullReport bool) {
	postTopic := getTopic(d.Name, PropertyPostTopicTemplate)
	msgID := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)

	var properties map[string]interface{}

	if isFullReport {
		properties = make(map[string]interface{})
		// 静态属性使用缓存的包装值
		for k, v := range d.StaticProps {
			properties[k] = v
		}
		// 动态属性使用新生成的包装值
		for k, v := range d.generateDynamicProperties() {
			properties[k] = v
		}
		log.Printf("[%s] [全量上报] 包含所有属性", d.Name)
	} else {
		// 定时上报仅使用新生成的动态属性的包装值
		properties = d.generateDynamicProperties()
		properties["relay"] = wrapValue(d.Relay)
		log.Printf("[%s] [定时上报] 仅上报动态属性", d.Name)
	}

	// 结构: {"params": {"key": {"value": data}}}
	payloadStruct := map[string]interface{}{
		"id":      msgID,
		"version": "1.0",
		"params":  properties,
	}

	payloadBytes, _ := json.Marshal(payloadStruct)
	payload := string(payloadBytes)

	token := d.Client.Publish(postTopic, 1, false, payload)

	if token.Wait() && token.Error() != nil {
		log.Printf("[%s] 属性上报失败: %v", d.Name, token.Error())
	} else {
		log.Printf("[%s] ✅ 属性上报成功 (ID: %s)", d.Name, msgID)
	}
}

// postDeviceEvent 模拟设备上报事件 (支持3种格式)
// 🚀 发布到: $sys/5S34OM4Rc6/{device-name}/thing/event/post 或 thing/pack/post
func (d *Device) postDeviceEvent(eventID string) {
	msgID := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	alarmValue := rand.Int31n(2)

	var payload string
	var postTopic string
	var formatName string

	// 🚨 关键：事件参数必须是 int32 类型 (0 或 1)
	rawEventParams := map[string]interface{}{
		"powerOff":    int32(alarmValue),
		"overcurrent": int32(0),
		"smoke":       int32(0),
		"IN1":         int32(1),
		"IN2":         int32(0),
	}

	switch CurrentEventFormat {
	case FormatDirect:
		// 格式 1：直接格式 - event/post
		// 修正：事件数据需要嵌套在 "value" 键下，以解决 2409 错误 (required value:identifier:alarm)
		// 结构: {"id": "...", "version": "1.0", "params": {"alarm": {"value": {...}}}}
		
		// 🚨 修正点：将原始参数包在一个以 "value" 为键的 map 中，然后再包在 eventID 下
		wrappedEventData := map[string]interface{}{
			"value": rawEventParams,
		}
		
		nestedParams := map[string]interface{}{
			eventID: wrappedEventData, // 使用修正后的嵌套结构
		}

		payloadStruct := map[string]interface{}{
			"id":      msgID,
			"version": "1.0",
			"params":  nestedParams,
		}

		payloadBytes, _ := json.Marshal(payloadStruct)
		payload = string(payloadBytes)
		postTopic = getTopic(d.Name, EventPostTopicTemplate)
		formatName = "直接格式 (event/post)"

	case FormatWrapped:
		// 格式 2：包装格式 - event/post
		wrappedParams := map[string]interface{}{}
		for k, v := range rawEventParams {
			wrappedParams[k] = wrapValue(v) // 包装成 {"value": 1}
		}

		payloadStruct := map[string]interface{}{
			"id":      msgID,
			"version": "1.0",
			"params":  wrappedParams,
		}

		payloadBytes, _ := json.Marshal(payloadStruct)
		payload = string(payloadBytes)
		postTopic = getTopic(d.Name, EventPostTopicTemplate)
		formatName = "包装格式 (event/post)"

	case FormatBatch:
		// 格式 3：批量格式 (pack/post topic) - 之前遇到 2307 错误
		event := map[string]interface{}{
			"identifier": "alarm",
			"params":     rawEventParams,
			"time": time.Now().Unix(), 
		}

		// 结构: {"params": {"properties": {}, "events": [...]}}
		payloadStruct := map[string]interface{}{
			"id":      msgID,
			"version": "1.0",
			"params": map[string]interface{}{
				"properties": map[string]interface{}{},
				"events":     []map[string]interface{}{event},
			},
		}

		payloadBytes, _ := json.Marshal(payloadStruct)
		payload = string(payloadBytes)
		postTopic = getTopic(d.Name, PackPostTopicTemplate)
		formatName = "批量格式 (pack/post)"
	}

	log.Printf("[%s] 使用%s上报事件", d.Name, formatName)
	log.Printf("[%s] Topic: %s", d.Name, postTopic)
	log.Printf("[%s] Payload: %s", d.Name, payload)

	token := d.Client.Publish(postTopic, 1, false, payload)

	if token.Wait() && token.Error() != nil {
		log.Printf("[%s] 🔥 事件上报失败: %v", d.Name, token.Error())
	} else {
		log.Printf("[%s] 🔥 事件上报尝试成功 (ID: %s)", d.Name, msgID)
	}
}

// ======================================================================
// 平台命令和回复逻辑 (消息处理)
// ======================================================================

// handlePropertyPostReply 处理平台对属性上报的回复
// ⬇️ 订阅: $sys/5S34OM4Rc6/{device-name}/thing/property/post_reply
func (d *Device) handlePropertyPostReply(payload []byte) {
	var reply map[string]interface{}
	if err := json.Unmarshal(payload, &reply); err != nil {
		log.Printf("[%s] 解析属性上报回复失败: %v", d.Name, err)
		return
	}

	id := reply["id"]
	code, codeOk := reply["code"]
	msg := reply["msg"]

	if codeOk && code.(float64) == 200 {
		log.Printf("[%s] ✅ 属性上报已确认 (ID: %v, Code: 200)", d.Name, id)
	} else {
		log.Printf("[%s] ❌ 属性上报被拒绝! ID: %v, Code: %v, Msg: %v",
			d.Name, id, code, msg)
	}
}

// handleEventPostReply 处理平台对事件上报的回复
// ⬇️ 订阅: $sys/5S34OM4Rc6/{device-name}/thing/event/post_reply
func (d *Device) handleEventPostReply(payload []byte) {
	var reply map[string]interface{}
	if err := json.Unmarshal(payload, &reply); err != nil {
		log.Printf("[%s] 解析事件上报回复失败: %v", d.Name, err)
		return
	}

	id := reply["id"]
	code, codeOk := reply["code"]
	msg := reply["msg"]

	if codeOk && code.(float64) == 200 {
		log.Printf("[%s] ✅ 事件上报已确认 (ID: %v, Code: 200)", d.Name, id)
	} else {
		log.Printf("[%s] ❌ 事件上报被拒绝! ID: %v, Code: %v, Msg: %v",
			d.Name, id, code, msg)
	}
}

// handlePackPostReply 处理平台对批量上报的回复
// ⬇️ 订阅: $sys/5S34OM4Rc6/{device-name}/thing/pack/post_reply
func (d *Device) handlePackPostReply(payload []byte) {
	var reply map[string]interface{}
	if err := json.Unmarshal(payload, &reply); err != nil {
		log.Printf("[%s] 解析批量上报回复失败: %v", d.Name, err)
		return
	}

	id := reply["id"]
	code, codeOk := reply["code"]
	msg := reply["msg"]

	if codeOk && code.(float64) == 200 {
		log.Printf("[%s] ✅ 批量上报已确认 (ID: %v, Code: 200)", d.Name, id)
	} else {
		log.Printf("[%s] ❌ 批量上报被拒绝! ID: %v, Code: %v, Msg: %v",
			d.Name, id, code, msg)
	}
}

// createMessageHandler 集中处理所有下行消息
func createMessageHandler(dev *Device) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("[%s] ⬇️ 收到消息 Topic: %s", dev.Name, msg.Topic())

		setTopic := getTopic(dev.Name, PropertySetTopicTemplate)
		getTopicVar := getTopic(dev.Name, PropertyGetTopicTemplate)
		postReplyTopic := getTopic(dev.Name, PropertyPostReplyTopicTemplate)
		eventReplyTopic := getTopic(dev.Name, EventPostReplyTopicTemplate)
		packReplyTopic := getTopic(dev.Name, PackPostReplyTopicTemplate)

		switch msg.Topic() {
		case setTopic:
			dev.handlePropertySet(msg.Payload())
		case getTopicVar:
			dev.handlePropertyGet(msg.Payload())
		case postReplyTopic:
			dev.handlePropertyPostReply(msg.Payload())
		case eventReplyTopic:
			dev.handleEventPostReply(msg.Payload())
		case packReplyTopic:
			dev.handlePackPostReply(msg.Payload())
		default:
			log.Printf("[%s] 收到未知 Topic 消息，忽略", dev.Name)
		}
	}
}

// handlePropertySet 处理平台下发的属性设置命令
// ⬇️ 订阅: $sys/5S34OM4Rc6/{device-name}/thing/property/set
func (d *Device) handlePropertySet(payload []byte) {
	var req map[string]interface{}
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("[%s] 解析属性设置命令失败: %v", d.Name, err)
		return
	}

	msgID := req["id"]
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		log.Printf("[%s] 命令格式错误，缺少 params 字段", d.Name)
		return
	}

	log.Printf("[%s] 开始设置属性: %v", d.Name, params)

	for k, v := range params {
		if val, isFloat := v.(float64); isFloat {
			intValue := int32(val)

			switch k {
			case "OUT":
				d.OUT = intValue
				log.Printf("[%s] 成功设置 OUT = %d", d.Name, intValue)
			case "relay":
				d.Relay = intValue
				log.Printf("[%s] 成功设置 relay = %d", d.Name, intValue)
			case "interval":
				d.Interval = intValue
				log.Printf("[%s] 成功设置上报周期 interval = %d", d.Name, intValue)
				select {
				case d.controlChan <- d.Interval:
					log.Printf("[%s] 已发送周期更新信号: %d秒", d.Name, d.Interval)
				default:
				}
			default:
				log.Printf("[%s] 忽略未知可写属性: %s", d.Name, k)
			}
		}
	}

	// 🚀 发布回复到: $sys/5S34OM4Rc6/{device-name}/thing/property/set_reply
	replyTopic := getTopic(d.Name, PropertySetReplyTopicTemplate)
	replyPayloadStruct := map[string]interface{}{
		"id":      msgID,
		"version": "1.0",
		"code":    200,
		"msg":     "success",
	}

	replyPayloadBytes, _ := json.Marshal(replyPayloadStruct)

	token := d.Client.Publish(replyTopic, 1, false, string(replyPayloadBytes))
	if token.Wait() && token.Error() != nil {
		log.Printf("[%s] 属性设置回复失败: %v", d.Name, token.Error())
	} else {
		log.Printf("[%s] ⬆️ 已回复属性设置确认, ID: %v", d.Name, msgID)
		// 回复后立即上报最新状态
		d.postDeviceProperty(false)
	}
}

// handlePropertyGet 处理平台的属性查询命令
// ⬇️ 订阅: $sys/5S34OM4Rc6/{device-name}/thing/property/get
func (d *Device) handlePropertyGet(payload []byte) {
	var req map[string]interface{}
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("[%s] 解析属性获取命令失败: %v", d.Name, err)
		return
	}

	msgID := req["id"]

	// 🚨 关键修正：获取未包装的原始属性值 (用于回复平台查询)
	rawStaticProps := d.generateRawStaticProperties()
	rawDynamicProps := d.generateRawDynamicProperties()

	allRawProperties := make(map[string]interface{})
	for k, v := range rawStaticProps {
		allRawProperties[k] = v
	}
	for k, v := range rawDynamicProps {
		allRawProperties[k] = v
	}

	// 🚀 发布回复到: $sys/5S34OM4Rc6/{device-name}/thing/property/get_reply
	// 结构: {"data": {"key": data}}
	replyTopic := getTopic(d.Name, PropertyGetReplyTopicTemplate)
	replyPayloadStruct := map[string]interface{}{
		"id":      msgID,
		"version": "1.0",
		"code":    200,
		"msg":     "success",
		"data":    allRawProperties, // 使用不含 {"value":...} 包装的原始值
	}

	replyPayloadBytes, err := json.Marshal(replyPayloadStruct)
	if err != nil {
		log.Printf("[%s] 序列化属性获取回复失败: %v", d.Name, err)
		return
	}

	log.Printf("[%s] ⬆️ 已回复平台属性查询, ID: %v", d.Name, msgID)

	token := d.Client.Publish(replyTopic, 1, false, string(replyPayloadBytes))
	if token.Wait() && token.Error() != nil {
		log.Printf("[%s] 属性获取回复失败: %v", d.Name, token.Error())
	}
}

// subscribeForCommands 订阅所有下行 Topic 和平台回复 Topic
func (d *Device) subscribeForCommands() {
	setTopic := getTopic(d.Name, PropertySetTopicTemplate)
	getTopicVar := getTopic(d.Name, PropertyGetTopicTemplate)
	postReplyTopic := getTopic(d.Name, PropertyPostReplyTopicTemplate)
	eventReplyTopic := getTopic(d.Name, EventPostReplyTopicTemplate)
	packReplyTopic := getTopic(d.Name, PackPostReplyTopicTemplate)

	handler := createMessageHandler(d)

	tokenSet := d.Client.Subscribe(setTopic, 1, handler)
	tokenGet := d.Client.Subscribe(getTopicVar, 1, handler)
	tokenPostReply := d.Client.Subscribe(postReplyTopic, 1, handler)
	tokenEventReply := d.Client.Subscribe(eventReplyTopic, 1, handler)
	tokenPackReply := d.Client.Subscribe(packReplyTopic, 1, handler)

	if (tokenSet.Wait() && tokenSet.Error() != nil) || (tokenGet.Wait() && tokenGet.Error() != nil) ||
		(tokenPostReply.Wait() && tokenPostReply.Error() != nil) || (tokenEventReply.Wait() && tokenEventReply.Error() != nil) ||
		(tokenPackReply.Wait() && tokenPackReply.Error() != nil) {
		log.Fatalf("[%s] 命令订阅失败: Set(%v), Get(%v), PostReply(%v), EventReply(%v), PackReply(%v)",
			d.Name, tokenSet.Error(), tokenGet.Error(), tokenPostReply.Error(), tokenEventReply.Error(), tokenPackReply.Error())
	} else {
		log.Printf("[%s] 🔑 成功订阅所有 Topic (属性设置、属性查询、各种回复)", d.Name)
	}
}

// startDeviceSimulation 启动设备的主循环
func (d *Device) startDeviceSimulation() {
	log.Printf("设备 [%s] 开始运行 (事件格式模式: %v)", d.Name, CurrentEventFormat)

	if d.Client.IsConnected() {
		// 首次连接，全量上报属性
		d.postDeviceProperty(true)
	}

	go d.runRunner()
}

// runRunner 负责处理定时上报和周期更新逻辑
func (d *Device) runRunner() {
	currentInterval := d.Interval
	ticker := time.NewTicker(time.Duration(currentInterval) * time.Second)
	// 假设事件每 20 秒上报一次
	eventTicker := time.NewTicker(20 * time.Second)

	log.Printf("[%s] Runner 启动，属性上报周期: %d秒，事件上报周期: 20秒", d.Name, currentInterval)

	defer func() {
		log.Printf("[%s] Runner 停止", d.Name)
		ticker.Stop()
		eventTicker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if d.Client.IsConnected() {
				d.postDeviceProperty(false) // 定时上报动态属性
			}

		case <-eventTicker.C:
			if d.Client.IsConnected() {
				d.postDeviceEvent("alarm") // 上报事件
			}

		case newInterval := <-d.controlChan:
			if newInterval > 0 && newInterval != currentInterval {
				currentInterval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(time.Duration(currentInterval) * time.Second)
				log.Printf("[%s] 🔄 属性上报周期已更新为: %d秒", d.Name, currentInterval)
			}
		}
	}
}
