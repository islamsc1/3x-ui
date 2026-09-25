package sub

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func tlsAllowInsecureStream(value string) string {
	settings := `{"fingerprint":"chrome"}`
	if value != "" {
		settings = `{"fingerprint":"chrome","allowInsecure":` + value + `}`
	}
	return `{"network":"tcp","security":"tls","tcpSettings":{"header":{"type":"none"}},"tlsSettings":{"serverName":"sni.example.com","settings":` + settings + `}}`
}

func allowInsecureInbound(protocol model.Protocol, settings, stream string) *model.Inbound {
	return &model.Inbound{
		Listen:         "203.0.113.1",
		Port:           443,
		Protocol:       protocol,
		Remark:         "allowinsecure",
		Settings:       settings,
		StreamSettings: stream,
	}
}

func TestSub_TLSAllowInsecure_RawLinks(t *testing.T) {
	vlessSettings := `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user"}],"decryption":"none","encryption":"none"}`
	vmessSettings := `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user","security":"auto"}]}`
	trojanSettings := `{"clients":[{"password":"trojanpw","email":"user"}]}`
	ssSettings := `{"method":"chacha20-ietf-poly1305","password":"inboundpw","clients":[{"password":"clientpw","email":"user"}]}`

	cases := []struct {
		name     string
		protocol model.Protocol
		settings string
		value    string
		wantSet  bool
	}{
		{"vless allowInsecure=true", model.VLESS, vlessSettings, "true", true},
		{"vless allowInsecure=false", model.VLESS, vlessSettings, "false", false},
		{"vless allowInsecure absent", model.VLESS, vlessSettings, "", false},
		{"vmess allowInsecure=true", model.VMESS, vmessSettings, "true", true},
		{"vmess allowInsecure=false", model.VMESS, vmessSettings, "false", false},
		{"vmess allowInsecure absent", model.VMESS, vmessSettings, "", false},
		{"trojan allowInsecure=true", model.Trojan, trojanSettings, "true", true},
		{"trojan allowInsecure=false", model.Trojan, trojanSettings, "false", false},
		{"trojan allowInsecure absent", model.Trojan, trojanSettings, "", false},
		{"shadowsocks allowInsecure=true", model.Shadowsocks, ssSettings, "true", true},
		{"shadowsocks allowInsecure=false", model.Shadowsocks, ssSettings, "false", false},
	}

	s := &SubService{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inbound := allowInsecureInbound(tc.protocol, tc.settings, tlsAllowInsecureStream(tc.value))
			var link string
			switch tc.protocol {
			case model.VLESS:
				link = s.genVlessLink(inbound, "user")
			case model.VMESS:
				link = s.genVmessLink(inbound, "user")
			case model.Trojan:
				link = s.genTrojanLink(inbound, "user")
			case model.Shadowsocks:
				link = s.genShadowsocksLink(inbound, "user")
			}
			if link == "" {
				t.Fatal("generator returned an empty link")
			}
			if tc.wantSet {
				if tc.protocol == model.VMESS {
					if !vmessObjHasAllowInsecure(t, link) {
						t.Fatalf("vmess obj should carry allowInsecure=true\n got: %q", link)
					}
					return
				}
				u, err := url.Parse(link)
				if err != nil {
					t.Fatalf("parse link %q: %v", link, err)
				}
				if got := u.Query().Get("allowInsecure"); got != "1" {
					t.Fatalf("link should carry allowInsecure=1\n got: %q", link)
				}
			} else if tc.protocol == model.VMESS {
				if vmessObjHasAllowInsecure(t, link) {
					t.Fatalf("vmess obj must not carry allowInsecure when the flag is %q: %s", tc.value, link)
				}
			} else if strings.Contains(link, "allowInsecure") {
				t.Fatalf("link must not carry allowInsecure when the flag is %q: %s", tc.value, link)
			}
		})
	}
}

func vmessObjHasAllowInsecure(t *testing.T, link string) bool {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(link, "vmess://"))
	if err != nil {
		t.Fatalf("vmess link is not base64: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("vmess payload is not json: %v", err)
	}
	v, ok := obj["allowInsecure"]
	return ok && v == true
}

func TestSubJsonService_TLSAllowInsecure(t *testing.T) {
	svc := NewSubJsonService("", "", "", "", nil)

	cases := []struct {
		name    string
		value   string
		wantSet bool
	}{
		{"allowInsecure=true is copied", "true", true},
		{"allowInsecure=false is dropped", "false", false},
		{"allowInsecure absent stays absent", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stream := svc.streamData(
				`{"network":"tcp","security":"tls","tlsSettings":{"serverName":"a.example.com","settings":`+tlsAllowInsecureSettings(tc.value)+`}}`,
				"",
			)
			tls, _ := stream["tlsSettings"].(map[string]any)
			if tls == nil {
				t.Fatalf("tlsSettings missing: %#v", stream)
			}
			got, ok := tls["allowInsecure"]
			if tc.wantSet {
				if !ok {
					t.Fatalf("tlsData should carry allowInsecure: %#v", tls)
				}
				if b, isBool := got.(bool); !isBool || !b {
					t.Fatalf("allowInsecure = %#v, want true", got)
				}
			} else if ok {
				t.Fatalf("tlsData must not carry allowInsecure, got %#v", got)
			}
		})
	}
}

func tlsAllowInsecureSettings(value string) string {
	settings := `{"fingerprint":"chrome"}`
	if value != "" {
		settings = `{"fingerprint":"chrome","allowInsecure":` + value + `}`
	}
	return settings
}
