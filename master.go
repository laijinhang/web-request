package main

import (
    "log"
    "net"
    "time"
    "os"
    "encoding/json"
    "fmt"
)

var ip string
var port string
var slaves []*slave

type slave struct {
    UserId string
    SBCNum     int           // 并发连接数
    QPSNum     int           // 总请求次数
    RTNum      time.Duration // 响应时间
    RTTNum     time.Duration // 响应时间
    SecNum     int           // 时间
    SuccessNum int           // 成功次数
    FailNum    int           // 失败次数
    Url string
    conn net.Conn
}

func (s *slave) Run() {
    var v interface{}
    buf := make([]byte, 1024)
    for {
        n, err := s.conn.Read(buf)
        if err != nil {
            log.Println(err)
            break
        }
        err = json.Unmarshal(buf[:n], &v)
        if err != nil {
            log.Println(err)
            continue
        }
        s.SBCNum = int(v.(map[string]interface{})["SBCNum"].(float64))  // 并发连接数
        s.RTNum = time.Duration(v.(map[string]interface{})["RTNum"].(float64))  // 响应时间
        s.SuccessNum = int(v.(map[string]interface{})["SuccessNum"].(float64))  //SuccessNum int           // 成功次数
        s.FailNum = int(v.(map[string]interface{})["FailNum"].(float64))            //FailNum    int           // 失败次数
        s.SecNum = int(v.(map[string]interface{})["SecNum"].(float64))
    }
}

func init() {
    if len(os.Args) != 3 {
        log.Fatal(os.Args[0] + " ip port")
    }
    ip = os.Args[1]
    port = os.Args[2]
}

func main() {
    s, err := net.Listen("tcp", ip + ":" + port)
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()
    buf := make([]byte, 128)
    fmt.Println("Run...")
    go func() {
        for range time.Tick(2 * time.Second) {
            show(slaves)
        }
    }()
    for {
        conn, err := s.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        n, err := conn.Read(buf)
        tempC := slave{conn:conn,UserId:conn.RemoteAddr().String(), Url:string(buf[:n])}
        go tempC.Run()
        slaves = append(slaves, &tempC)
    }
}

func show(clients []*slave) {
    if len(clients) == 0 {
        return
    }
    temp := slave{}
    num := 0
    for _, client := range clients {
        if client.SecNum == 0 {
            continue
        }
        num++
        fmt.Printf("用户id：%s,url: %s,并发数：%d,请求次数：%d,平均响应时间：%s,成功次数：%d,失败次数：%d\n",
            client.UserId,
            client.Url,
            client.SBCNum,
            client.SuccessNum + client.FailNum,
            client.RTNum / (time.Duration(client.SecNum) * time.Second),
            client.SuccessNum,
            client.FailNum)
        temp.SBCNum += client.SBCNum
        temp.RTNum += client.RTNum / (time.Duration(client.SecNum) * time.Second)
        temp.SecNum += client.SecNum
        temp.SuccessNum += client.SuccessNum
        temp.FailNum += client.FailNum
    }
    if num == 0 {
        return
    }
    fmt.Printf("并发数：%d,请求次数：%d,平均响应时间：%s,成功次数：%d,失败次数：%d\n",
        temp.SBCNum,
        temp.SuccessNum + temp.FailNum,
        temp.RTNum / time.Duration(num),
        temp.SuccessNum,
        temp.FailNum)
    fmt.Println()
}

func heartbeat(clients []slave) []slave {   // 标记耦合
    tempC := []slave{}
    for _, client := range clients {
        _, err := client.conn.Write([]byte(""))
        if err == nil { // 删除
            tempC = append(tempC, client)
        }
    }
    return tempC
}
