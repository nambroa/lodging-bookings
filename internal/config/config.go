package config

import (
	"github.com/alexedwards/scs/v2"
	"github.com/nambroa/lodging-bookings/internal/models"
	"html/template"
	"log"
)

// AppConfig holds the configuration for this application.
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	Mailchan      chan models.MailData
}
