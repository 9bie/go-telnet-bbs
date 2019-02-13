package main

import "fmt"
import "net"

import "net/http"
import "io/ioutil"

import "encoding/json"
import "time"
import "strings"
import "strconv"

type comments_in struct {
	Aid         int    `json:aid`
	Author_id   int    `json:author_id`
	Author_name string `json:author_name`
	Content     string `json:content`
	Rid         int    `json:rid`
	Time        int64  `json:time`
}
type comments struct {
	Code int           `json:code`
	Data []comments_in `json:Data`
}
type Archivers_Data struct {
	Aid         int    `json:aid,int`
	Author_id   int    `json:author_id`
	Author_name string `json:author_name,string`
	Content     string `json:content,omitempty`
	Last_time   int64  `json:last_time`
	Title       string `json:Title`
}
type Archivers_only struct {
	Code int            `json:code`
	Data Archivers_Data `json:data`
}
type Archivers struct {
	Code int              `json:code`
	Data []Archivers_Data `json:data`
}
type MUser struct {
	Name  string `json:name,omitempty`
	Uid   int    `json:uid,omitempty`
	Uname string `json:uname,omitempty`
	Aid   int    `json:aid,omitempty`
}
type LUser struct {
	Code int    `json:code`
	Msg  string `json:msg,omitempty`
	Data MUser  `json:data,omitempty`
}

func httpPost(url string, data map[string]string) string {
	param := ""
	if len(data) != 0 {
		for v, i := range data {
			param += v + "=" + i + "&"
		}
	} else {
		param = ""
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(param))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	return string(body)
}

func httpGet(url string, data map[string]string) string {
	if len(data) != 0 {
		url += "?"
		for v, i := range data {
			url += v + "=" + i + "&"
		}
	}

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err.Error())
		return ""

	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return ""

	}

	return string(body)
}

// func SortArchivers(d []Archivers_Data, methods int, New_ *[]Archivers_Data) {
// 	//method:
// 	//1 时间，正序
// 	//2 时间，倒叙
// 	var sort []int
// 	for i, _ := range d {
// 		append(sort, i)
// 	}
// 	for j, _ := range sort {
// 		for k := 0; k <= len(sort)-j; j++ {

// 		}
// 	}

// }
func ReadReply(aid string, page string, c net.Conn) {
	data := make(map[string]string)
	data["aid"] = aid
	data["page"] = page
	ret := httpGet("http://127.0.0.1:8080/replies", data)
	fmt.Println(ret)
	if ret == "" {
		return
	} else {
		var p comments
		err := json.Unmarshal([]byte(ret), &p)
		if err != nil {
			return
		}
		for _, j := range p.Data {
			c.Write([]byte("\033[36m" + j.Author_name + ":\033[35m" + j.Content + "\r\n"))
		}
	}
}

