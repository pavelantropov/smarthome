package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	deviceProxy    *httputil.ReverseProxy
	telemetryProxy *httputil.ReverseProxy
}

func NewProxyHandler(deviceURL, telemetryURL string) *ProxyHandler {
	return &ProxyHandler{
		deviceProxy:    newServiceProxy(deviceURL, "/api/v1/devices", "/api/devices"),
		telemetryProxy: newServiceProxy(telemetryURL, "/api/v1/telemetry", "/api/telemetry"),
	}
}

func (h *ProxyHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.Any("/devices", h.proxyDevices)
	router.Any("/devices/*path", h.proxyDevices)
	router.Any("/telemetry", h.proxyTelemetry)
	router.Any("/telemetry/*path", h.proxyTelemetry)
}

func (h *ProxyHandler) proxyDevices(c *gin.Context) {
	h.deviceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) proxyTelemetry(c *gin.Context) {
	h.telemetryProxy.ServeHTTP(c.Writer, c.Request)
}

func newServiceProxy(rawURL, publicPrefix, servicePrefix string) *httputil.ReverseProxy {
	target, err := url.Parse(strings.TrimRight(rawURL, "/"))
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = servicePrefix + strings.TrimPrefix(req.URL.Path, publicPrefix)
		req.Host = target.Host
	}
	return proxy
}
