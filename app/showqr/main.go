package main

import (
	"log"
	"os"

	"bytes"
	"image/png"
	"path"

	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/configuration"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store"

	"github.com/balasanjay/totp"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Login string `short:"l" long:"Login" description:"user Login" required:"true"`
	Label string `short:"t" long:"text" description:"Label" required:"true"`
}

func main() {
	opts := options{}
	var parser = flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		log.Fatalf("Unable to parse command line: %s", err.Error())
	}

	selfDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Cant get current working directory")
	}
	pathResolver := shared.NewPathResolver(selfDir, path.Join(selfDir, "../data"), path.Join(selfDir, "../etc"))

	configFileName := pathResolver.PathToAbs("${dir.config}/root.config")
	config, err := configuration.NewAppConfig(configFileName)
	if err != nil {
		log.Fatalf("parse config file [%v] failed; error:%v", configFileName, err.Error())
	}

	dbType := config.GetOrDie("db.type")
	dbConnectionString := config.GetOrDie("db.connectionString")
	db, err := store.NewUsersDb(dbType.(string), dbConnectionString.(string), pathResolver)
	if err != nil {
		log.Fatalf("db create failed: %v", err.Error())
	}

	if err = db.Open(); err != nil {
		log.Fatalf("Unable to open db: %s", err.Error())
	}

	defer db.Close()

	user, err := db.FindTOTPUserOptions(opts.Login)
	if err != nil {
		log.Fatalf("unable to find Login %s in db die to error:%s", opts.Login, err.Error())
	}

	qr, err := totp.BarcodeImage(opts.Label, user.Secret, &totp.Options{
		Hash:     user.Hash,
		Digits:   user.Digits,
		TimeStep: user.TimeStep,
		Tries:    user.Tries,
		Time:     user.Time,
	})

	if err != nil {
		log.Fatalf("Unable to create QR code: %s", err.Error())
	}

	const BLACK = "\033[40m  \033[0m"
	const WHITE = "\033[47m  \033[0m"

	w := os.Stdout
	img, _ := png.Decode(bytes.NewReader(qr))
	imgBox := img.Bounds()
	for i := 0; i < imgBox.Dx(); i += 8 {
		for j := 0; j < imgBox.Dy(); j += 8 {
			pix := img.At(i, j)
			r, g, b, _ := pix.RGBA()
			if r > 128 || g > 128 || b > 128 {
				w.Write([]byte(WHITE))
			} else {
				w.Write([]byte(BLACK))
			}
		}
		w.Write([]byte("\n"))
	}
	w.Write([]byte("\n"))
}
