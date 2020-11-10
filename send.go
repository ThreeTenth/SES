package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

var auth _PlainAuth
var proxyAddress, smtpAddress string

type _PlainAuth struct {
	identity, username, nickname, password string
	host                                   string
}

// Start 开始
func (a *_PlainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

// Next
func (a *_PlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}

func baseEncode(s string) string {
	return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
}

// func plainAuth(identity, username, password string, host string) smtp.Auth {
// 	return &_PlainAuth{identity, username, password, host}
// }

func init() {
	flag.StringVar(&proxyAddress, "proxy", "", "Send mail (SMTP) server proxy address")
	flag.StringVar(&smtpAddress, "smtp", "smtp.gmail.com:587", "Send mail (SMTP) server address")
}

// SetupMailCredentials 设置邮箱信息
func SetupMailCredentials(enterUsernameTip, enterNicknameTip, enterPasswordTip string) {
	flag.Parse()
	flag.Usage()

	if "" != proxyAddress {
		fmt.Println("Proxy address: ", proxyAddress)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print(enterUsernameTip)
	username, err := reader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	if "" == username {
		panic("Error: Can't empty")
	}

	fmt.Print(enterNicknameTip)
	nickname, _ := reader.ReadString('\n')

	fmt.Print(enterPasswordTip)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		panic(err)
	}

	if 0 == len(bytePassword) {
		panic("Error: Can't empty")
	}

	auth.username = strings.Trim(username, "\r\n")
	auth.nickname = strings.Trim(nickname, "\r\n")
	auth.password = strings.Trim(string(bytePassword), "\r\n")

	fmt.Println(auth.username, auth.nickname, auth.password, proxyAddress, smtpAddress)
}

// SendMail 发送短信
func SendMail(to string, subjcet string, body string) error {
	var c *smtp.Client
	var err error

	if "" != proxyAddress {
		dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
		if err != nil {
			return err
		}
		conn, err := dialer.Dial("tcp", smtpAddress)
		if err != nil {
			return err
		}

		host, _, _ := net.SplitHostPort(smtpAddress)
		c, err = smtp.NewClient(conn, host)
	} else {
		c, err = smtp.Dial(smtpAddress)
	}

	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.Hello("localhost"); err != nil {
		return err
	}

	host, _, err := net.SplitHostPort(smtpAddress)
	if err != nil {
		return err
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		spli := strings.Split(smtpAddress, ":")
		if 2 != len(spli) {
			return errors.New("smtpAddress is an invalid value")
		}
		c.StartTLS(&tls.Config{ServerName: spli[0]})
	}

	// auth := plainAuth("", auth.username, auth.password, host)
	auth.host = host
	if err = c.Auth(&auth); err != nil {
		return err
	}

	// 设置发件人
	if err := c.Mail(auth.username); err != nil {
		return err
	}

	// 设置接收人
	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	headers := map[string]string{}
	headers["Subject"] = baseEncode(subjcet)
	headers["To"] = to
	headers["From"] = auth.nickname + " <" + auth.username + ">"
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Message-ID"] = fmt.Sprintf("<%f.%d@%s>", rand.Float64(), time.Now().UnixNano(), hostname)

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += fmt.Sprintf("\r\n" + body)

	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
