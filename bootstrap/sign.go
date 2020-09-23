package bootstrap

import (
	"crypto/md5"
	"fmt"
	"github.com/webrtc-demo-go/config"
)

// 根据当前时间（毫秒），生成用于开放平台Restful token请求的签名
func calTokenSign(ts int64) string {
	data := fmt.Sprintf("%s%s%d", config.App.ClientID, config.App.Secret, ts)

	val := md5.Sum([]byte(data))

	// md5值转换为大写
	res := fmt.Sprintf("%X", val)
	return res
}

// 根据当前时间（毫秒），生成用于开放平台Restful 业务请求的签名
func calBusinessSign(ts int64) string {
	data := fmt.Sprintf("%s%s%s%d", config.App.ClientID, config.App.AccessToken, config.App.Secret, ts)

	val := md5.Sum([]byte(data))

	// md5值转换为大写
	res := fmt.Sprintf("%X", val)
	return res
}
