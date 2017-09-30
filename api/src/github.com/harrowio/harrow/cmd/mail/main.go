package harrowMail

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

const ProgramName = "mail"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	FromAddress := flag.String("from", "notifications@www.app.harrow.io", "From address to use in emails")
	AttachmentsFrom := flag.String("attachments", "global/attachments", "Directory containing attachments for every email")
	TemplateDir := flag.String("templates", "mail", "Directory containing mail templates")
	flag.Parse()
	ctxt := NewContext(*FromAddress)
	contextDir, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal().Err(err)
	}

	if err := ctxt.LoadFromDirectory(NewOSFile(contextDir)); err != nil {
		log.Fatal().Err(err)
	}

	mailTemplateDir, err := ctxt.MailTemplateDir(*TemplateDir)
	if err != nil {
		log.Fatal().Err(err)
	}

	attachmentsFrom := (File)(nil)
	if attachmentsDir, err := os.Open(*AttachmentsFrom); err != nil {
		log.Error().Msgf("os.open(%q): %s", *AttachmentsFrom, err)
	} else {
		attachmentsFrom = NewOSFile(attachmentsDir)
	}

	mailCtxt, err := ctxt.ToMailContext()
	if err != nil {
		log.Fatal().Err(err)
	}

	mailToSend, err := NewMail(ctxt.FromAddress, mailTemplateDir, attachmentsFrom)
	if err != nil {
		log.Fatal().Err(err)
	}

	message, err := mailToSend.Compose(mailCtxt, ctxt.ToAddress)
	if err != nil {
		log.Fatal().Err(err)
	}

	fmt.Println(string(message.Bytes()))
}
