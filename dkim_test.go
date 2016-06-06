package dkim

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

var dkimSampleEML1 string = "A: X \r\n" +
	"B : Y\t\r\n" +
	"\tZ  \r\n" +
	"\r\n" +
	" C \r\n" +
	"D \t E\r\n" +
	"\r\n" +
	"\r\n"

var dkimSampleEML2 string = "From: Joe SixPack <joe@football.example.com>\r\n" +
	"To: Suzie Q <suzie@shopping.example.net>\r\n" +
	"Subject: Is dinner ready?\r\n" +
	"Date: Fri, 11 Jul 2003 21:00:37 -0700 (PDT)\r\n" +
	"Message-ID: <20030712040037.46341.5F8J@football.example.com>\r\n" +
	"\r\n" +
	"Hi.\r\n" +
	"\r\n" +
	"We lost the game. Are you hungry yet?\r\n" +
	"\r\n" +
	"Joe.\r\n"

var dkimSampleEML3 string = "Return-Path: aws@s3ig.com\r\n" +
	"MIME-Version: 1.0\r\n" +
	"From: aws@s3ig.com\r\n" +
	"To: check-auth@verifier.port25.com\r\n" +
	"Reply-To: aws@s3ig.com\r\n" +
	"Date: 10 Mar 2011 10:41:56 +0000\r\n" +
	"Subject: dkim test email\r\n" +
	"Content-Type: text/plain; charset=us-ascii\r\n" +
	"Content-Transfer-Encoding: quoted-printable\r\n" +
	"\r\n" +
	"This is the body of \t the message.=0D=0AThis is the second line\r\n" +
	"\r\n"

