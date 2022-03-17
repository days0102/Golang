package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx/types"
)

var dbConnect *pgx.Conn
var imgUrl string = "https://gw.alipayobjects.com/zos/rmsportal/mqaQswcyDLcXyDKnZfES.png"

//const path = "^/(?P<one>\\w+)/(?P<two>\\w+)"

type replyProto struct {
	//Status, 0: success, others: fault
	Status int `json:"status"`

	//Msg, Action result describe by literal
	Msg string `json:"msg,omitempty"`

	//Data, operand
	Data types.JSONText `json:"data,omitempty"`

	// RowCount, just row count
	RowCount int64 `json:"rowCount,omitempty"`

	//API, call target
	API string `json:"API,omitempty"`

	//Method, using http method
	Method string `json:"method,omitempty"`

	//SN, call order
	SN int `json:"SN,omitempty"`

	Articles []Article `json:"articles"`

	Comments []Comment `json:"comments"`

	User User `json:"user"`
}

type Article struct {
	Id          int              `json:"id"`
	CreateBy    string           `json:"create_by"`
	Avatar      string           `json:"avatar"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Figure      string           `json:"figure"`
	Content     string           `json:"content"`
	Stars       int              `json:"stars"`
	Likes       int              `json:"likes"`
	Like        bool             `json:"like"`
	Star        bool             `json:"star"`
	Comments    int              `json:"comments"`
	CreateAt    pgtype.Timestamp `json:"create_at"`
	UpdateAt    pgtype.Timestamp `json:"update_at"`
	CreateTime  string           `json:"create_time"`
	UpdateTime  string           `json:"update_time"`
}

type Comment struct {
	Id         int              `json:"id"`
	CommentTo  int              `json:"comment_to"`
	CreateBy   string           `json:"create_by"`
	Username   string           `json:"username"`
	Avatar     string           `json:"avatar"`
	Content    string           `json:"content"`
	Stars      int              `json:"stars"`
	Likes      int              `json:"likes"`
	Like       bool             `json:"like"`
	Star       bool             `json:"star"`
	Comments   int              `json:"comments"`
	CreateAt   pgtype.Timestamp `json:"create_at"`
	Oneself    int              `json:"oneself"`
	StarsAry   pgtype.TextArray `json:"starsAry"`
	LikesAry   pgtype.TextArray `json:"likesAry"`
	CreateTime string           `json:"create_time"`
}

type User struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Account string `json:"account"`
	Avatar  string `json:"avatar"`
}

func serve(w http.ResponseWriter, r *http.Request) {
	//msg := replyProto{
	//	Status: 0,
	//	Msg:    "success",
	//}

	//fmt.Println("method:" + strings.ToLower(r.Method))
	//
	//fmt.Println(r.URL.Path)
	// fmt.Println(r.RemoteAddr)

	switch r.URL.Path {
	case "/api/login":
		login(r, w)
		break

	case "/api/register":
		register(r, w)
		break

	case "/api/articles":
		ReadArticles(w)
		break

	//case "/api/article":
	//	ReadArticle(r, w)
	//	break

	case "/api/comment":
		ReadComment(r, w)
		break

	case "/api/post":
		PostComment(r, w)
		break

	case "/api/comment/delete":
		DeleteComment(r, w)
		break

	case "/api/article/create":
		CreateArticle(r, w)
		break

	case "/api/upload-img":
		UploadImg(r, w)
		break

	case "/api/upload-video":
		UploadVideo(r, w)
		break

	case "/api/comment/dislike":
		CommentDislike(r, w)
		break

	case "/api/comment/like":
		CommentLike(r, w)
		break

	default:
		//ReadArticles(r, w)
		break
	}

	//regPath := regexp.MustCompile(path)
	//fieldName := regPath.SubexpNames()
	//field := regPath.FindStringSubmatch(r.URL.Path)
	//fmt.Println(fieldName)
	//fmt.Println(field)
	//
	//for i, k := range fieldName {
	//	if i == 0 {
	//		continue
	//	}
	//	fmt.Println(i, k)
	//}

}

type Video struct {
	Url string `json:"url"`
}

//wangEditor指定返回json格式
type VideoRes struct {
	Errno int   `json:"errno"`
	Data  Video `json:"data"`
}

func UploadVideo(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	err := r.ParseMultipartForm(1 << 20 * 20)
	if err != nil {
		return
	}

	fmt.Println(r.MultipartForm.File)

	label := strconv.FormatInt(time.Now().Unix(), 10)
	var name string
	var data Video
	res := VideoRes{
		Errno: 0,
		Data:  data,
	}

	for key, value := range r.MultipartForm.File {
		fmt.Println(key, value)
		file, m, err := r.FormFile(key)
		if err != nil {
			return
		}
		// 保存图片
		//err = os.Mkdir("./images/", 0777)
		//if err != nil {
		//	return
		//}
		name = m.Filename
		saveFile, err := os.OpenFile("../../images/"+label+m.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		_, err = io.Copy(saveFile, file)
		if err != nil {
			msg.Msg = err.Error()
			msg.Status = -1000
			fmt.Println(err.Error())
			return
		}

		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}(file)
		defer func(saveFile *os.File) {
			err := saveFile.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}(saveFile)

		url := "http://8.142.102.189:8083/" + label + name
		fmt.Println(url)
		// imgUrl=url

		data.Url = url
		res.Data = data

	}

	//Marshal()将数据编码成json字符串
	buf, err := json.Marshal(&res)
	if err != nil {
		_, err := w.Write([]byte(fmt.Sprintf(`{"code":-300,"msg":"%s"}`, err.Error())))
		if err != nil {
			return
		}
		fmt.Println(err.Error())
		return
	}

	_, err = w.Write(buf)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func CommentLike(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	// fmt.Println(jsonMap)

	var exec pgconn.CommandTag
	flag, _ := jsonMap["flag"]
	if flag.(float64) == 1 {
		sql := "update comment set likes=array_append((select likes from comment where id=$1),$2) where id=$3"
		dbConnect, err = pgx.Connect(context.Background(), dbString)
		exec, err = dbConnect.Exec(context.Background(), sql, jsonMap["id"], jsonMap["account"], jsonMap["id"])
		if err != nil {
			msg.Status = -600
			msg.Msg = err.Error()
			fmt.Println(err.Error())
			reply(w, &msg)
			err := dbConnect.Close(context.Background())
			if err != nil {
				return
			}
			return
		}
	} else {
		sql := "update comment set likes=array_remove((select likes from comment where id=$1),$2) where id=$3"
		dbConnect, err = pgx.Connect(context.Background(), dbString)
		exec, err = dbConnect.Exec(context.Background(), sql, jsonMap["id"], jsonMap["account"], jsonMap["id"])
		if err != nil {
			msg.Status = -600
			msg.Msg = err.Error()
			fmt.Println(err.Error())
			reply(w, &msg)
			err := dbConnect.Close(context.Background())
			if err != nil {
				return
			}
			return
		}
	}

	if exec.Update() {
		fmt.Println("Update")
		reply(w, &msg)
	}

}

func CommentDislike(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	// fmt.Println(jsonMap)

	var exec pgconn.CommandTag
	flag, _ := jsonMap["flag"]
	if flag.(float64) == 1 {
		sql := "update comment set stars=array_append((select stars from comment where id=$1),$2) where id=$3"
		dbConnect, err = pgx.Connect(context.Background(), dbString)
		exec, err = dbConnect.Exec(context.Background(), sql, jsonMap["id"], jsonMap["account"], jsonMap["id"])
		if err != nil {
			msg.Status = -600
			msg.Msg = err.Error()
			fmt.Println(err.Error())
			reply(w, &msg)
			err := dbConnect.Close(context.Background())
			if err != nil {
				return
			}
			return
		}
	} else {
		sql := "update comment set stars=array_remove((select stars from comment where id=$1),$2) where id=$3"
		dbConnect, err = pgx.Connect(context.Background(), dbString)
		exec, err = dbConnect.Exec(context.Background(), sql, jsonMap["id"], jsonMap["account"], jsonMap["id"])
		if err != nil {
			msg.Status = -600
			msg.Msg = err.Error()
			fmt.Println(err.Error())
			reply(w, &msg)
			err := dbConnect.Close(context.Background())
			if err != nil {
				return
			}
			return
		}
	}

	if exec.Update() {
		reply(w, &msg)
	}

}

//wangEditor指定返回json格式
type Res struct {
	Errno int      `json:"errno"`
	Data  []string `json:"data"`
}

func UploadImg(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	err := r.ParseMultipartForm(1 << 20 * 20)
	if err != nil {
		return
	}

	fmt.Println(r.MultipartForm.File)

	label := strconv.FormatInt(time.Now().Unix(), 10)
	var name string
	data := make([]string, 0, 0)
	res := Res{
		Errno: 0,
		Data:  data,
	}

	for key, value := range r.MultipartForm.File {
		fmt.Println(key, value)
		file, m, err := r.FormFile(key)
		if err != nil {
			return
		}
		// 保存图片
		//err = os.Mkdir("./images/", 0777)
		//if err != nil {
		//	return
		//}
		name = m.Filename
		saveFile, err := os.OpenFile("../../images/"+label+m.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		_, err = io.Copy(saveFile, file)
		if err != nil {
			msg.Msg = err.Error()
			msg.Status = -1000
			fmt.Println(err.Error())
			return
		}

		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}(file)
		defer func(saveFile *os.File) {
			err := saveFile.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}(saveFile)

		url := "http://8.142.102.189:8083/" + label + name
		fmt.Println(url)
		imgUrl = url

		data = append(data, url)
		res.Data = data

	}

	//Marshal()将数据编码成json字符串
	buf, err := json.Marshal(&res)
	if err != nil {
		_, err := w.Write([]byte(fmt.Sprintf(`{"code":-300,"msg":"%s"}`, err.Error())))
		if err != nil {
			return
		}
		fmt.Println(err.Error())
		return
	}

	_, err = w.Write(buf)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func CreateArticle(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	//fmt.Println(jsonMap)

	sql := "insert into article(create_by,title,description,content,figure,stars,likes,comments) values($1,$2,$3,$4,$5,0,0,0)"

	dbConnect, err = pgx.Connect(context.Background(), dbString)
	exec, err := dbConnect.Exec(context.Background(), sql, jsonMap["create_by"], jsonMap["title"], jsonMap["description"], jsonMap["content"], imgUrl)
	fmt.Println("insert:" + imgUrl)
	imgUrl = "https://gw.alipayobjects.com/zos/rmsportal/mqaQswcyDLcXyDKnZfES.png"

	if err != nil {
		msg.Status = -600
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		err := dbConnect.Close(context.Background())
		if err != nil {
			return
		}
		return
	}

	if exec.Insert() {
		reply(w, &msg)
		fmt.Println("发布成功")
	}

	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}

}

//func ReadArticle(r *http.Request, w http.ResponseWriter) {
//	//fmt.Println("ReadArticles")
//	msg := replyProto{
//		Status: 0,
//		Msg:    "success",
//	}
//
//	body, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		msg.Status = -100
//		msg.Msg = err.Error()
//		fmt.Println(err.Error())
//		reply(w, &msg)
//		return
//	}
//
//	jsonMap := make(map[string]interface{})
//	//将json字符串转换成map
//	err = json.Unmarshal(body, &jsonMap)
//	if err != nil {
//		msg.Status = -100
//		reply(w, &msg)
//		return
//	}
//
//	fmt.Println(body)
//
//	var article Article
//
//	sql := "select article.id,article.create_by,article.title,article.description,article.content,article.figure,article.stars,article.likes,article.comments,article.create_at,article.update_at,users.avatar from article,users where article.create_by=users.id and article.id=$1"
//	dbConnect, err = pgx.Connect(context.Background(), dbString)
//	result := dbConnect.QueryRow(context.Background(), sql, jsonMap["id"])
//
//	err = result.Scan(&article.Id, &article.CreateBy, &article.Title, &article.Description, &article.Content, &article.Figure, &article.Stars, &article.Likes, &article.Comments, &article.CreateAt, &article.UpdateAt, &article.Avatar)
//	if err != nil {
//		fmt.Println(err.Error())
//		_ = dbConnect.Close(context.Background())
//		return
//	}
//	msg.Articles = append(msg.Articles, article)
//	reply(w, &msg)
//	err = dbConnect.Close(context.Background())
//	if err != nil {
//		fmt.Println("数据库关闭失败")
//		os.Exit(0)
//	}
//}

func DeleteComment(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	//fmt.Println("ReadComment")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	sql := "delete from comment where id=$1"
	dbConnect, err = pgx.Connect(context.Background(), dbString)
	exec, err := dbConnect.Exec(context.Background(), sql, jsonMap["id"])
	if err != nil {
		msg.Status = -1000
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		_ = dbConnect.Close(context.Background())
		return
	}

	if exec.Delete() {
		reply(w, &msg)
		fmt.Println("删除成功")
	}
	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

func PostComment(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	sql := "insert into comment (comment_to, create_by, content, stars,likes,comments) values($1,$2,$3,'{}','{}',$4)"
	dbConnect, err = pgx.Connect(context.Background(), dbString)
	exec, err := dbConnect.Exec(context.Background(), sql, jsonMap["comment_to"], jsonMap["create_by"], jsonMap["content"], jsonMap["comments"])
	if err != nil {
		msg.Status = -900
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		_ = dbConnect.Close(context.Background())
		return
	}
	if exec.Insert() {
		reply(w, &msg)
		fmt.Println("评论成功")
	}
	//fmt.Println(jsonMap)
	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

func ReadComment(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	//fmt.Println("ReadComment")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	//fmt.Println(body)

	var comment Comment
	sql := "select comment.id,comment_to,create_by,name,avatar,content,stars,likes,comments,comment.create_at from users,comment where users.id=comment.create_by and comment_to=$1 order by create_at desc"

	dbConnect, err = pgx.Connect(context.Background(), dbString)
	result, err := dbConnect.Query(context.Background(), sql, jsonMap["id"])
	if err != nil {
		fmt.Println(err.Error())
		msg.Status = -800
		msg.Msg = err.Error()
		reply(w, &msg)
		_ = dbConnect.Close(context.Background())
		return
	}
	for result.Next() {
		err := result.Scan(&comment.Id, &comment.CommentTo, &comment.CreateBy, &comment.Username, &comment.Avatar, &comment.Content, &comment.StarsAry, &comment.LikesAry, &comment.Comments, &comment.CreateAt)
		if err != nil {
			fmt.Println(err.Error())
			msg.Status = -800
			msg.Msg = err.Error()
			reply(w, &msg)
			_ = dbConnect.Close(context.Background())
			return
		}
		//fmt.Println(comment)
		// fmt.Println(comment.StarsAry)
		comment.Like = false
		comment.Star = false
		msg.Comments = append(msg.Comments, comment)
	}

	var article Article
	sql = "select article.id,article.create_by,article.title,article.description,article.content,article.figure,article.stars,article.likes,article.comments,article.create_at,article.update_at,users.avatar from article,users where article.create_by=users.id and article.id=$1"
	dbConnect, err = pgx.Connect(context.Background(), dbString)
	var articleResult = dbConnect.QueryRow(context.Background(), sql, jsonMap["id"])
	err = articleResult.Scan(&article.Id, &article.CreateBy, &article.Title, &article.Description, &article.Content, &article.Figure, &article.Stars, &article.Likes, &article.Comments, &article.CreateAt, &article.UpdateAt, &article.Avatar)
	if err != nil {
		fmt.Println(err.Error())
		_ = dbConnect.Close(context.Background())
		return
	}
	msg.Articles = append(msg.Articles, article)

	reply(w, &msg)
	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

func ReadArticles(w http.ResponseWriter) {
	//fmt.Println("ReadArticles")
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	var article Article
	sql := "select article.id,create_by,title,description,content,figure,stars,likes,comments,article.create_at,article.update_at,avatar from article,users where article.create_by=users.id order by update_at desc"

	dbConnect, _ = pgx.Connect(context.Background(), dbString)

	result, err := dbConnect.Query(context.Background(), sql)
	if err != nil {
		msg.Status = -600
		msg.Msg = err.Error()
		fmt.Println(err.Error())
		reply(w, &msg)
		err = dbConnect.Close(context.Background())
		if err != nil {
			return
		}
		return
	}

	for result.Next() {
		err := result.Scan(&article.Id, &article.CreateBy, &article.Title, &article.Description, &article.Content, &article.Figure, &article.Stars, &article.Likes, &article.Comments, &article.CreateAt, &article.UpdateAt, &article.Avatar)
		if err != nil {
			fmt.Println(err.Error())
			msg.Msg = err.Error()
			msg.Status = -1000
			reply(w, &msg)
			err := dbConnect.Close(context.Background())
			if err != nil {
				return
			}
			return
		}

		//fmt.Println(article)
		article.Like = false
		article.Star = false
		msg.Articles = append(msg.Articles, article)
	}
	reply(w, &msg)
	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

func register(r *http.Request, w http.ResponseWriter) {
	//fmt.Println("register")
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		msg.Status = -400
		msg.Msg = "null body"
		reply(w, &msg)
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	if jsonMap["user"] == "" {
		msg.Status = -500
		msg.Msg = "user error"
		reply(w, &msg)
		return
	}

	if jsonMap["cert"] == "" {
		msg.Status = -500
		msg.Msg = "cert error"
		reply(w, &msg)
		return
	}

	sql := "insert into login(account,password) values($1,$2)"

	dbConnect, err = pgx.Connect(context.Background(), dbString)
	exec, err := dbConnect.Exec(context.Background(), sql, jsonMap["user"], jsonMap["cert"])
	if err != nil {
		msg.Status = -600
		msg.Msg = err.Error()
		reply(w, &msg)
		err := dbConnect.Close(context.Background())
		if err != nil {
			return
		}
		return
	}

	if exec.Insert() {
		reply(w, &msg)
		fmt.Println("注册成功")
	}

	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

func reply(w http.ResponseWriter, r *replyProto) {
	if w == nil || r == nil {
		fmt.Println("invalid message")
		return
	}

	//Marshal()将数据编码成json字符串
	buf, err := json.Marshal(&r)
	if err != nil {
		_, err := w.Write([]byte(fmt.Sprintf(`{"code":-300,"msg":"%s"}`, err.Error())))
		if err != nil {
			return
		}
		fmt.Println(err.Error())
		return
	}

	_, err = w.Write(buf)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func login(r *http.Request, w http.ResponseWriter) {
	msg := replyProto{
		Status: 0,
		Msg:    "success",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg.Status = -200
		msg.Msg = "null body"
		return
	}

	jsonMap := make(map[string]interface{})
	//将json字符串转换成map
	err = json.Unmarshal(body, &jsonMap)

	//fmt.Println(jsonMap)

	var user User
	account := jsonMap["user"]
	cert := jsonMap["cert"]

	sql := "select avatar,password from users where id=$1 and password=$2"

	dbConnect, err = pgx.Connect(context.Background(), dbString)
	result := dbConnect.QueryRow(context.Background(), sql, account, cert)
	err = result.Scan(&user.Avatar, &cert)
	nonexistent := err == pgx.ErrNoRows

	if nonexistent {
		fmt.Println("login failure")
		msg.Status = -300
		msg.Msg = "invalid account or password"
		reply(w, &msg)
		err := dbConnect.Close(context.Background())
		if err != nil {
			return
		}
		return
	}
	msg.User = user
	//fmt.Println("login success")
	reply(w, &msg)
	err = dbConnect.Close(context.Background())
	if err != nil {
		fmt.Println("数据库关闭失败")
		os.Exit(0)
		return
	}
}

const dbString = "postgres://test:test@localhost:5432/test"

func main() {

	//var err error

	//if err != nil {
	//	fmt.Println("Connect failure")
	//	return
	//}

	//fmt.Println("PostgresSQL Connected")

	fmt.Println("Listening")
	http.HandleFunc("/", serve)
	err := http.ListenAndServe(":6666", nil)
	if err != nil {
		return
	}

	//if dbConnect.Close(context.Background()) != nil {
	//	fmt.Println("Close failure")
	//	return
	//}

}
