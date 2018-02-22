package qutil

import (
	"fmt"
)

func ExampleCIDR2IpRange() {
	ipRange, _ := CIDR2IpRange("62.76.47.12/31")
	fmt.Printf("%#v", ipRange)
	// Output: []string{"62.76.47.12", "62.76.47.13"}
}

func ExampleRot13() {
	i := "a to zed"
	fmt.Println(Rot13(i))
	fmt.Println(Rot13(Rot13(i)))
	// Output:
	// n gb mrq
	// a to zed
}

func ExampleBytesMerge() {
	a := []byte("abc")
	b := []byte("123")
	c := []byte("!@#")

	fmt.Printf("%s", BytesMerge(a, b, c))
	// Output: abc123!@#
}

func ExampleGetHost() {
	items := []string{
		"",
		"a.b.c.d.com",
		"a.b.c.d.com.",
		" a.b.C.D.com.",
		"http://a.b.c.d.com/",
		"http://a.b.c.d.com",
		"http://a.b.c.d.com.//\r\n/\t/",
		"http://a.b.c.d.com.//?C=1&a=2&&d=4&sid=pp&B=3&",
		"http://%31%36%38%2e%31%38%38%2e%39%39%2e%32%36/%2E/",
		"http://%20leadingspace.com/",
		"  http://www.google.com/  ",
		"http:// leadingspace.com/",
		"http://host/%%%25%32%35asd%%",
	}

	for _, u := range items {
		fmt.Println(GetHost(u))
	}
	// Output:
	//
	// a.b.c.d.com
	// a.b.c.d.com
	// a.b.c.d.com
	// a.b.c.d.com
	// a.b.c.d.com
	// a.b.c.d.com
	// a.b.c.d.com
	// %31%36%38%2e%31%38%38%2e%39%39%2e%32%36
	// %20leadingspace.com
	// www.google.com
	// leadingspace.com
	// host
}
