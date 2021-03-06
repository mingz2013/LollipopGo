package main

import (
	"LollipopGo/LollipopGo/log"
	"Proto"
	"Proto/Proto2"
	"flag"
	"fmt"
	_ "fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	"LollipopGo/LollipopGo/util"
	_ "LollipopGo/ReadCSV"

	_ "LollipopGo/LollipopGo/player"

	"code.google.com/p/go.net/websocket"
)

var addrDSQ = flag.String("addrDSQ", "127.0.0.1:8888", "http service address") // 链接gateway
var ConnDSQ *websocket.Conn                                                    // 保存用户的链接信息，数据会在主动匹配成功后进行链接
var ConnDSQRPC *rpc.Client                                                     // 链接DB server
var DSQAllMap map[string]*RoomPlayerDSQ                                        // 游戏逻辑存储
var DSQ_qi = []int{                                                            // 1-8 A ;9-16 B
	Proto2.Elephant, Proto2.Lion, Proto2.Tiger, Proto2.Leopard, Proto2.Wolf, Proto2.Dog, Proto2.Cat, Proto2.Mouse,
	Proto2.Mouse + Proto2.Elephant, Proto2.Mouse + Proto2.Lion, Proto2.Mouse + Proto2.Tiger, Proto2.Mouse + Proto2.Leopard,
	Proto2.Mouse + Proto2.Wolf, Proto2.Mouse + Proto2.Dog, Proto2.Mouse + Proto2.Cat, 2 * Proto2.Mouse}

// 斗兽棋游戏结构
// 每个房间都存在一个
type RoomPlayerDSQ struct {
	OpenIDA   string
	OpenIDB   string
	InitData  [4][4]int // 斗兽棋的棋盘的数据
	WhoChuPai string    // 当前谁出牌
	GoAround  int       // 回合，如果每人出10次都没有吃子，系统推送平局;第七个回合提示数据 第10局平局
}

/*

	-----------------------------------------
	|										|
	|	[0,0]01	[1,0]02	[2,0]03	[3,0]04		|
	|										|
	|										|
	|	[0,1]05	[1,1]06	[2,1]07	[3,1]08		|
	|										|
	|										|
	|	[0,2]09	[1,2]10	[2,2]11	[3,2]12		|
	|									    |
	|										|
	|	[0,3]13	[1,3]14	[2,3]15	[3,3]16		|
	|										|
	-----------------------------------------

*/

// 初始化操作
func init() {
	// if !initDSQGateWayNet() {
	// 	fmt.Println("链接 gateway server 失败!")
	// 	return
	// }
	// fmt.Println("链接 gateway server 成功!")
	// // 初始化数据
	// initDSQNetRPC()
	return
}

// 初始化RPC
func initDSQNetRPC() {
	client, err := jsonrpc.Dial("tcp", service)
	if err != nil {
		log.Debug("dial error:", err)
	}
	ConnDSQRPC = client
}

// 初始化网关
func initDSQGateWayNet() bool {

	fmt.Println("用户客户端客户端模拟！")
	log.Debug("用户客户端客户端模拟！")
	url := "ws://" + *addrDSQ + "/GolangLtd"
	conn, err := websocket.Dial(url, "", "test://golang/")
	if err != nil {
		fmt.Println("err:", err.Error())
		return false
	}
	ConnDSQ = conn
	// 协程支持  --接受线程操作 全球协议操作
	go GameServerReceiveDSQ(Conn)
	// 发送链接的协议 ---》
	initConnDSQ(Conn)
	return true
}

// 链接到网关
func initConnDSQ(conn *websocket.Conn) {
	// 协议修改
	data := &Proto2.G2GW_ConnServer{
		Protocol:  Proto.G_GameGlobal_Proto, // 游戏主要协议
		Protocol2: Proto2.G2GW_ConnServerProto2,
		ServerID:  util.MD5_LollipopGO("8895" + "DSQ server"),
	}
	fmt.Println(data)
	// 2 发送数据到服务器
	PlayerSendToServer(conn, data)
	return
}

// 处理数据
func GameServerReceiveDSQ(ws *websocket.Conn) {
	for {
		var content string
		err := websocket.Message.Receive(ws, &content)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		// decode
		fmt.Println(strings.Trim("", "\""))
		fmt.Println(content)
		content = strings.Replace(content, "\"", "", -1)
		contentstr, errr := base64Decode([]byte(content))
		if errr != nil {
			fmt.Println(errr)
			continue
		}
		// 解析数据 --
		fmt.Println("返回数据：", string(contentstr))
		go SyncMeassgeFunDSQ(string(contentstr))
	}
}

