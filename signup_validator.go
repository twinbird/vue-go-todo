package main

import (
	"log"
	"regexp"
)

type SignupValidator struct {
	Email           string
	Password        string
	PasswordConfirm string
}

func (v *SignupValidator) validate() (bool, []string) {
	var msg []string

	// check already exist user
	b, err := isExistUser(v.Email)
	if err != nil {
		log.Println(err)
		return false, []string{"エラーが発生しました。時間をおいてもう一度やり直してください。"}
	}
	if b == true {
		return false, []string{"既に使われているメールアドレスです。"}
	}

	// Email
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if v.Email == "" {
		msg = append(msg, "メールアドレスを入力してください。")
	} else if re.MatchString(v.Email) == false {
		msg = append(msg, "メールアドレスの形式が誤っています。")
	}

	// Password
	if v.Password == "" {
		msg = append(msg, "パスワードを入力してください。")
	} else if len(v.Password) < 8 {
		msg = append(msg, "パスワードは8文字以上にしてください。")
	}
	if v.Password != v.PasswordConfirm {
		msg = append(msg, "パスワード再入力欄はパスワード欄と同じ文字を入力してください。")
	}

	if len(msg) > 0 {
		return false, msg
	} else {
		return true, msg
	}
}
