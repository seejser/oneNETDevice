# 通过 MQTT 和 oneNET 通信（oneNETDevice）

基于[中国移动物联网云平台](https://iot.10086.cn)的传感器项目测试

## 运行

1. 先填写配置
2. 再运行

```go
go mod tidy

go run ./
```

## userinfo

### 完整物模型

```json
{
  "version": "1.0",
  "profile": {
    "industryId": "1",
    "sceneId": "18",
    "categoryId": "93",
    "productId": "5S34OM4Rc6"
  },
  "properties": [
    {
      "identifier": "$OneNET_LBS",
      "name": "基站定位",
      "functionType": "s",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "array",
        "specs": {
          "length": 3,
          "type": "struct",
          "specs": [
            {
              "name": "移动网号",
              "identifier": "mnc",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "移动国家号码",
              "identifier": "mcc",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "地区区域码",
              "identifier": "lac",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "基站码",
              "identifier": "cid",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "网络制式",
              "identifier": "networkType",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "信号强度",
              "identifier": "ss",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "当前基站广播信号强度",
              "identifier": "signalLength",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "移动台距以确定其发往基站的定时超前量",
              "identifier": "ta",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            },
            {
              "name": "基站信息数字进制",
              "identifier": "flag",
              "dataType": {
                "type": "int32",
                "specs": {
                  "max": "2147483647",
                  "min": "-2147483648",
                  "step": "",
                  "unit": ""
                }
              }
            }
          ]
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "$OneNET_LBS_WIFI",
      "name": "WiFi定位",
      "functionType": "s",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "struct",
        "specs": [
          {
            "name": "移动用户识别码",
            "identifier": "imsi",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          },
          {
            "name": "设备接入基站时对应的网关ip",
            "identifier": "serverip",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          },
          {
            "name": "可以接收到的热点mac信息",
            "identifier": "macs",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          },
          {
            "name": "已连热点mac信息",
            "identifier": "mmac",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          },
          {
            "name": "手机mac码",
            "identifier": "smac",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          },
          {
            "name": "IOS手机的idfa",
            "identifier": "idfa",
            "dataType": {
              "type": "string",
              "specs": {
                "length": 255
              }
            }
          }
        ]
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "OUT",
      "name": "OUT_J9输出控制",
      "functionType": "u",
      "accessMode": "rw",
      "desc": "",
      "dataType": {
        "type": "int32",
        "specs": {
          "max": "1",
          "min": "0",
          "step": "",
          "unit": ""
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "cell_info",
      "name": "获取小区基站信息",
      "functionType": "u",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "array",
        "specs": {
          "length": 10,
          "type": "string",
          "specs": {
            "length": 256
          }
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "csq",
      "name": "信号质量",
      "functionType": "u",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "int32",
        "specs": {
          "max": "31",
          "min": "0",
          "step": "",
          "unit": ""
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "imsi",
      "name": "sim卡号",
      "functionType": "u",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "string",
        "specs": {
          "length": 50
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "interval",
      "name": "上报周期",
      "functionType": "u",
      "accessMode": "rw",
      "desc": "",
      "dataType": {
        "type": "int32",
        "specs": {
          "max": "65535",
          "min": "1",
          "step": "",
          "unit": ""
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "macs",
      "name": "获取MAC地址",
      "functionType": "u",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "array",
        "specs": {
          "length": 10,
          "type": "string",
          "specs": {
            "length": 256
          }
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "relay",
      "name": "控制继电器开关",
      "functionType": "u",
      "accessMode": "rw",
      "desc": "0关继电器 1 开继电器",
      "dataType": {
        "type": "int32",
        "specs": {
          "max": "1",
          "min": "0",
          "step": "",
          "unit": ""
        }
      },
      "functionMode": "property",
      "required": false
    },
    {
      "identifier": "temperature",
      "name": "温度",
      "functionType": "u",
      "accessMode": "r",
      "desc": "",
      "dataType": {
        "type": "int32",
        "specs": {
          "max": "125",
          "min": "-55",
          "step": "",
          "unit": ""
        }
      },
      "functionMode": "property",
      "required": false
    }
  ],
  "events": [
    {
      "identifier": "alarm",
      "name": "预警事件",
      "functionType": "u",
      "eventType": "alert",
      "desc": "",
      "outputData": [
        {
          "identifier": "powerOff",
          "name": "断电",
          "dataType": {
            "type": "int32",
            "specs": {
              "max": "1",
              "min": "0",
              "step": "",
              "unit": ""
            }
          }
        },
        {
          "identifier": "overcurrent",
          "name": "过流",
          "dataType": {
            "type": "int32",
            "specs": {
              "max": "1",
              "min": "0",
              "step": "",
              "unit": ""
            }
          }
        },
        {
          "identifier": "smoke",
          "name": "烟雾报警",
          "dataType": {
            "type": "int32",
            "specs": {
              "max": "1",
              "min": "0",
              "step": "",
              "unit": ""
            }
          }
        },
        {
          "identifier": "IN1",
          "name": "IN1检测IO",
          "dataType": {
            "type": "int32",
            "specs": {
              "max": "1",
              "min": "0",
              "step": "",
              "unit": ""
            }
          }
        },
        {
          "identifier": "IN2",
          "name": "IN2检测IO",
          "dataType": {
            "type": "int32",
            "specs": {
              "max": "1",
              "min": "0",
              "step": "",
              "unit": ""
            }
          }
        }
      ],
      "functionMode": "event",
      "required": false
    }
  ],
  "services": [],
  "combs": []
}
```

### 物模型 topic

$sys/5S34OM4Rc6/{device-name}/thing/property/post
发布
直连设备上报属性
$sys/5S34OM4Rc6/{device-name}/thing/property/post/reply
订阅
直连设备上报属性响应
$sys/5S34OM4Rc6/{device-name}/thing/event/post
发布
直连设备上报事件
$sys/5S34OM4Rc6/{device-name}/thing/event/post/reply
订阅
直连设备上报事件响应
$sys/5S34OM4Rc6/{device-name}/thing/property/set
订阅
设置直连设备属性
$sys/5S34OM4Rc6/{device-name}/thing/property/set_reply
发布
直连设备属性设置响应
$sys/5S34OM4Rc6/{device-name}/thing/sub/property/set
订阅
设置子设备属性
$sys/5S34OM4Rc6/{device-name}/thing/sub/property/set_reply
发布
子设备属性设置响应
$sys/5S34OM4Rc6/{device-name}/thing/property/desired/get
发布
直连设备获取期望值
$sys/5S34OM4Rc6/{device-name}/thing/property/desired/get/reply
订阅
直连设备获取期望值响应
$sys/5S34OM4Rc6/{device-name}/thing/property/desired/delete
发布
直连设备清除期望值
$sys/5S34OM4Rc6/{device-name}/thing/property/desired/delete/reply
订阅
直连设备清除期望值响应
$sys/5S34OM4Rc6/{device-name}/thing/property/get
订阅
平台获取直连设备的属性
$sys/5S34OM4Rc6/{device-name}/thing/property/get_reply
发布
直连设备回复平台获取设备属性
$sys/5S34OM4Rc6/{device-name}/thing/sub/property/get
订阅
平台获取子设备属性
$sys/5S34OM4Rc6/{device-name}/thing/sub/property/get_reply
发布
设备回复获取子设备属性
$sys/5S34OM4Rc6/{device-name}/thing/service/{identifier}/invoke
订阅
平台调用直连设备服务(下发数据并期望设备执行完成后给出响应)
$sys/5S34OM4Rc6/{device-name}/thing/service/{identifier}/invoke_reply
发布
直连设备回复“平台调用设备服务”
$sys/5S34OM4Rc6/{device-name}/thing/sub/service/invoke
订阅
平台调用子设备服务
$sys/5S34OM4Rc6/{device-name}/thing/sub/service/invoke_reply
发布
子设备回复"调用子设备服务"
$sys/5S34OM4Rc6/{device-name}/thing/pack/post
发布
直连设备或子设备批量上报属性或事件
$sys/5S34OM4Rc6/{device-name}/thing/pack/post/reply
订阅
平台回复"设备批量上报属性或事件"
$sys/5S34OM4Rc6/{device-name}/thing/history/post
发布
直连设备或子设备上报历史数据
$sys/5S34OM4Rc6/{device-name}/thing/history/post/reply
订阅
平台回复"设备上报历史数据"
