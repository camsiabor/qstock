package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type SimpleHttp struct {
}

var simpleHttpInstance = &SimpleHttp{}

func GetSimpleHttp() {

}

func hotelcrawler() {

	var checkin = "2019-01-25"
	var checkout = "2019-01-26"
	var url = "http://www.booking.com/hotel/mo/galaxy-macau.en-gb.html?aid=397617;label=gog235jc-1FCAEoggI46AdIM1gDaJcBiAEBmAEJuAEXyAEM2AEB6AEB-AEMiAIBqAID;sid=2e5b23a4ac97dfeaa5a1087e5d1e15f3;all_sr_blocks=30141306_89159273_2_2_0;checkin=" + checkin + ";checkout=" + checkout + ";dest_id=-1204094;dest_type=city;dist=0;hapos=1;highlighted_blocks=30141306_89159273_2_2_0;hp_group_set=0;hpos=1;room1=A%2CA;sb_price_type=total;sr_order=popularity;srepoch=1547982423;srpvid=89754e2b105d0188;type=total;ucfs=1&"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Host", "www.booking.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	// req.Header.Set("Cookie", "cors_js=1; _ga=GA1.2.1212520393.1532961488; zz_cook_tms_seg1=1; zz_cook_tms_ed=1; cto_lwid=90e954dd-e522-4417-8504-a8b682fca7ec; zz_cook_tms_ep=1; zz_cook_tms_em=1; zz_cook_tms_eg=1; esadm=02UmFuZG9tSVYkc2RlIyh9YbxZGyl9Y5%2BPCQ%2Be6L1iyuiQmlDq6ydyWKALPUDlLlmpVsAwmz%2FLoOU%3D; he=02UmFuZG9tSVYkc2RlIyh9YbxZGyl9Y5%2BPCQ%2Be6L1iyuiIqKIjP5uh%2F%2BUGiDKk0Q8iPbTZpbGgypc%3D; _gcl_au=1.1.895656110.1545647297; zz_cook_tms_seg3=7; _gid=GA1.2.2019762290.1547982083; header_joinapp_prompt_retargeting=1; vpmss=1; BJS=-; has_preloaded=1; 11_srd=%7B%22features%22%3A%5B%7B%22id%22%3A16%7D%5D%2C%22score%22%3A3%2C%22detected%22%3Afalse%7D; zz_cook_tms_hlist=301413; utag_main=v_id:0164eba007f40062ae11f441708003072004806a00978$_sn:32$_ss:0$_st:1547984362921$4split:3$4split2:1$ses_id:1547982084055%3Bexp-session$_pn:11%3Bexp-session; lastSeen=1547984659586; bkng=11UmFuZG9tSVYkc2RlIyh9YSvtNSM2ADX0BnR0tqAEmjsc8vSgxGCDYesZRvY29jItmxhxeLNutqxXn9%2F%2B4iCE4c2sZ9zdrG0w0o2O6bSqDap%2F0hyVpKPzwJ2o0Ttz5PeL2tRFJ6JpXYcG4AGSTr1HDsQY63sh8LuHO7Yl44l5xkNlb8whIeNaZh7TuxLw7l5booev1kS%2BN3HyFILFosNfoQ%3D%3D");
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var html = string(body)
	// strings.Index()

	log.Println(html)

}

func main2() {

	//resp, _ := doGet("http://www.baidu.com")
	//resp, _ := doPost("http://www.baidu.com", "application/json;charset=utf-8")

	var url = "http://www.booking.com/hotel/mo/galaxy-macau.en-gb.html?aid=397617;label=gog235jc-1FCAEoggI46AdIM1gDaJcBiAEBmAEJuAEXyAEM2AEB6AEB-AEMiAIBqAID;sid=2e5b23a4ac97dfeaa5a1087e5d1e15f3;all_sr_blocks=30141306_89159273_2_2_0;checkin=2019-01-25;checkout=2019-01-26;dest_id=-1204094;dest_type=city;dist=0;hapos=1;highlighted_blocks=30141306_89159273_2_2_0;hp_group_set=0;hpos=1;room1=A%2CA;sb_price_type=total;sr_order=popularity;srepoch=1547982423;srpvid=89754e2b105d0188;type=total;ucfs=1&"

	resp, _ := doGet(url)
	defer resp.Body.Close() //go的特殊语法，main函数执行结束前会执行resp.Body.Close()

	fmt.Println(resp.StatusCode)          //有http的响应码输出
	if resp.StatusCode == http.StatusOK { //如果响应码为200
		body, err := ioutil.ReadAll(resp.Body) //把响应的body读出
		if err != nil {                        //如果有异常
			fmt.Println(err) //把异常打印
			log.Fatal(err)   //日志
		}
		fmt.Println(string(body)) //把响应的文本输出到console
	}

}

/**
以GET的方式请求
**/
func doGet(url string) (r *http.Response, e error) {

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(resp.StatusCode)
		fmt.Println(err)
		log.Fatal(err)
	}
	return resp, err
}

/**
以POST的方式请求
**/
func doPost(url string, bodyType string) (r *http.Response, e error) {

	resp, err := http.Post(url, bodyType, nil)

	if err != nil {
		fmt.Println(resp.StatusCode)
		fmt.Println(err)
		log.Fatal(err)
	}

	return resp, err
}

/**
以Post表单的方式请求
**/
func doPostForm(urlStr string) (r *http.Response, e error) {

	v := url.Values{"method": {"get"}, "id": {"1"}}
	v.Add("name1", "1")
	v.Add("name2", "2")

	resp, err := http.PostForm(urlStr, v)

	if err != nil {
		fmt.Println(resp.StatusCode)
		fmt.Println(err)
		log.Fatal(err)
	}

	return resp, err

}