// 链接分发 处理
func SyncMeassgeFunDSQ(content string) {
	var r Requestbody
	r.req = content

	if ProtocolData, err := r.Json2map(); err == nil {
		// 处理我们的函数
		HandleCltProtocolDSQ(ProtocolData["Protocol"], ProtocolData["Protocol2"], ProtocolData)
	} else {
		log.Debug("解析失败：", err.Error())
	}
}

//  主协议处理
func HandleCltProtocolDSQ(protocol interface{}, protocol2 interface{}, ProtocolData map[string]interface{}) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			strerr := fmt.Sprintf("%s", err)
			//发消息给客户端
			ErrorST := Proto2.G_Error_All{
				Protocol:  Proto.G_Error_Proto,      // 主协议
				Protocol2: Proto2.G_Error_All_Proto, // 子协议
				ErrCode:   "80006",
				ErrMsg:    "亲，您发的数据的格式不对！" + strerr,
			}
			// 发送给玩家数据
			fmt.Println("Global server的主协议!!!", ErrorST)
		}
	}()

	// 协议处理
	switch protocol {
	case float64(Proto.G_GameGlobal_Proto):
		{ // Global Server 主要协议处理
			fmt.Println("Global server 主协议!!!")
			HandleCltProtocol2DSQ(protocol2, ProtocolData)

		}
	default:
		panic("主协议：不存在！！！")
	}
	return
}

// 子协议的处理
func HandleCltProtocol2DSQ(protocol2 interface{}, ProtocolData map[string]interface{}) {

	switch protocol2 {
	case float64(Proto2.GW2G_ConnServerProto2):
		{ // 网关返回数据
			fmt.Println("gateway server 返回给global server 数据信息！！！")
		}
	case float64(Proto2.G2GW_PlayerEntryHallProto2):
		{ // 网关请求获取大厅数据
			fmt.Println("玩家请求获取大厅数：默认获奖列表、跑马灯等")
			// G2GW_PlayerEntryHallProto2Fucn(Conn, ProtocolData)
		}

	default:
		panic("子协议：不存在！！！")
	}
	return
}

//------------------------------------------------------------------------------
// 初始化牌型
func InitDSQ(data1 []int) [4][4]int {

	data, erdata, j, k := data1, [4][4]int{}, 0, 0

	for i := 0; i < Proto2.Mouse*2; i++ {
		icount := util.RandInterval_LollipopGo(0, int32(len(data))-1)
		fmt.Println("随机数：", icount)
		if len(data) == 1 {
			erdata[3][3] = data[0]
		} else {
			//------------------------------------------------------------------
			if int(icount) < len(data) {
				erdata[j][k] = data[icount]
				k++
				if k%4 == 0 {
					j++
					k = 0
				}
				data = append(data[:icount], data[icount+1:]...)
			} else {
				erdata[j][k] = data[icount]
				k++
				if k%4 == 0 {
					j++
					k = 0
				}
				data = data[:icount-1]
			}
			//------------------------------------------------------------------
		}
		fmt.Println("生成的数据", erdata)
	}

	return erdata
}

// 判断棋子大小
// 玩家操作移动,操作协议
func CheckIsEat(fangx int, qizi int, qipan [4][4]int) (bool, int, [4][4]int) {
	if qizi > 16 || qizi < 1 {
		log.Debug("玩家发送棋子数据不对！")
		return false, Proto2.DATANOEXIT, qipan
	}
	// 1 寻找 玩家的棋子在棋牌的位置/或者这个棋子是否存在
	bret, Posx, posy := CheckChessIsExit(qizi, qipan)
	if bret {
		bret, iret, data := CheckArea(fangx, Posx, posy, qipan)
		return bret, iret, data
	} else {
		log.Debug("玩家发送棋子不存在！疑似外挂。")
		return false, Proto2.DATAERROR, qipan
	}

	return true, 0, qipan
}

// 检查棋盘中是不是存在
func CheckChessIsExit(qizi int, qipan [4][4]int) (bool, int, int) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if qipan[i][j] == qizi {
				return true, i, j
			}
		}
	}
	return false, 0, 0
}