var dkimSamplePEMData []byte = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDptLAWO6YJyVGKwayFzVCEIhY0bjE4ObCr7JeILVWi9ELaWwdm
PEKW0afsqB/jR+yhMIdeyGbW9li4P2X8oFDLFwAkAng857ngz2RYNjf+7bGgvH8n
gbKXuFtnO1kI5+ou3C1+Tixk8aY44uWCFamXueAW1ZzlzexViyG4gdVXlQIBEQKB
gQDb9VpvRzLcCMU3TN6cDIgD49ip0R9D+g+w3qy8Zucv9POgVaycdPNgxVLAnjwh
NKJ5lxX+2rskq57LhvaTabVyDNkRjp8Oks1+nCbklaCEdkJ0S9z/IqjmmqCY4bP1
kJ0Eu18NHBId8dXoK3EQ8nzRCKE+tjkiMl5jC1kJzVdg8QJBAPrDkKMURlLRnMEA
LzXHwoNAPdK94IFwdypq62yDVm4u2rSMUtYlKmJQFiSkytQ2vjBb5HRxqL+p4mnq
30ZHWdUCQQDulfC32vcY7e2IetYhda+sysdZJnfrbquJpdlfBn2QFH8gjC2KM/q+
YtwQGJU/zjtwWN+/joi4vinlKD7RYSbBAkEAk4IY2GZHfALUrcPfiQwYEPic1lGT
Hvbcr4owIbarT99TeUN8BX9GG7ajnRWkfNToWK6GYp02FmPumKhHGkgWuQJBAKhp
1xheVBGY4+fePMxTEpgWqtWEkOJsPNmiPxXmdsAOd9q9TVJ/C1k2uXTGDv/c3qmo
JXgoYIJoHZKy/ypisfECQQC9FxTwEfzFLTANQVnUQKEbRUk3slaigQb3QoBKXOZr
oCyn+rxDcflW1RbZnsilaMbpN/PMw/IbqRjXA2Tg3Ty6
-----END RSA PRIVATE KEY-----`)

func TestNew(t *testing.T) {
	dkim, err := New(Conf{}, nil)
	if err == nil {
		t.Fatal(err)
	}
	if dkim != nil {
		t.Fatal(dkim)
	}
	conf, err := NewConf("domain", "selector")
	if err != nil {
		t.Fatal(err)
	}
	dkim, err = New(conf, nil)
	if err == nil {
		t.Fatal(err)
	}
	dkim, err = New(conf, dkimSamplePEMData)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalBody(t *testing.T) {
	dkim := &DKIM{}
	msg, err := readEML([]byte(" "))
	if err != nil {
		t.Fatal(err)
	}
	body := dkim.canonicalBody(msg)
	if x := string(body); x != "\r\n" {
		t.Fatal("uninitialized struct should yield CRLF body", x)
	}

	conf, _ := NewConf("domain", "selector")
	conf[CanonicalizationKey] = "relaxed/relaxed"
	dkim, _ = New(conf, dkimSamplePEMData)
	msg, _ = readEML([]byte(dkimSampleEML1))
	dkim.readBody(msg)
	if x := string(dkim.canonicalBody(msg)); x != " C\r\nD E\r\n" {
		t.Fatal(x)
	}

	msg, _ = readEML([]byte(dkimSampleEML1))
	conf[CanonicalizationKey] = "relaxed/simple"
	if x := string(dkim.canonicalBody(msg)); x != " C \r\nD \t E\r\n" {
		t.Fatal(x)
	}

	msg, _ = readEML([]byte(dkimSampleEML2))
	dkim.readBody(msg)
	conf[CanonicalizationKey] = "relaxed/simple"
	if x := string(dkim.canonicalBody(msg)); x != "Hi.\r\n\r\nWe lost the game. Are you hungry yet?\r\n\r\nJoe.\r\n" {
		t.Fatal(x)
	}

	msg, _ = readEML([]byte(dkimSampleEML3))
	dkim.readBody(msg)
	conf[CanonicalizationKey] = "relaxed/relaxed"
	if x := string(dkim.canonicalBody(msg)); x != "This is the body of the message.=0D=0AThis is the second line\r\n" {
		t.Fatal(x)
	}
}

func TestCanonicalBodyHash(t *testing.T) {
	conf, _ := NewConf("domain", "selector")
	conf[CanonicalizationKey] = "relaxed/simple"

	dkim, _ := New(conf, dkimSamplePEMData)

	msg, _ := readEML([]byte(dkimSampleEML2))
	enc := base64.StdEncoding
	dkim.readBody(msg)
	if x := enc.EncodeToString(dkim.canonicalBodyHash(msg)); x != "2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=" {
		t.Fatal(x)
	}

	msg, _ = readEML([]byte(dkimSampleEML3))
	conf[CanonicalizationKey] = "relaxed/relaxed"
	dkim.readBody(msg)
	if x := enc.EncodeToString(dkim.canonicalBodyHash(msg)); x != "vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=" {
		t.Fatal(x)
	}

	msg, err := readEML([]byte(" "))
	if err != nil {
		t.Fatal(err)
	}
	dkim.readBody(msg)

	// Simple canonical empty body
	conf[CanonicalizationKey] = "relaxed/simple"
	if x := enc.EncodeToString(dkim.canonicalBodyHash(msg)); x != "frcCV1k9oG9oKj3dpUqdJg1PxRT2RSN/XKdLCPjaYaY=" {
		t.Fatal(x)
	}

	// Relaxed canonical empty body
	conf[CanonicalizationKey] = "relaxed/relaxed"
	if x := enc.EncodeToString(dkim.canonicalBodyHash(msg)); x != "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=" {
		t.Fatal(x)
	}
}

func dkimSignable() *DKIM {
	conf, _ := NewConf("s3ig.com", "dkim")
	conf[CanonicalizationKey] = "relaxed/relaxed"
	conf[TimestampKey] = "1299753716"
	dkim, _ := New(conf, dkimSamplePEMData)
	dkim.signableHeaders = []string{
		"Content-Type",
		"From",
		"Subject",
		"To",
	}
	return dkim
}

func TestSignableHeaderBlock(t *testing.T) {
	msg, _ := readEML([]byte(dkimSampleEML3))
	dkim := dkimSignable()
	dkim.readBody(msg)
	block := dkim.signableHeaderBlock(msg)
	expect := "content-type:text/plain; charset=us-ascii\r\n" +
		"from:aws@s3ig.com\r\n" +
		"subject:dkim test email\r\n" +
		"to:check-auth@verifier.port25.com\r\n" +
		"dkim-signature:v=1; a=rsa-sha256; c=relaxed/relaxed;" +
		" d=s3ig.com; q=dns/txt; s=dkim; t=1299753716;" +
		" bh=vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=;" +
		" h=Content-Type:From:Subject:To; b="
	if block != expect {
		t.Fatal(fmt.Sprintf("signable header block invalid:\n\n%q\n\nshould be\n\n%q",
			block,
			expect))
	}
}

func TestSignature(t *testing.T) {
	msg, _ := readEML([]byte(dkimSampleEML3))
	dkim := dkimSignable()
	dkim.readBody(msg)
	sig, err := dkim.signature(msg)
	if err != nil {
		t.Fatal("error not nil", err)
	}
	if sig != "enIert1AWY8K9AIxTw0qQLOO3TKuRENfJvwYWDXi6xM7IWaz+Bb83xi5YnjBH0Q8opLn643qIaXGVIU2+LBA2a44PZGtTRXYMG3sbQpcEMjfJRPAhAQOazsSlVdq4SmAChAU3g8uPj4r71JdROucZSdm/mW8IoT4IympoCiLKdQ=" {
		t.Fatal("signature invalid", sig)
	}
}

func TestSignedEML(t *testing.T) {
	signed, err := dkimSignable().Sign([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal("error not nil", err)
	}
	expect := "Return-Path: aws@s3ig.com\r\n" +
		"MIME-Version: 1.0\r\n" +
		"From: aws@s3ig.com\r\n" +
		"To: check-auth@verifier.port25.com\r\n" +
		"Reply-To: aws@s3ig.com\r\n" +
		"Date: 10 Mar 2011 10:41:56 +0000\r\n" +
		"Subject: dkim test email\r\n" +
		"Content-Type: text/plain; charset=us-ascii\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed; d=s3ig.com; q=dns/txt; s=dkim;" +
		" t=1299753716; bh=vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=;" +
		" h=Content-Type:From:Subject:To;" +
		" b=enIert1AWY8K9AIxTw0qQLOO3TKuRENfJvwYWDXi6xM7IWaz+Bb83xi5YnjBH0Q8opLn643qIaXGVIU2+LBA2a44PZGtTRXYMG3sbQpcEMjfJRPAhAQOazsSlVdq4SmAChAU3g8uPj4r71JdROucZSdm/mW8IoT4IympoCiLKdQ=\r\n" +
		"\r\n" +
		"This is the body of \t the message.=0D=0AThis is the second line\r\n" +
		"\r\n"

	expectMsg, _ := readEML([]byte(expect))
	xMsg, _ := readEML(signed)

	for k, _ := range expectMsg.Header {
		e := xMsg.Header[k][0]
		v := expectMsg.Header[k][0]
		if v != e {
			t.Fatalf("\n%q\n----\n%q", v, e)
		}
	}

	ebuf := new(bytes.Buffer)
	ebuf.ReadFrom(expectMsg.Body)

	xbuf := new(bytes.Buffer)
	xbuf.ReadFrom(xMsg.Body)

	if ebuf.String() != xbuf.String() {
		t.Fatalf("\n%q\n----\n%q", xbuf.String(), ebuf.String())
	}
}
