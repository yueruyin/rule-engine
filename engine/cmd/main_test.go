package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/sjson"
	"golang.org/x/crypto/bcrypt"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"zenith.engine.com/engine/internal/handler"
	"zenith.engine.com/engine/pkg/util"
)

func TestComputeFuncResult1(t *testing.T) {
	//r := util.ComputeFuncResult1("if(0-0.4474<=0.35||21.0000=0,30,if(0-0.4474<=0.3,0.4474*100*0.8,if(0-0.4474<=0.2,0.4474*100*0.7,if(0-0.4474<=0.1,0.4474*100*0.6,if(0-0.4474<=0,0.4474*100*0.5,2)))))", 10)
	r, _ := util.ComputeFuncResult1("max([1,2,3,4,5,6])", 0, true)

	t.Logf(r)
}

func TestComputeResultBool(t *testing.T) {
	b := util.ComputeResultBool("1==1")
	t.Log(b)
}

func TestFormatHasArgumentsJson(t *testing.T) {
	j := util.FormatHasArgumentsJson("{\"a\":\"测试\",$d:123}")
	t.Logf(j)
}

func TestReplaceLong(t *testing.T) {
	l := util.ReplaceLong("{ descText: \"我有一个参数 11111111111111111111111 ssssfff f\"}")
	ln := util.ReplaceLong("{descText: 1812371231239837412}")
	t.Log(l)
	t.Log(ln)
}

func TestRound(t *testing.T) {
	t.Log(util.Round(1.337, 2))
}

func TestInt(t *testing.T) {
	t.Log(util.Int(2.9))
}

func TestIsDate(t *testing.T) {
	t.Log(util.IsDate("2022-06-16"))
}

func TestQueue(t *testing.T) {
	var que util.Queue
	que.Push("2")
	que.Push("1")
	que.Push("0")
	que.Push("2")
	que.Push("1")
	que.Push("0")
	que.Push("2")
	que.Push("1")
	que.Push("0")
	que.Push("2")
	que.Push("1")
	que.Push("0")
	for !que.QueueIsEmpty() {
		t.Log(que.Pop())
	}
}

func TestIf2(t *testing.T) {
	v := util.IF2("'较差'=='较差1'||'较差'=='较差1'||'较差'=='较差'", true, false)
	t.Log(v)

}

func GoSay(s string) {
	for i := 0; i < 2; i++ {
		runtime.Gosched()
		fmt.Println(s)
	}
}

func TestSpinLock(t *testing.T) {
	runtime.GOMAXPROCS(1)
	go GoSay("h")
	GoSay("w")
}

var done bool

func read(name string, c *sync.Cond) {
	c.L.Lock()
	if !done {
		c.Wait()
	}
	fmt.Println(name, "starts reading")
	c.L.Unlock()
}

func write(c *sync.Cond) {
	time.Sleep(time.Second * 3)
	done = true
	c.Broadcast()
}

func TestCond(t *testing.T) {
	cond := sync.NewCond(&sync.Mutex{})
	go read("pool 1", cond)
	go read("pool 2", cond)
	go read("pool 3", cond)
	go read("pool 4", cond)
	write(cond)

	time.Sleep(time.Second * 2)
}

func TestAnts(t *testing.T) {
	pool, _ := ants.NewPool(runtime.NumCPU() * 100)
	defer pool.Release()
	for i := 0; i < 10; i++ {
		go pool.Submit(func() {
			fmt.Println(i)
		})
	}
	time.Sleep(time.Second)
	fmt.Println("ants测试main")
}

func TestCountIf(t *testing.T) {
	var vals = make([]interface{}, 10)
	vals = append(vals, "a")
	vals = append(vals, "ab")
	vals = append(vals, "dc")
	count := util.CountIf(vals, "a*")
	fmt.Println(count)
}

func TestSumIf(t *testing.T) {
	var vals = make([]interface{}, 10)
	vals = append(vals, "10")
	vals = append(vals, "11")
	vals = append(vals, "15.22")
	sum := util.SumIf(vals, "!=10")
	fmt.Println(sum)
}

func TestCompareString(t *testing.T) {
	fmt.Println(util.CompareString("iqwqweqwe cc", "*cc"))
}

func TestFind(t *testing.T) {
	//l := util.Left("这是一个字符串", 4)
	//fmt.Println(l)
	//k := util.Right("这是一个字符串", 3)
	//fmt.Println(k)
	r, _ := util.ComputeFuncResult1("find(abc,cd)", 10, true)
	fmt.Println(r)
	i, _ := util.Find("字符串字符我的", "acw一个字符串字符串字符串字符串字符串")
	fmt.Println(i)
}

func TestGJsonSet(t *testing.T) {
	nJson, _ := sjson.Set("{\"a\":\"123\"}", "b.c", "ttt")
	fmt.Println(nJson)
}

func TestGJsonGet(t *testing.T) {
	var setResult []util.ExecuteSetResult
	b, _ := json.Marshal("[\"key\":\"12\",\"value\":\"www\",\"selected\":true]")
	json.Unmarshal(b, &setResult)

	local, _ := time.LoadLocation("Asia/Shanghai")
	var t1, _ = time.ParseInLocation("2006-01-02", "2023-01-02", local)
	var t2, _ = time.ParseInLocation("2006-01-02 00:00:00", "2023-01-03", local)
	handler.TimeAfter(t1, t2)
	handler.TimeBefore(t1, t2)
}

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

//func TestGJsonGet1(t *testing.T) {
//	db, err := gorm.Open("dm", "dm://SYSDBA:SYSDBA@192.168.100.165:8989")
//	if err != nil {
//		panic("failed to connect database")
//	}
//	defer db.Close()
//
//	// Migrate the schema
//	db.AutoMigrate(&Product{})
//
//	// 创建
//	db.Create(&Product{Code: "L1212", Price: 1000})
//
//	// 读取
//	var product Product
//	db.First(&product, 1)                   // 查询id为1的product
//	db.First(&product, "code = ?", "L1212") // 查询code为l1212的product
//
//	// 更新 - 更新product的price为2000
//	db.Model(&product).Update("Price", 2000)
//
//	// 删除 - 删除product
//	db.Delete(&product)
//}

func TestFloat64Parse(t *testing.T) {
	//f, _ := strconv.ParseFloat("1735232529562095618", 64)
	f, _ := strconv.ParseInt("1735232529562095618", 10, 64)
	println(f)
}

func TestBcrypt(t *testing.T) {
	p, _ := bcrypt.GenerateFromPassword([]byte("rule"), bcrypt.DefaultCost)
	println(string(p))
	err := bcrypt.CompareHashAndPassword(p, []byte("rule"))
	println(err)
}
