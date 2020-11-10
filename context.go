// Adopt the RESTful API design principle,
// and the returned status code can be found at:
// https://developer.amazon.com/zh/docs/amazon-drive/ad-restful-api-response-codes.html

package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/gorilla/schema"
)

// Context http response content
type Context struct {
	*gin.Context
}

var decoder = schema.NewDecoder()

func inout(handle func(p *Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := &Context{c}

		proxy = proxy.Any()

		handle(proxy)
	}
}

// Render writes the response headers and calls render.Render to render data.
func (p *Context) Render(code int, r render.Render) error {
	p.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(p.Writer)
		p.Writer.WriteHeaderNow()
		return nil
	}

	return r.Render(p.Writer)
}

// Any Set to be accessible by anyone
func (p *Context) Any() *Context {
	p.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	p.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	p.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	p.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	return p
}

// FormValue returns the form value of the specified key,
// return error if the value does not exist or is empty
func (p *Context) FormValue(key string) (string, error) {
	if values, ok := p.GetPostFormArray(key); ok || "" == values[0] {
		return values[0], nil
	}
	return "", errors.New(key + " is empty")
}

// ParseForm writes form content to dst,
// return error if the content does not exist or is empty
func (p *Context) ParseForm(dst interface{}) error {
	err := p.Request.ParseForm()
	if err != nil {
		return err
	}

	err = decoder.Decode(dst, p.Request.PostForm)
	if err != nil {
		return err
	}
	return nil
}

// BadRequest writes a BadRequest code(400) with the given string into the response body.
// Bad input parameter. Error message should indicate which one and why.
func (p *Context) BadRequest(format string, values ...interface{}) error {
	return p.Render(http.StatusBadRequest, render.String{Format: format, Data: values})
}

// Unauthorized writes a Unauthorized request code(401) request with the given string into the response body.
// The client passed in the invalid Auth token. Client should refresh the token and then try again.
func (p *Context) Unauthorized(format string, values ...interface{}) error {
	return p.Render(http.StatusUnauthorized, render.String{Format: format, Data: values})
}

// Forbidden writes a Forbidden request code(403) with the given string into the response body.
// * Customer doesn’t exist.
// * Application not registered.
// * Application try to access to properties not belong to an App.
// * Application try to trash/purge root node.
// * Application try to update contentProperties.
// * Operation is blocked (for third-party apps).
// * Customer account over quota.
func (p *Context) Forbidden(format string, values ...interface{}) error {
	return p.Render(http.StatusForbidden, render.String{Format: format, Data: values})
}

// NotFound writes a NotFound request code(404) with the given string into the response body.
// Resource not found.
func (p *Context) NotFound(format string, values ...interface{}) error {
	return p.Render(http.StatusNotFound, render.String{Format: format, Data: values})
}

// MethodNotAllowed writes a MethodNotAllowed request code(405) with the given string into the response body.
// The resource doesn't support the specified HTTP verb.
func (p *Context) MethodNotAllowed(format string, values ...interface{}) error {
	return p.Render(http.StatusMethodNotAllowed, render.String{Format: format, Data: values})
}

// Conflict writes a Conflict request code(409) with the given string into the response body.
// Conflict
func (p *Context) Conflict(format string, values ...interface{}) error {
	return p.Render(http.StatusConflict, render.String{Format: format, Data: values})
}

// LengthRequired writes a LengthRequired request code(411) with the given string into the response body.
// The Content-Length header was not specified.
func (p *Context) LengthRequired(format string, values ...interface{}) error {
	return p.Render(http.StatusLengthRequired, render.String{Format: format, Data: values})
}

// PreconditionFailed writes a PreconditionFailed request code(412) with the given string into the response body.
// Precondition failed.
func (p *Context) PreconditionFailed(format string, values ...interface{}) error {
	return p.Render(http.StatusPreconditionFailed, render.String{Format: format, Data: values})
}

// TooManyRequests writes a TooManyRequests request code(429) with the given string into the response body.
// Too many request for rate limiting.
func (p *Context) TooManyRequests(format string, values ...interface{}) error {
	return p.Render(http.StatusTooManyRequests, render.String{Format: format, Data: values})
}

// InternalServerError writes a InternalServerError request code(500) with the given string into the response body.
// Servers are not working as expected. The request is probably valid but needs to be requested again later.
func (p *Context) InternalServerError(format string, values ...interface{}) error {
	return p.Render(http.StatusInternalServerError, render.String{Format: format, Data: values})
}

// ServiceUnavailable writes a ServiceUnavailable request code(503) with the given string into the response body.
// Service Unavailable.
func (p *Context) ServiceUnavailable(format string, values ...interface{}) error {
	return p.Render(http.StatusServiceUnavailable, render.String{Format: format, Data: values})
}

// Ok serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (p *Context) Ok(obj interface{}) error {
	return p.Render(http.StatusOK, render.JSON{Data: obj})
}

// Created writes a Created request code(201) with the given string into the response body.
// Indicates that request has succeeded and a new resource has been created as a result.
func (p *Context) Created(format string, values ...interface{}) error {
	return p.Render(http.StatusCreated, render.String{Format: format, Data: values})
}

// NoContent writes a NoContent request code(204) with the given string into the response body.
// The server has fulfilled the request but does not need to return a response body.
// The server may return the updated meta information.
func (p *Context) NoContent() error {
	p.Status(http.StatusNoContent)
	render.String{}.WriteContentType(p.Writer)
	p.Writer.WriteHeaderNow()
	return nil
}

// Path writes the specified file into the body stream in a efficient way.
func (p *Context) Path(filepath string) error {
	http.ServeFile(p.Writer, p.Request, filepath)
	return nil
}

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}

	return true
}
