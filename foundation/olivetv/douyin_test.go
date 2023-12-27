package olivetv_test

import (
	"testing"

	"github.com/go-olive/olive/foundation/olivetv"
)

func TestDouyin_Snap(t *testing.T) {
	u := "https://live.douyin.com/461940043501"
	// cookie := `__ac_nonce=06487158a00d5c2d4634c; __ac_signature=_02B4Z6wo00f01CjqXDgAAIDDeJ0YwjlhZKAoyliAAG7EcZEUW2MdeGRISr2fobBUEzpAtX24xyQL5JQQzCHcosKQXhCMWI0W4MacAadtfeEO0QnuggwfMl6vwiakM6a3ROFyKBXeWrGE1FRZ87;`
	dy, err := olivetv.NewWithURL(u)
	if err != nil {
		println(err.Error())
		return
	}
	dy.Snap()
	t.Log(dy)
}
