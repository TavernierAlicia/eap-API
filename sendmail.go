package main

import (
	"fmt"
	"net/smtp"
	"time"

	eapFact "github.com/TavernierAlicia/eap-FACT"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func addPWD(subForm Subscription, token string) (err error) {
	to := subForm.Mail
	from := viper.GetString("sendmail.service_mail")
	pass := viper.GetString("sendmail.service_pwd")

	subject := "Bienvenue chez EAP - créez votre mot de passe"

	message := "Bonjour " + subForm.Civility + " " + subForm.Name + " " + subForm.Surname + ", votre compte est fin prêt! Vous pouvez maintenant cliquer sur le lien suivant afin de créer votre mot de passe: " + viper.GetString("links.create_pwd") + "?token=" + token

	msg := "From: " + from + " " + "\n" + "To: " + to + "\n" + "Subject: " + subject + "\n\n" + message

	err = smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"), from, []string{to}, []byte(msg))

	if err != nil {
		fmt.Println("smtp error %s", err)
	} else {
		fmt.Println("mail okay")
	}
	return err
}

func newPWD(ownerInfos Owner, token string) (err error) {
	to := ownerInfos.Mail
	from := viper.GetString("sendmail.service_mail")
	pass := viper.GetString("sendmail.service_pwd")

	subject := "Votre nouveau mot de passe"

	message := "Bonjour " + ownerInfos.Civility + " " + ownerInfos.Name + " " + ownerInfos.Surname + " vous avez demandé à créer un nouveau mot de passe pour l'établissement suivant: " + ownerInfos.Entname + " Siret: " + ownerInfos.Siret + ", " + ownerInfos.Addr + ", " + ownerInfos.City + ", cliquez sur le lien suivant pour créer un nouveau mot de passe: " + viper.GetString("links.create_pwd") + "?token=" + token

	msg := "From: " + from + " " + "\n" + "To: " + to + "\n" + "Subject: " + subject + "\n\n" + message

	err = smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, pass, "smtp.gmail.com"), from, []string{to}, []byte(msg))

	if err != nil {
		fmt.Println("smtp error %s", err)
	} else {
		fmt.Println("mail okay")
	}
	return err
}

func sendCliFact(link string, mail string) (err error) {
	to := mail
	from := viper.GetString("sendmail.service_mail")
	pass := viper.GetString("sendmail.service_pwd")

	subject := "Votre commande du " + time.Now().Format("02-01-2006 15:04:05")

	message := `
	<h1>Bonjour, Vous trouverez votre facture au format pdf ci-jointe, à bientôt sur Easy As Pie! </h1> 

	<h2>Facture n°?</h2>
	
	</br>
	<table style='border: 1px solid black; margin-right:10px;'>
			<tr>
				<th><b>Quantité</b></th>
				<th><b>Produit</b></th>
				<th><b>Prix Unitaire €</b></th>
				<th><b>Montant € </b></th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td style='border:none'>2</td>
				<td>Jus d'orange</td>
				<td style='border:none'>10.00</td>
				<td style='border:none'>20.00</td>
			</tr>
		</tbody>
		<tr>
			<th></br></br>TOTAL EUROS </b></th>
			<th></br></br></b></th>
			<th></br></br></b></th>
			<th></br></br>20.00</b></th>
		</tr>
		<tr>
			<th>TVA 20%</th>
			<th></br></br></b></th>
			<th></br></br></b></th>
			<th>4.00</th>
		</tr>
	
	</table>
	
	<p>Nous vous souhaitons une agréable journée!</p>`

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)
	m.Attach(link)

	d := gomail.NewPlainDialer("smtp.gmail.com", 587, from, pass)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return err
}

func sendBossFact(etab eapFact.FactEtab) (err error) {
	to := etab.Mail
	from := viper.GetString("sendmail.service_mail")
	pass := viper.GetString("sendmail.service_pwd")

	subject := "Facturation du " + etab.Fact_infos.Date

	message := "Bonjour, " + etab.Owner_civility + " " + etab.Owner_name + ", vous trouverez votre facture du " + etab.Fact_infos.Date + " au format pdf ci-jointe, à bientôt sur Easy As Pie! "

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)
	m.Attach(etab.Fact_infos.Link)

	d := gomail.NewPlainDialer("smtp.gmail.com", 587, from, pass)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return err
}
