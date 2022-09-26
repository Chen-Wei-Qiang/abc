package main

import (
	"fmt"
	"gitea.com/lunny/log"
	"github.com/axgle/mahonia"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"net"
	"time"
)

// 登录函数
func loginEmail(Eserver, UserName, Password string) (*client.Client, error) {
	dial := new(net.Dialer)
	dial.Timeout = time.Duration(3) * time.Second
	c, err := client.DialWithDialerTLS(dial, Eserver, nil)
	if err != nil {
		c, err = client.DialWithDialer(dial, Eserver) // 非加密登录
	}
	if err != nil {
		return nil, err
	}
	//登陆
	if err = c.Login(UserName, Password); err != nil {
		return nil, err
	}
	return c, nil
}

func parseEmail(mr *mail.Reader) (body []byte, fileMap map[string][]byte) {
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}
		if p != nil {
			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				body, err = ioutil.ReadAll(p.Body)
				if err != nil {
					fmt.Println("read body err:", err.Error())
				}

			case *mail.AttachmentHeader:
				fileName, _ := h.Filename()
				fileContent, _ := ioutil.ReadAll(p.Body)
				fileMap[fileName] = fileContent
			}
		}
	}
	return
}

func main() {
	emailListByUid("imap.qq.com:993", "1790040642@qq.com", "nnmietnbqrvpcjib")
}

// ConvertToString 将字符串转为utf-8编码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// IsGBK 判断byte是否为gbk 编码
func IsGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0xff { //编码小于等于127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else { //大于127的使用双字节编码
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func ParseBody(body io.Reader) (eBody []byte, err error) {
	bodyByte, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println(err)
	}
	if bodyByte != nil {
		emailBody := string(bodyByte)
		if IsGBK(bodyByte) {
			emailBody = ConvertToString(emailBody, "gbk", "utf-8")
		}
		eBody = []byte(emailBody)
	}
	return
}

func emailListByUid(Eserver, UserName, Password string) (err error) {
	c, err := loginEmail(Eserver, UserName, Password)
	if err != nil {
		log.Info("login:", err)
		return
	}
	idClient := id.NewClient(c)
	idClient.ID(
		id.ID{
			id.FieldName:    "IMAPClient",
			id.FieldVersion: "2.1.0",
		},
	)

	defer c.Close()
	mailboxes := make(chan *imap.MailboxInfo, 10)
	mailboxeDone := make(chan error, 1)
	go func() {
		mailboxeDone <- c.List("", "*", mailboxes)
	}()
	for box := range mailboxes {
		//fmt.Println("切换目录:", box.Name)
		mbox, err := c.Select(box.Name, false)
		// 选择收件箱
		if err != nil {
			fmt.Println("select inbox err: ", err)
			continue
		}
		if mbox.Messages == 0 {
			continue
		}

		// 选择收取邮件的时间段
		criteria := imap.NewSearchCriteria()
		// 收取7天之内的邮件
		criteria.Since = time.Now().Add(-7 * time.Hour * 24)
		// 按条件查询邮件
		ids, err := c.UidSearch(criteria)
		// fmt.Println(len(ids))
		if err != nil {
			continue
		}
		if len(ids) == 0 {
			continue
		}
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)
		sect := &imap.BodySectionName{Peek: true}

		messages := make(chan *imap.Message, 100)
		messageDone := make(chan error, 1)

		go func() {
			messageDone <- c.UidFetch(seqset, []imap.FetchItem{sect.FetchItem()}, messages)
		}()
		for msg := range messages {
			r := msg.GetBody(sect)
			//b, _ := ioutil.ReadAll(r)
			//fmt.Println("Got text: %v\n", string(b))
			//mr, _ := mail.CreateReader(r)
			body1, _ := ParseBody(r)
			fmt.Println("-----------------------------")
			fmt.Println(string(body1))
			fmt.Println("-----------------------------")
			//if err != nil {
			//	fmt.Println(err)
			//	continue
			//}
			//header, _ := mr.NextPart()

			//fmt.Println(string(header.Body))

			//_, fileName := parseEmail(mr)
			//for k, _ := range fileName {
			//	//fmt.Println("xxxx:", k)
			//}
		}
	}
	return
}
