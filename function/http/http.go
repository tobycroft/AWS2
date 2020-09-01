package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"main.go/config"
	"main.go/function/ws"
	"main.go/tuuz/Calc"
	"main.go/tuuz/Input"
	"main.go/tuuz/Jsong"
)

func Handler(c *gin.Context) {
	key, ok := c.GetPostForm("key")
	if !ok {
		c.JSON(200, map[string]interface{}{
			"code": 400,
			"data": "key",
		})
		fmt.Println("key")
		c.Abort()
		return
	}
	if key != config.KEY {
		c.JSON(200, map[string]interface{}{
			"code": 403,
			"data": "key_error",
		})
		c.Abort()
		return
	}
	uids, ok := c.GetPostForm("uids")
	if !ok {
		c.JSON(200, map[string]interface{}{
			"code": 400,
			"data": "需要uids来发送数据",
		})
		fmt.Println("uids")
		c.Abort()
		return
	}

	to_users, err := Jsong.JArray(uids)
	if err != nil {
		c.JSON(200, map[string]interface{}{
			"code": 400,
			"data": "to_users_err",
		})
		fmt.Println("to_user")
		c.Abort()
		return
	}
	dest, ok := Input.Post("dest", c, false)
	if !ok {
		fmt.Println("dest")
		return
	}
	data, ok := Input.Post("data", c, false)
	if !ok {
		fmt.Println("data")
		return
	}
	Type, ok := c.GetPostForm("type")
	if !ok {
		c.JSON(200, map[string]interface{}{
			"code": 400,
			"data": "需要type类型,chat/message_list/system",
		})
		fmt.Println("type")
		c.Abort()
		return
	}
	json, jerr := Jsong.JObject(data)
	if jerr != nil {
		c.JSON(200, map[string]interface{}{
			"code": 400,
			"data": "data需要提交",
		})
		c.Abort()
		return
	}
	json_handler(c, json, to_users, dest, Type)
}

func json_handler(c *gin.Context, json map[string]interface{}, to_users []interface{}, dest string, Type string) {
	fmt.Println("json_http", json)
	uids := []interface{}{}
	uidf := []interface{}{}
	data := map[string]interface{}{
		"code": 0,
		"data": json,
		"type": Type,
	}
	switch Type {
	case "system":
		for _, uid := range to_users {
			conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
			if has {
				uids = append(uids, uid)
				conn.(*websocket.Conn).WriteJSON(data)
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "refresh_list":
		for _, uid := range to_users {
			conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
			if has {
				uids = append(uids, uid)
				conn.(*websocket.Conn).WriteJSON(data)
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "private_chat":
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == dest {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					uids = append(uids, uid)
					conn.(*websocket.Conn).WriteJSON(data)
				}
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "group_chat":
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == dest {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					conn.(*websocket.Conn).WriteJSON(data)
				}
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "request_count":
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == "0" {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					uids = append(uids, uid)
					conn.(*websocket.Conn).WriteJSON(data)
				}
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "push":
		for _, uid := range to_users {
			conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
			if has {
				uids = append(uids, uid)
				conn.(*websocket.Conn).WriteJSON(data)
			} else {
				uidf = append(uidf, uid)
			}
		}
		break

	case "message":
		for _, uid := range to_users {
			conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
			if has {
				uids = append(uids, uid)
				conn.(*websocket.Conn).WriteJSON(data)
			}
		}
		break

	case "indoor_message":
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == "0" {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					uids = append(uids, uid)
					conn.(*websocket.Conn).WriteJSON(data)
				}
			}
		}
		break

	case "outer_all":
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == "0" {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					uids = append(uids, uid)
					conn.(*websocket.Conn).WriteJSON(data)
				}
			}
		}
		break

	//case "user_room":
	//	user_room := map[string]string{}
	//	for _, uid := range to_users {
	//		id := Calc.Any2String(uid)
	//		user_room[id] = ws.Room[id]
	//	}
	//	c.JSON(200, map[string]interface{}{
	//		"code": 0,
	//		"data": user_room,
	//	})
	//	break

	default:
		fmt.Println("default")
		for _, uid := range to_users {
			room, has := ws.Room2.Load(Calc.Any2String(uid))
			if has && room.(string) == dest {
				conn, has := ws.User2Conn2.Load(Calc.Any2String(uid))
				if has {
					uids = append(uids, uid)
					conn.(*websocket.Conn).WriteJSON(data)
				}
			} else {
				uidf = append(uidf, uid)
			}
		}
		break
	}
	resp := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"success": uids,
			"fail":    uidf,
		},
	}
	if config.DEBUG_SEND_RET {
		fmt.Println("resp:", resp)
	}
	c.JSON(200, resp)
	c.Abort()
	return

}
