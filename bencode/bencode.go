package bencode

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/IncSW/go-bencode"
)

func ToMagnet(buf []byte) string {
	data, err := bencode.Unmarshal(buf)
	if err != nil {
		panic(err)
	}
	buf, err = bencode.Marshal(data.(map[string]interface{})["info"])
	if err != nil {
		panic(err)
	}
	sh := sha1.New()
	sh.Write(buf)
	hash := hex.EncodeToString(sh.Sum(nil))

	var trackers string
	dt := data.(map[string]interface{})["announce-list"]
	for _, buf := range dt.([]interface{}) {
		trackers += "&tr=" + string(buf.([]interface{})[0].([]byte))
	}
	name := string(data.(map[string]interface{})["info"].(map[string]interface{})["name"].([]byte))
	return "magnet:?xt=urn:btih:" + hash + "&dn=" + name + trackers
}
