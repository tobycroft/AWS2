package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"main.go/config"
	"main.go/tuuz/Calc"
	"main.go/tuuz/Date"
	"main.go/tuuz/Jsong"
	"main.go/tuuz/Net"
	"sync"
	"time"
)

var Conn2User = make(map[*websocket.Conn]string)

var User2Conn2 sync.Map
var Conn2User2 sync.Map
var Room2 sync.Map

func On_connect(conn *websocket.Conn) {
	//err := conn.WriteMessage(1, []byte("连入成功"))
	message := map[string]interface{}{
		"remote_addr":  conn.RemoteAddr(),
		"connect_time": Date.Int2Date(time.Now().Unix()),
	}
	str := map[string]interface{}{
		"code": 0,
		"data": message,
		"type": "connected",
	}
	err := conn.WriteJSON(str)

	if err != nil {
		fmt.Printf("write fail = %v\n", err)
		return
	}
}

func On_close(conn *websocket.Conn) {
	On_exit(conn)
	// 发送 websocket 结束包
	conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 真正关闭 conn
	conn.Close()
}

func On_exit(conn *websocket.Conn) {
	uid, has := Conn2User2.Load(conn)
	if has {
		Room2.Delete(uid.(string))
		User2Conn2.Delete(uid.(string))
		Conn2User2.Delete(conn)
	}
}

func Handler(json_str string, conn *websocket.Conn) {
	if config.DEBUG {
		fmt.Println("json_ws:", json_str)
	}
	json, jerr := Jsong.JObject(json_str)
	if jerr != nil {
		fmt.Println("jsonerr", jerr)
		return
	}
	if config.DEBUG_WS_REQ {
		fmt.Println("DEBUG_WS_REQ", json_str)
	}
	data, derr := Jsong.ParseObject(json["data"])
	if derr != nil {
		fmt.Println("ws_derr:", derr)
		data = map[string]interface{}{}
		return
	}
	Type := Calc.Any2String(json["type"])
	switch Type {
	case "init", "INIT":
		auth_init(conn, data, Type)
		break

	case "join_room", "JOIN_ROOM":
		join_room(conn, data, Type)
		break

	case "exit_room", "EXIT_ROOM":
		exit_room(conn, data, Type)
		break

	case "msg_list", "MSG_LIST":
		msg_list(conn, data, Type)
		break

	case "private_msg", "PRIVATE_MSG":
		private_msg(conn, data, Type)
		break

	case "group_msg":
		group_msg(conn, data, Type)
		break

	case "requst_count":
		requst_count(conn, data, Type)
		break

	case "ping":
		ping(conn, data, Type)
		break

	case "api":
		api(conn, data, Type)
		break

	case "clear_private_unread":
		clear_private_unread(conn, data, Type)
		break

	case "clear_group_unread":
		clear_group_unread(conn, data, Type)
		break

	default:
		fmt.Println("undefine_type:", Type)
		break
	}
}

func auth_init(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid := Calc.Any2String(data["uid"])
	token := Calc.Any2String(data["token"])
	if uid == "" || token == "" {
		res := map[string]interface{}{
			"code": 400,
			"data": "uid&token",
			"type": Type,
		}
		if config.DEBUG {
			fmt.Println("auth_init", res)
		}
		conn.WriteJSON(res)
		On_close(conn)
		fmt.Println("uid_not_exists,UID-token不存在")
	}
	ret, err := Net.Post(config.CHAT_URL+config.AuthURL, nil, map[string]interface{}{
		"uid":   uid,
		"token": token,
		"type":  1,
		"ip":    conn.RemoteAddr(),
	}, nil, nil)
	if config.DEBUG_AUTH {
		fmt.Println("DEBUG_AUTH", ret, err)
	}
	if err != nil {
		res := map[string]interface{}{
			"code": 400,
			"data": "网络错误请重试",
			"type": Type,
		}
		conn.WriteJSON(res)
	} else {
		rtt, err := Jsong.JObject(ret)
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			if config.DEBUG {
				fmt.Println("DEBUG", ret, err)
			}
			conn.WriteJSON(res)
		} else {
			if config.DEBUG {
				fmt.Println(rtt)
			}
			if rtt["code"].(float64) == 0 {
				User2Conn2.Store(uid, conn)
				Conn2User2.Store(conn, uid)
				Room2.Store(uid, "0")
				message := "欢迎" + uid + "连入聊天服务器"
				if config.DEBUG {
					fmt.Println(message)
				}
				res := map[string]interface{}{
					"code":    0,
					"data":    "初始化完成",
					"message": message,
					"type":    Type,
				}
				if config.DEBUG_WS_RET {
					fmt.Println("DEBUG", res)
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": -1,
					"data": "未登录",
					"type": Type,
				}
				if config.DEBUG_WS_RET {
					fmt.Println("DEBUG", res)
				}
				conn.WriteJSON(res)
			}
		}
	}
}

