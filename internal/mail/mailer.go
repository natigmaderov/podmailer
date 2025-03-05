package mail

import (
	"crypto/tls"
	"fmt"
	"strings"

	podmailerv1alpha1 "github.com/natigmaderov/podmailer/api/v1alpha1"
	"gopkg.in/gomail.v2"
)

// Mailer handles email sending operations
type Mailer struct {
	config podmailerv1alpha1.SMTPConfig
}

// NewMailer creates a new Mailer instance
func NewMailer(config podmailerv1alpha1.SMTPConfig) *Mailer {
	return &Mailer{
		config: config,
	}
}

// SendPodDownNotification sends an email notification about down pods
func (m *Mailer) SendPodDownNotification(recipients []string, downPods []podmailerv1alpha1.PodStatus) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.FromEmail)
	msg.SetHeader("To", recipients...)
	msg.SetHeader("Subject", "Kubernetes Pod Alert: Pods are down!")
	msg.SetBody("text/plain", formatPodDownMessage(downPods))

	// Create the dialer with SMTP server settings
	dialer := gomail.NewDialer(
		m.config.Server,
		int(m.config.Port),
		m.config.Username,
		m.config.Password,
	)

	// Configure TLS settings
	dialer.SSL = false // Disable SSL since secure: false in original config
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         m.config.Server,
	}

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func formatPodDownMessage(pods []podmailerv1alpha1.PodStatus) string {
	var sb strings.Builder

	sb.WriteString("The following pods are currently down:\n\n")

	for _, pod := range pods {
		sb.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
		sb.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
		sb.WriteString(fmt.Sprintf("Status: %s\n", pod.Status))
		sb.WriteString("-------------------\n")
	}

	return sb.String()
}
