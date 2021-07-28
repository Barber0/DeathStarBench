package common_test

import (
	"fmt"
	"net/url"
	"testing"
)

func TestName(t *testing.T) {
	myUrl := url.Values{
		"abc": {"123"},
		"ddd": {"456"},
	}
	u := url.URL{
		Scheme: "http",
		Host:   "www.baidu.com",
		Path: url.Values{
			"abc": {"ddd"},
			"def": {"123"},
		}.Encode(),
	}
	fmt.Println(u.String())

	fmt.Println(myUrl.Encode())
}
