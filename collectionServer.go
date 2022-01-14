package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	JSON_SUCCESS int = 1
	JSON_ERROR   int = 0
)

type (
	todoLog struct {
		ID        uint   `gorm:"primarykey"`
		CreatedAt int64  `gorm:"autoCreateTime"`
		Logger    string `json:"logger"`
		Time      string `json:"time"`
		Level     string `json:"level"`
		Log       string `json:"log"`
	}
)

func (todoLog) TableName() string {
	return "collectionServer"
}

var db *gorm.DB

// 初始化
func init() {
	var err error
	//var constr string
	//constr = fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "123456", "localhost", 3306, "test")
	dsn := "root:123456@tcp(127.0.0.1:3306)/collectionServer?charset=utf8mb4&parseTime=True&loc=Local"
	//db, err = gorm.Open("mysql", constr)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: nil})
	if err != nil {
		panic("数据库连接失败")
	}
	err = db.AutoMigrate(&todoLog{})
	if err != nil {
		return
	}
}

func main() {
	go socket()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	r := gin.Default()
	v1 := r.Group("/api/v1/todo")
	{
		v1.GET("/", all)        // 查询所有条目
		v1.GET("/total", total) //查询条目总数
	}
	err := r.Run(":9089")
	if err != nil {
		return
	}

}

func socket() {
	listen, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("Listen() failed, err: ", err)
		return
	}
	for {
		conn, err := listen.Accept() // 监听客户端的连接请求
		if err != nil {
			fmt.Println("Accept() failed, err: ", err)
			continue
		}
		go Process(conn) // 启动一个goroutine来处理客户端的连接请求
		//fmt.Println(conn)
	}
}

func all(c *gin.Context) {
	DB := db
	var todos []todoLog
	var totalSize int64
	Order := c.Query("order")
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	pageNum, _ := strconv.Atoi(c.Query("pageNum"))
	beginTime, _ := strconv.Atoi(c.Query("beginTime"))
	endTime, _ := strconv.Atoi(c.Query("endTime"))

	if ID, _ := strconv.Atoi(c.Query("id")); ID != 0 {
		if Order != "asc" {
			DB = DB.Where("id <= ?", ID)
		}
	}

	if beginTime != 0 && endTime != 0 {
		DB = DB.Where("created_at BETWEEN ? AND ?", beginTime/1000, endTime/1000)
	}

	if Log := c.Query("log"); Log != "" {
		DB = DB.Where("log LIKE ?", "%"+Log+"%")
	}
	if Level := c.Query("level"); Level != "" {
		DB = DB.Where("level = ?", Level)
	}
	if Logger := c.Query("file"); Logger != "" {
		DB = DB.Where("logger = ?", Logger)
	}
	DB.Model(&todoLog{}).Count(&totalSize)
	if pageNum > 0 && pageSize > 0 {
		DB = DB.Limit(pageSize).Offset((pageNum - 1) * pageSize)
	}
	if Order == "desc" {
		if err := DB.Order("id desc").Find(&todos).Error; err != nil {
			fmt.Println(err.Error())
		}
	} else {
		if err := DB.Order("id asc").Find(&todos).Error; err != nil {
			fmt.Println(err.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    todos,
		"total":   totalSize,
	})
}

func total(c *gin.Context) {
	DB := db
	var totalSize int64
	DB.Model(&todoLog{}).Count(&totalSize)
	c.JSON(http.StatusOK, gin.H{
		"total":   totalSize,
		"message": "ok",
	})
}

func Process(conn net.Conn) {
	defer conn.Close()
	//ticker := time.NewTicker(time.Second * 1)
	//log := make([]todoLog, 0)
	//go func() {
	//	for {
	//		//select {
	//		//case <-ticker.C:
	//			//length := 0
	//			//if len(log) > 50 {
	//			//    length = 50
	//			//} else {
	//			//    length = len(log)
	//			//}
	//			//if length == 0 {
	//			//	continue
	//			//}
	//			//temp := log[:length]
	//			db.CreateInBatches(&log, 100)
	//			//log = make([]todoLog, 0)
	//			//fmt.Println(log)
	//		//}
	//	}
	//}()
	for {
		reader := bufio.NewReader(conn)
		msg, err := Decode(reader)
		//fmt.Println(msg)
		if err != nil {
			return
		}
		arr := strings.Fields(msg)
		//fmt.Println(msg)
		//fmt.Println(arr[0], arr[1], arr[2], arr[3])
		//ticker = time.NewTicker(time.Second * 5)
		//log = append(log, todoLog{Logger: arr[2][1 : len(arr[2])-1], Time: arr[0][1:] + " " + arr[1][:len(arr[1])-1], Level: arr[3][1 : len(arr[3])-1], Log: msg[len(arr[0])+len(arr[1])+len(arr[2])+len(arr[3])+4:]})
		//fmt.Println("3")
		go db.Create(&todoLog{Logger: arr[2][1 : len(arr[2])-1], Time: arr[0][1:] + " " + arr[1][:len(arr[1])-1], Level: arr[3][1 : len(arr[3])-1], Log: msg[len(arr[0])+len(arr[1])+len(arr[2])+len(arr[3])+4:]})
		//fmt.Println(conn, cnt)
		//if err == io.EOF {
		//	return
		//}
		//fmt.Println(db)
		//fmt.Printf("%v %v %v %v\n", arr[0][1:]+" "+arr[1][:len(arr[1])-1], arr[2][1:len(arr[2])-1], arr[3][1:len(arr[3])-1], msg[len(arr[0])+len(arr[1])+len(arr[2])+len(arr[3])+4:])
	}
}

func Decode(reader *bufio.Reader) (string, error) {
	// 读消息长度
	lengthByte, _ := reader.Peek(4)
	lengthBuff := bytes.NewBuffer(lengthByte)

	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	// buffer返回缓冲中现有的可读的字节数
	if int32(reader.Buffered()) < length+4 {
		return "", errors.New("not enough")
	}
	// 读取真正的数据

	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return "", err
	}
	//fmt.Println(n)
	return string(pack[4:]), nil
}
