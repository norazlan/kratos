package hook

import (
	"net/http"
	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	_ login.PostHookExecutor        = new(Redirector)
	_ registration.PostHookExecutor = new(Redirector)
)

type Redirector struct {
	returnTo         func() *url.URL
	whitelist        func() []url.URL
	allowUserDefined func() bool
	publicURL        func() *url.URL
}

func NewRedirector(
	returnTo func() *url.URL,
	whitelist func() []url.URL,
	allowUserDefined func() bool,
	publicURL func() *url.URL,
) *Redirector {
	return &Redirector{
		returnTo:         returnTo,
		whitelist:        whitelist,
		allowUserDefined: allowUserDefined,
		publicURL:        publicURL,
	}
}

func (e *Redirector) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, sr *registration.Request, _ *session.Session) error {
	return e.do(w, r, sr.RequestURL)
}

func (e *Redirector) ExecuteSettingsPostHook(w http.ResponseWriter, r *http.Request, pr *settings.Request, _ *session.Session) error {
	return e.do(w, r, pr.RequestURL)
}

func (e *Redirector) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, sr *login.Request, _ *session.Session) error {
	return e.do(w, r, sr.RequestURL)
}

func (e *Redirector) do(w http.ResponseWriter, r *http.Request, originalURL string) error {
	ou, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return herodot.ErrInternalServerError.WithReasonf("The redirect hook was unable to parse the original request URL: %s", err)
	}

	var returnTo string
	if e.allowUserDefined() {
		returnTo, err = x.DetermineReturnToURL(ou, e.returnTo(), e.whitelist())
	} else {
		returnTo, err = x.DetermineReturnToURL(ou, e.returnTo(), []url.URL{*urlx.AppendPaths(e.publicURL(), "self-service")})
	}

	if err != nil {
		return err
	}

	http.Redirect(w, r, returnTo, http.StatusFound)
	return nil
}
