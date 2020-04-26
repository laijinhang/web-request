package main

import (
    "net"
    "github.com/lunny/log"
    "encoding/json"
    "time"
    "os"
    "net/http"
    "sync"
    "strconv"
    "fmt"
)

type master struct {
    ip string
    port string
    conn net.Conn
}

var (
    SBCNum     int           // 并发连接数
    QPSNum     int           // 总请求次数
    RTNum      time.Duration // 响应时间
    RTTNum     time.Duration // 平均响应时间
    SuccessNum int           // 成功次数
    FailNum    int           // 失败次数
    SecNum     int

    mt master
    err error
    mu sync.Mutex // 必须加锁
    RQNum int    // 最大并发数，由命令行传入
    Url   string // url，由命令行传入
)

func init() {
    if len(os.Args) != 5 {
        log.Fatalf("%s 并发数 url ip port", os.Args[0])
    }
    RQNum, err = strconv.Atoi(os.Args[1])
    if err != nil {
        log.Println(err)
    }
    Url = os.Args[2]
    mt.ip = os.Args[3]
    mt.port = os.Args[4]
}

func main() {
    mt.conn, err = net.Dial("tcp", mt.ip + ":" + mt.port)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("连接服务器成功。。。")
    _, err = mt.conn.Write([]byte(Url))
    if err != nil {
        log.Println(err)
    }
    go func() {
        for range time.Tick(1 * time.Second) {
            sendToMaster(mt, map[string]interface{}{
                "SBCNum": SBCNum,           // 并发连接数
                "RTNum": RTNum,             // 响应时间
                "SecNum": SecNum,           // 时间
                "SuccessNum": SuccessNum,   // 成功次数
                "FailNum": FailNum,         // 失败次数
            })
        }
    }()
    go func() {
        for range time.Tick(1 * time.Second) {
            SecNum++
        }
    }()
    requite(RQNum, Url)
}

func requite(RQNum int, url string) {
    c := make(chan int)
    for i := 0;i < RQNum;i++ {
        SBCNum = i + 1
        go func(url string) {
            var tb time.Time
            var el time.Duration
            for {
                tb = time.Now()
                _, err := http.Get(url)
                if err == nil {
                    el = time.Since(tb)
                    mu.Lock() // 上锁
                    SuccessNum++
                    RTNum += el
                    mu.Unlock() // 解锁
                } else {
                    mu.Lock() // 上锁
                    FailNum++
                    mu.Unlock() // 解锁
                }
                time.Sleep(1 * time.Second)
            }
        }(url)
        time.Sleep(45 * time.Millisecond)
    }
    <- c    // 阻塞
}

func sendToMaster(mt master, data map[string]interface{}) {
    r, err := json.Marshal(data)
    if err != nil {
        log.Println(err)
    }
    _, err = mt.conn.Write(r)
    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
}
