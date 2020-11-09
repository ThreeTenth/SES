package main

import "github.com/gin-gonic/gin"

func main() {
	SetupMailCredentials("Enter your email: ", "Enter email's password: ")
	router := gin.Default()
	v1 := router.Group("/v1")
	v1.POST("/send/mail", inout(PostSendMail))
	router.Run(":19825")
}

// PostSendMail send mail
func PostSendMail(c *Context) error {
	type Mail struct {
		Email   string `schema:"email,required"`
		Subject string `schema:"subject,required"`
		Body    string `schema:"body,required"`
	}

	var mail Mail
	var err error

	if err = c.ParseForm(&mail); err != nil {
		return c.BadRequest(err.Error())
	}

	if err = SendMail(mail.Email, mail.Subject, mail.Body); err != nil {
		return c.BadRequest(err.Error())
	}

	return c.NoContent()
}
