package main

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
)

func AddPWD(subForm Subscription, token string) (err error) {
	to := subForm.Mail
	from := viper.GetString("sendmail.service_mail")
	pass := viper.GetString("sendmail.service_pwd")

	subject := "Bienvenue chez EAP - créez votre mot de passe"

	message := "Bonjour " + subForm.Civility + " " + subForm.Name + " " + subForm.Surname + ", votre compte est fin prêt! Vous pouvez maintenant clmiquer sur le lien suivant afin de créer votre mot de passe: " + viper.GetString("links.create_pwd") + token

	msg := "From: " + from + " " + "\n" + "To: " + to + "\n" + "Subject: " + subject + "\n\n" + message

	err = smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"), from, []string{to}, []byte(msg))

	if err != nil {
		fmt.Println("smtp error %s", err)
	} else {
		fmt.Println("mail okay")
	}
	return err
}