// 边界判断
func CheckArea(fangx, iposx, iposy int, qipan [4][4]int) (bool, int, [4][4]int) {

	if fangx == Proto2.UP {
		if iposy == 0 {
			return false, Proto2.MOVEFAIL, qipan // 无法移动
		}
		data_yidong := qipan[iposx][iposy-1]
		data := qipan[iposx][iposy]
		// 判断是空地不
		if data_yidong == 0 {
			data_ret := UpdateChessData(Proto2.MOVE, iposx, iposy-1, iposx, iposy, qipan)
			return true, Proto2.MOVESUCC, data_ret // 移动成功
		}
		// 对比棋子大小
		if data < 9 {
			if data_yidong < 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		} else {
			if data_yidong > 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx, iposy-1, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		}

	} else if fangx == Proto2.DOWN {

		if iposy == 3 {
			return false, Proto2.MOVEFAIL, qipan // 无法移动
		}
		data_yidong := qipan[iposx][iposy+1]
		data := qipan[iposx][iposy]
		// 判断是空地不
		if data_yidong == 0 {
			data_ret := UpdateChessData(Proto2.MOVE, iposx, iposy+1, iposx, iposy, qipan)
			return true, Proto2.MOVESUCC, data_ret // 移动成功
		}
		// 对比棋子大小
		if data < 9 {
			if data_yidong < 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		} else {
			if data_yidong > 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx, iposy+1, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		}

	} else if fangx == Proto2.LEFT {
		if iposx == 0 {
			return false, Proto2.MOVEFAIL, qipan // 无法移动
		}
		data_yidong := qipan[iposx-1][iposy]
		data := qipan[iposx][iposy]
		// 判断是空地不
		if data_yidong == 0 {
			data_ret := UpdateChessData(Proto2.MOVE, iposx-1, iposy, iposx, iposy, qipan)
			return true, Proto2.MOVESUCC, data_ret // 移动成功
		}
		// 对比棋子大小
		if data < 9 {
			if data_yidong < 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		} else {
			if data_yidong > 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx-1, iposy, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		}

	} else if fangx == Proto2.RIGHT {
		if iposx == 3 {
			return false, Proto2.MOVEFAIL, qipan // 无法移动
		}
		data_yidong := qipan[iposx+1][iposy]
		data := qipan[iposx][iposy]
		// 判断是空地不
		if data_yidong == 0 {
			data_ret := UpdateChessData(Proto2.MOVE, iposx+1, iposy, iposx, iposy, qipan)
			return true, Proto2.MOVESUCC, data_ret // 移动成功
		}
		// 对比棋子大小
		if data < 9 {
			if data_yidong < 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		} else {
			if data_yidong > 9 {
				return false, Proto2.TEAMMATE, qipan // 自己人
			} else {
				if data_yidong > data {
					data_ret := UpdateChessData(Proto2.DISAPPEAR, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.DISAPPEAR, data_ret // 自残
				} else if data_yidong == data {
					data_ret := UpdateChessData(Proto2.ALLDISAPPEAR, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.ALLDISAPPEAR, data_ret // 同归于尽
				} else if data_yidong < data {
					data_ret := UpdateChessData(Proto2.BEAT, iposx+1, iposy, iposx, iposy, qipan)
					return true, Proto2.BEAT, data_ret // 吃掉对方
				}
			}
		}
	}

	return false, Proto2.ITYPEINIY, qipan // 不存在的操作
}

// 更新棋盘数据
// 注：移动成功后，原来位置如何变化
//    目标的位置如何变化
//    fangxPos fangyPos表示变化的位置
//    iPosx iPosy 原来棋子的位置
func UpdateChessData(iType, fangxPos, fangyPos, iPosx, iPosy int, qipan [4][4]int) [4][4]int {
	// 保存副本数据
	data := qipan
	// 原来棋子的数据
	yd_data := data[iPosx][iPosy]
	//  更新棋盘 数据
	if iType == Proto2.MOVE { // 更新空地 ,1 更新原来棋盘的位置为0， 目标的位置为现在数据
		data[iPosx][iPosy] = 0
		data[fangxPos][fangyPos] = yd_data
	} else if iType == Proto2.DISAPPEAR { // 自残 ,1 更新原来棋盘的位置为0，目标的位置数据不变
		data[iPosx][iPosy] = 0
	} else if iType == Proto2.ALLDISAPPEAR { // 同归于尽 ,1 更新原来棋盘的位置为0，目标的位置数据为0
		data[iPosx][iPosy] = 0
		data[fangxPos][fangyPos] = 0
	} else if iType == Proto2.BEAT { // 吃掉对方 ,1 更新原来位置为0，目标位置为现在数据
		data[iPosx][iPosy] = 0
		data[fangxPos][fangyPos] = yd_data
	}
	return data
}
