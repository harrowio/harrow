package harrowMail

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/harrowio/harrow/hmail"
	"github.com/mohamedattahri/mail"
)

var (
	TemplateFunctions = template.FuncMap{
		"indent": func(with string, text string) string {
			source := bytes.NewBufferString(text)
			result := bytes.NewBufferString("")
			lines := bufio.NewScanner(source)
			for lines.Scan() {
				fmt.Fprintf(result, "%s%s\n", with, lines.Text())
			}

			return result.String()
		},
	}
)

func init() {
	mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
}

type Mail struct {
	TemplateDir    File
	AttachmentsDir File
	From           *mail.Address
}

func NewMail(from string, templateDir, attachmentsDir File) (*Mail, error) {
	fromAddr, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}
	return &Mail{
		TemplateDir:    templateDir,
		AttachmentsDir: attachmentsDir,
		From:           fromAddr,
	}, nil
}

func (self *Mail) Compose(ctxt *hmail.MailContext, to string) (*mail.Message, error) {
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return nil, err
	}
	message := mail.NewMessage()
	message.SetSubject(ctxt.Subject)
	message.SetFrom(self.From)
	message.To().Add(toAddr)
	multipart := mail.NewMultipart("multipart/alternative", message)

	displayName := toAddr.Name
	if displayName == "" {
		displayName = toAddr.Address
	}
	ctxt.Recipient.DisplayName = displayName

	parts, err := self.TemplateDir.Children()
	if err != nil {
		return nil, err
	}

	specificAttachmentDir := (File)(nil)
	for _, part := range parts {
		if filename := part.Name(); !strings.Contains(filename, ".tmpl") {
			if filepath.Base(filename) == "attachments" {
				specificAttachmentDir = part
			}
			continue
		}
		base := filepath.Base(part.Name())
		base = removeExtension(base)
		mimetype := mime.TypeByExtension("." + base)
		if mimetype == "" {
			return nil, fmt.Errorf("Could not determine mimetype for %q", base)
		}

		partBody, err := part.Open()
		if err != nil {
			return nil, err
		}
		templateContent, err := ioutil.ReadAll(partBody)
		partBody.Close()
		if err != nil {
			return nil, err
		}

		partTemplate, err := template.New(part.Name()).Funcs(TemplateFunctions).Parse(string(templateContent))
		if err != nil {
			return nil, err
		}
		renderedPart := new(bytes.Buffer)
		if err := partTemplate.Execute(renderedPart, ctxt); err != nil {
			return nil, err
		}
		multipart.AddText(mimetype, renderedPart)
	}

	// Add global attachments
	err = addAttachments(multipart, self.AttachmentsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	// Add mail-specific attachments
	err = addAttachments(multipart, specificAttachmentDir)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Add attachments from the given folder
// These can be referred to with the url cid:<%-encoded bassename of the attachment>
// for example: <img src="cid:logo.png">.
// Attachments are always refered to by basename in the mail.
func addAttachments(mp *mail.Multipart, attachmentDir File) error {
	if attachmentDir == nil {
		return nil
	}

	attachmentFiles, err := attachmentDir.Children()
	if err != nil {
		return err
	}

	if len(attachmentFiles) == 0 {
		return nil
	}

	for _, attachmentFile := range attachmentFiles {
		attachment, err := attachmentFile.Open()
		if err != nil {
			return err
		}
		if err := mp.AddAttachment(
			"inline",
			filepath.Base(attachmentFile.Name()),
			"", /* auto-mimetype */
			attachment,
		); err != nil {
			return err
		}
	}

	return nil
}

func removeExtension(filename string) string {
	return filename[0 : len(filename)-len(filepath.Ext(filename))]
}
