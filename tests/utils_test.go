package tests

import (
	"testing"

	"github.com/hmkwizu/ngauth"
)

func TestGetStringOrEmpty(t *testing.T) {

	//valid string
	result := ngauth.GetStringOrEmpty("Hello")
	if result != "Hello" {
		t.Fail()
	}

	//empty string
	result = ngauth.GetStringOrEmpty("")
	if len(result) != 0 {
		t.Fail()
	}

	//nil
	result = ngauth.GetStringOrEmpty(nil)
	if len(result) != 0 {
		t.Fail()
	}

	//numbers are converted to string
	result = ngauth.GetStringOrEmpty(1)
	if len(result) == 0 {
		t.Fail()
	}
}

func TestGetInt64OrZero(t *testing.T) {

	//valid int
	result := ngauth.GetInt64OrZero(1)
	if result != 1 {
		t.Fail()
	}

	//valid but is string
	result = ngauth.GetInt64OrZero("1")
	if result != 1 {
		t.Fail()
	}

	//nil
	result = ngauth.GetInt64OrZero(nil)
	if result != 0 {
		t.Fail()
	}

	//not an int
	result = ngauth.GetInt64OrZero("hello")
	if result != 0 {
		t.Fail()
	}
}

func TestBcrypt(t *testing.T) {

	pwd := "1234"
	pwd2 := "12345"

	//success
	hash := ngauth.BcryptHashMake(pwd)
	flag := ngauth.BcryptHashCheck(hash, pwd)

	if flag != true {
		t.Fail()
	}

	//should fail
	flag = ngauth.BcryptHashCheck(hash, pwd2)
	if flag != false {
		t.Fail()
	}
}