func join_room(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		if data["chat_type"] == "private" {
			res := map[string]interface{}{
				"code": 0,
				"data": "已经加入和" + Calc.Any2String(data["id"]),
				"type": Type,
			}
			conn.WriteJSON(res)
		} else if data["chat_type"] == "group" {
			res := map[string]interface{}{
				"code": 0,
				"data": "已经加入和" + Calc.Any2String(data["id"]),
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			res := map[string]interface{}{
				"code": 400,
				"data": "类型不存在",
				"type": Type,
			}
			conn.WriteJSON(res)
		}
		Room2.Store(uid, Calc.Any2String(data["id"]))
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func exit_room(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		Room2.Store(uid, "0")
		res := map[string]interface{}{
			"code": 0,
			"data": "退出至大厅",
			"type": Type,
		}
		conn.WriteJSON(res)
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func msg_list(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		ret, err := Net.Post(config.CHAT_URL+config.Msg_list, nil, map[string]interface{}{
			"uid": uid,
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "消息列表数据不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}

}

func private_msg(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		ret, err := Net.Post(config.CHAT_URL+config.Private_msg, nil, map[string]interface{}{
			"uid": uid,
			"fid": data["uid"],
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "消息列表数据不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func group_msg(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		ret, err := Net.Post(config.CHAT_URL+config.Group_msg, nil, map[string]interface{}{
			"uid": uid,
			"fid": data["uid"],
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "消息列表数据不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func requst_count(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		ret, err := Net.Post(config.CHAT_URL+config.Request_count, nil, map[string]interface{}{
			"uid": uid,
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "Req_Count不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func ping(conn *websocket.Conn, data map[string]interface{}, Type string) {
	res := map[string]interface{}{
		"code": 0,
		"data": "PONG",
		"type": Type,
	}
	conn.WriteJSON(res)
}

func api(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		function := Calc.Any2String(data["func"])
		ret, err := Net.Post(config.CHAT_URL+config.ManualAPI+function, nil, map[string]interface{}{
			"uid": uid,
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "API不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func clear_private_unread(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		if data["id"] == nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "id",
				"type": Type,
			}
			conn.WriteJSON(res)
			return
		}
		ret, err := Net.Post(config.CHAT_URL+config.Clear_private_unread, nil, map[string]interface{}{
			"uid": uid,
			"fid": data["id"],
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "API不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}

func clear_group_unread(conn *websocket.Conn, data map[string]interface{}, Type string) {
	uid, has := Conn2User2.Load(conn)
	if has {
		if data["id"] == nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "id",
				"type": Type,
			}
			conn.WriteJSON(res)
			return
		}
		ret, err := Net.Post(config.CHAT_URL+config.Clear_group_unread, nil, map[string]interface{}{
			"uid": uid,
			"gid": data["id"],
			"ip":  conn.RemoteAddr(),
		}, nil, nil)
		if config.DEBUG_REMOTE_RET {
			fmt.Println("DEBUG_REMOTE_RET", ret, err)
		}
		if err != nil {
			res := map[string]interface{}{
				"code": 400,
				"data": "网络错误请重试",
				"type": Type,
			}
			conn.WriteJSON(res)
		} else {
			rtt, err := Jsong.JObject(ret)
			if err != nil {
				res := map[string]interface{}{
					"code": 404,
					"data": "API不完整",
					"type": Type,
				}
				conn.WriteJSON(res)
			} else {
				res := map[string]interface{}{
					"code": rtt["code"],
					"data": rtt["data"],
					"type": Type,
				}
				conn.WriteJSON(res)
			}
		}
	} else {
		conn.WriteJSON(map[string]interface{}{"code": -1, "data": "Auth_Fail", "type": Type})
	}
}