func ReadArichvers(aid string, c net.Conn, u LUser, islogin bool) {
	data := make(map[string]string)
	data["aid"] = aid
	ret := httpGet("http://127.0.0.1:8080/detail", data)
	if ret == "" {
		return
	} else {
		var p Archivers_only
		err := json.Unmarshal([]byte(ret), &p)
		if err != nil {
			return
		}
		c.Write([]byte("\033[34;47m标题：\033[44;37m" + p.Data.Title + "\r\n\033[34;47m作者：\033[44;37m" + p.Data.Author_name + "\033[0m\r\n\033[1;4m" + p.Data.Content + "\033[0m\r\n"))
		c.Write([]byte("\033[32m※※※※※※※※※\r\n输入a[d] 上【下】页。输入q退出帖子。输入reply回帖\r\n当前评论数：8/36 \r\n※※※※※※※※※\r\n\033[0m"))
		ReadReply(aid, "1", c)
		c.Write([]byte("\033[0mview>" + aid + ">"))
		for {
			tmp := make([]byte, 128)
			n, _ := c.Read(tmp)
			s := string(tmp[:n])
			s = CharDele(s)
			s = strings.ToLower(s)
			s = strings.Replace(s, string(255)+string(241), "", -1)
			if s == "\r\n" || len(tmp[:n]) == 2 && tmp[0] == byte(255) && tmp[1] == byte(241) {
				continue
			} else if s == "q" || s == "Q" {
				break
			} else if s == "reply" || s == "REPLY" {
				if islogin == true {
					c.Write([]byte("you say>"))
					for {
						tmp := make([]byte, 128)
						n, _ = c.Read(tmp)
						if string(tmp[:n]) == "\r\n" || len(tmp[:n]) == 2 && tmp[0] == byte(255) && tmp[1] == byte(241) {
							continue
						}
						rep := string(tmp[:n])

						data := make(map[string]string)
						data["aid"] = aid
						data["content"] = rep
						fmt.Println(u.Data.Uid, u.Data.Aid)
						data["author_id"] = strconv.Itoa(u.Data.Uid)
						ret := httpPost("http://127.0.0.1:8080/reply", data)
						ret = strings.Replace(ret, string(255)+string(241), "", -1)
						fmt.Println(ret)
						if strings.Index(ret, `code": 0`) == -1 {
							c.Write([]byte("\033[31m × 回帖失败\033[0m\r\nview>" + aid + ">"))
						} else {
							c.Write([]byte("\033[32m √ 回帖成功\033[0m\r\n"))
							ReadReply(aid, "1", c)
							c.Write([]byte("\033[0mview>" + aid + ">"))
						}
						break
					}
				} else {
					c.Write([]byte("\033[31m × 请先登陆！\033[0m\r\nview>"))
				}
			}
		}
		c.Write([]byte(">"))
	}
}
func GetPage(c net.Conn) map[string]int {
	ArchiversList := make(map[string]int)
	data := make(map[string]string)
	var archivers Archivers
	ret := httpGet("http://127.0.0.1:8080/articles", data)
	err := json.Unmarshal([]byte(ret), &archivers)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ret)
	if archivers.Code != 0 {
		c.Write([]byte("\033[31m ×  获取文章列表失败\033[0m\r\n>"))
		return ArchiversList
	} else {
		if len(archivers.Data) == 0 {
			c.Write([]byte("    抱歉，暂无文章哟，赶紧来发新文章吧！\r\n>"))
		} else {

			c.Write([]byte("文章列表：\r\n"))
			i := 0
			for _, j := range archivers.Data {
				i += 1

				c.Write([]byte("   \033[33m" + strconv.Itoa(i) + "\033[0m.标题：\033[35m" + j.Title + "\033[0m\r\n     作者：\033[32m" + j.Author_name + "\033[0m\r\n     最后更新:" + time.Unix(j.Last_time, 0).String() + "\r\n"))
				ArchiversList[strconv.Itoa(i)] = j.Aid
			}
			return ArchiversList
		}
		return ArchiversList
	}
}
func CharDele(s string) string {
	if len(s) <= 2 {
		return s
	}
	if s[len(s)-2:len(s)] == "\r\n" {
		return s[:len(s)-2]
	}
	return s
}
func ParmarHandle(s string, c net.Conn, u *LUser, islogin *bool) {

	s = CharDele(s)
	// fmt.Println("Char:", s)
	// fmt.Println("charb:", []byte(s))
	s = strings.ToLower(s)
	s = strings.Replace(s, string(255)+string(241), "", -1)
	if s == "HELP" || s == "help" {
		c.Write([]byte("\033[42;33m####欢迎使用帮助文件#### \033[0m\r\n\r\n"))
		c.Write([]byte("\033[33mLOGin 登陆\r\nregister 注册用户\r\nchat 讨论模式(尽情期待)\r\nLOGOUT 退出登陆\r\nVIEW 浏览模式\r\nWRITE 发新帖子\r\nexit 退出论坛程序\r\n\033[36mcls 清屏\r\nmy 用户信息\r\nabout 关于\033[0m\r\n\r\n>"))
	} else if s == "exit" || s == "EXIT" {
		c.Close()
	} else if s == "LOGIN" || s == "login" {
		c.Write([]byte("\033[35m请输入你的用户名:\033[0m\r\nlogin>username:"))
		User := ""
		Pwd := ""

		tmp := make([]byte, 128)
		for {
			n, _ := c.Read(tmp)
			if string(tmp[:n]) == "\r\n" || len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			User = string(tmp[:n])
			User = CharDele(User)
			break
		}
		c.Write([]byte("\033[35m请输入你的密码:\033[0m\r\nlogin>password:\033[37;47m"))

		for {
			n, _ := c.Read(tmp)
			if string(tmp[:n]) == "\r\n" || len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			Pwd = string(tmp[:n])
			Pwd = CharDele(Pwd)
			break
		}
		c.Write([]byte("\033[0m"))
		data := make(map[string]string)
		data["uname"] = User
		data["pwd"] = Pwd
		ret := httpGet("http://127.0.0.1:8080/login", data)
		if ret == "" {
			fmt.Println("error")
			return
		}
		fmt.Println(ret)
		err := json.Unmarshal([]byte(ret), u)
		if err != nil {
			fmt.Println(err.Error())
		}
		if (*u).Code == 0 {
			c.Write([]byte("\033[32m √ 登陆成功\033[0m\r\n"))
			*islogin = true
			(*u).Data.Uname = User
			time.Sleep(30000)
			//c.Write([]byte("\033[2J"))
			c.Write([]byte([]byte("欢迎回来：" + (*u).Data.Name + "\r\n当前时间是：" + time.Now().String() + "\r\n>")))
			//c.Write([]byte(">"))
		} else {
			c.Write([]byte("\033[31m × 登陆失败\r\n    错误信息：" + (*u).Msg + "+\033[0m\r\n"))
		}
	} else if s == "LOGOUT" || s == "logout" {
		if *islogin == true {
			*islogin = false
			c.Write([]byte("\033[32m √ 登出成功\033[0m\r\n>"))
		} else {
			c.Write([]byte("\033[31m × 您未登陆\033[0m\r\n>"))
		}

	} else if s == "View" || s == "view" {
		List := GetPage(c)
		if len(List) == 0 {
			c.Write([]byte("\033[31m × 获取文章列表失败\033[0m\r\n>"))
			return
		}
		c.Write([]byte("\033[32m※※※※※※※※※※※※※※※※※※\r\n当前页面 1/36\r\n请输入文章ID阅读文章,/q退出\033[0m\r\n"))
		c.Write([]byte("view>"))
		aid := ""
		for {
			tmp := make([]byte, 128)
			n, _ := c.Read(tmp)
			s2 := string(tmp[:n])
			s2 = strings.Replace(s2, string(255)+string(241), "", -1)
			if string(tmp[:n]) == "\r\n" {
				//c.Write([]byte("view>"))
				continue
			} else if len(s2) == 2 && ([]byte(s2)[0] == byte(255) && ([]byte(s2)[1] == byte(241))) {
				continue
			} else if len(s2) == 2 {
				if s2[:2] == "/q" {
					return
				}
			}

			aid = string(tmp[:n])
			fmt.Println(aid)
			aid = CharDele(aid)
			break
		}
		aid = strconv.Itoa(List[aid])
		c.Write([]byte("\033[2J"))
		ReadArichvers(aid, c, *u, *islogin)
	} else if s == "write" || s == "WRITE" {
		if *islogin == false {
			c.Write([]byte("\033[31m × 请先登陆！\033[0m\r\n>"))
			return
		}
		c.Write([]byte("\033[35m标题>\033[0m"))
		title := ""
		for {
			tmp := make([]byte, 128)
			n, _ := c.Read(tmp)
			s = string(tmp[:n])
			s = strings.Replace(s, string(255)+string(241), "", -1)
			fmt.Println(s)
			if string(tmp[:n]) == "\r\n" {
				//c.Write([]byte("\033[35m标题>\033[0m"))
				continue
			} else if len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			title = CharDele(s)
			break
		}
		content := ""
		c.Write([]byte("\033[32m※※※※※※※※※※※※※※※※※※\r\n按行输入，请谨慎输入，/end结束，/del删除一行,/q退出编辑\r\n※※※※※※※※※※※※※※※※※※※※※※※※※※※\033[0m\r\n"))
		for {
			tmp := make([]byte, 512)
			n, _ := c.Read(tmp)
			s = string(tmp[:n])
			s = strings.Replace(s, string(255)+string(241), "", -1)
			if len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			} else if s[:4] == "/end" {
				break
			}

			content += s
		}
		data := make(map[string]string)
		data["title"] = title
		data["content"] = content
		data["author_id"] = strconv.Itoa(u.Data.Uid)

		ret := httpPost("http://127.0.0.1:8080/post", data)
		if strings.Index(ret, ``) == -1 {
			c.Write([]byte("\033[31m × 发送文章失败\033[0m\r\n>"))
		} else {
			c.Write([]byte("\033[32m √ 发送文章成功\033[0m\r\n>"))
		}

	} else if s == "register" || s == "REGISTER" {
		c.Write([]byte("请输入你的用户名:\r\nregister>"))
		tmp := make([]byte, 512)
		User := ""
		Name := ""
		Pwd := ""
		for {
			n, _ := c.Read(tmp)
			if string(tmp[:n]) == "\r\n" {
				//c.Write([]byte("register>"))
				continue
			} else if len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			User = string(tmp[:n])
			User = CharDele(User)
			break
		}
		c.Write([]byte("请输入你的昵称:\r\nregister>"))
		for {
			n, _ := c.Read(tmp)
			if string(tmp[:n]) == "\r\n" {
				//c.Write([]byte("register>"))
				continue
			} else if len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			Name = string(tmp[:n])
			Name = CharDele(Name)
			break
		}
		c.Write([]byte("请输入你的密码:\r\nregister>"))

		for {
			n, _ := c.Read(tmp)
			if string(tmp[:n]) == "\r\n" {
				//c.Write([]byte("register>"))
				continue
			} else if len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
				continue
			}
			Pwd = string(tmp[:n])
			Pwd = CharDele(Pwd)
			break
		}
		fmt.Println(User, Name, Pwd)
		data := make(map[string]string)
		data["uname"] = User
		data["name"] = Name
		data["pwd"] = Pwd
		ret := httpGet("http://127.0.0.1:8080/register", data)
		if ret == "" {
			fmt.Println("error")
			return
		}

		fmt.Println(ret)
		var M LUser
		err := json.Unmarshal([]byte(ret), &M)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(M.Code, M.Msg)
		if M.Code != 0 {
			c.Write([]byte("\033[31m × 注册失败\r\n    错误信息：" + M.Msg + "+\033[0m\r\n"))
		} else {
			c.Write([]byte("\033[32m √ 注册成功\033[0m\r\n"))
		}
		c.Write([]byte(">"))

		return

	} else if s == "cls" || s == "CLS" {
		c.Write([]byte("\033[2J"))
		c.Write([]byte("\033[41;36m欢迎来到我也不知道叫什么好的论坛，请输入/HELP以显示帮助文档。\033[0m\r\n"))
		if *islogin == true {
			c.Write([]byte("欢迎回来：" + (*u).Data.Name + "    当前时间是：" + time.Now().String() + "\r\n"))
			c.Write([]byte(">"))
		}
	} else if s == "\r\n" {
		c.Write([]byte(">"))

	} else if s == "" || len([]byte(s)) == 2 && ([]byte(s)[0] == byte(255) && ([]byte(s)[1] == byte(241))) {
		return
	} else if s == "my" || s == "MY" {
		if *islogin == false {
			c.Write([]byte("\033[31m × 请先登陆\033[0m\r\n"))
		} else {
			c.Write([]byte("\033[32m 您的信息为:\r\n+昵称：" + (*u).Data.Name + "\r\n+用户名：" + (*u).Data.Uname + "\r\n+用户ID：" + strconv.Itoa((*u).Data.Uid) + " \033[0m\r\n>"))
		}

	} else if s == "about" {
		c.Write([]byte("\r\n\033[34;47m请加入我们的怠惰势力，一起玩耍！\r\n群：\033[44;37m558226805\033[0m\r\n\r\n>"))
	} else {
		c.Write([]byte("\033[31m × 无此命令！\033[0m\r\n\r\n>"))
	}

}
func HandleConnectcion(c net.Conn) {
	var User LUser
	islogin := false
	c.Write([]byte("\033[41;36m欢迎来到我也不知道叫什么好的论坛，请输入HELP以显示帮助文档。\033[0m\r\n>"))
	ret := make([]byte, 128)
	defer c.Close()
	for {

		n, err := c.Read(ret)
		if err != nil {
			break
		}
		fmt.Println(string(ret[:n]) + "\n")
		fmt.Println(ret[:n])
		ParmarHandle(string(ret[:n]), c, &User, &islogin)
		//获取用户名
	}

}
func main() {
	l, err := net.Listen("tcp", ":23")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go HandleConnectcion(conn)
	}

}
