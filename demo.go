package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"reflect"
)

var client *elastic.Client
var host = "http://127.0.0.1:9200"

type Employee struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	About     string   `json:"about"`
	Interests []string `json:"interests"`
}

//初始化
func init() {
	errorLog := log.New(os.Stdout, "APP", log.LstdFlags)
	errorLog = errorLog
	var err error
	client, err = elastic.NewClient(elastic.SetURL(host))
	if err != nil {
		panic(err)
	}
	info, code, err := client.Ping(host).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch 返回代码 %d 和版本 %s\n", code, info.Version.Number)

	esVersion, err := client.ElasticsearchVersion(host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch 版本 %s\n", esVersion)

}

/*下面是简单的CURD*/

//创建
func create() {

	//使用结构体
	e1 := Employee{"彭", "国朝", 24, "我喜欢打乒乓球", []string{"运动"}}
	put1, err := client.Index().
		Index("megacorp").
		Id("1").
		BodyJson(e1).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put1.Id, put1.Index, put1.Type)

	//使用字符串
	e2 := `{"first_name":"彭","last_name":"国朝","age":25,"about":"我喜欢骑行","interests":["运动"]}`
	put2, err := client.Index().
		Index("megacorp").
		Id("2").
		BodyJson(e2).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put2.Id, put2.Index, put2.Type)

	e3 := `{"first_name":"Douglas","last_name":"Fir","age":35,"about":"I like to build cabinets","interests":["forestry"]}`
	put3, err := client.Index().
		Index("megacorp").
		Id("3").
		BodyJson(e3).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put3.Id, put3.Index, put3.Type)

}

//删除
func esDelete() {

	res, err := client.Delete().Index("megacorp").
		Id("1").
		Do(context.Background())
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("%#v\n\n",res)
	fmt.Printf("delete result %s\n", res.Result)
}

//修改
func update() {
	res, err := client.Update().
		Index("megacorp").
		Id("2").
		Doc(map[string]interface{}{"age": 18}).
		Do(context.Background())
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("%#v\n\n",res)
	if res != nil {
		fmt.Printf("update age %s\n", res.Result)
	}

}

//查找
func gets() {
	//通过id查找
	get1, err := client.Get().Index("megacorp").Id("2").Do(context.Background())
	if err != nil {
		panic(err)
	}
	if get1.Found {
		fmt.Println(string(get1.Source))
		fmt.Printf("Got document %s in version %d from index %s, type %s\n", get1.Id, get1.Version, get1.Index, get1.Type)
	}
}

//搜索
func query() {
	var res *elastic.SearchResult
	var err error
	//取所有
	res, err = client.Search("megacorp").Do(context.Background())
	printEmployee(res, err)

	//字段相等
	q := elastic.NewQueryStringQuery("last_name:国朝")
	res, err = client.Search("megacorp").Query(q).Do(context.Background())
	if err != nil {
		println(err.Error())
	}
	printEmployee(res, err)

	//条件查询
	//年龄大于30岁的
	boolQ := elastic.NewBoolQuery()
	boolQ.Must(elastic.NewMatchQuery("last_name", "Fir"))
	boolQ.Filter(elastic.NewRangeQuery("age").Gt(30))
	res, err = client.Search("megacorp").Query(boolQ).Do(context.Background())
	printEmployee(res, err)

	//短语搜索 搜索about字段中有 rock climbing
	matchPhraseQuery := elastic.NewMatchPhraseQuery("about", "build")
	res, err = client.Search("megacorp").Query(matchPhraseQuery).Do(context.Background())
	printEmployee(res, err)

	//分析 interests
	aggregations := elastic.NewTermsAggregation().Field("age")
	res, err = client.Search("megacorp").Aggregation("age", aggregations).Do(context.Background())
	printEmployee(res, err)

}

//简单分页
func list(size, page int) {
	if size < 0 || page < 1 {
		fmt.Printf("param error")
		return
	}
	res, err := client.Search("megacorp").
		Size(size).
		From((page - 1) * size).
		Do(context.Background())
	printEmployee(res, err)

}

//打印查询到的Employee
func printEmployee(res *elastic.SearchResult, err error) {
	if err != nil {
		print(err.Error())
		return
	}
	var typ Employee
	for _, item := range res.Each(reflect.TypeOf(typ)) { //从搜索结果中取数据的方法
		t := item.(Employee)
		fmt.Printf("%#v\n", t)
	}
}

func main() {
	create()
	esDelete()
	update()
	gets()
	query()
	list(1, 1)
}