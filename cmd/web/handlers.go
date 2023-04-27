package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"subscription-service/data"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	pathToManual = "./pdf"
	tmpPath      = "./tmp"
)

const (
	userID      = "userID"
	userDataKey = "user"

	loginPage    = "/login"
	registerPage = "/register"
	homePage     = "/"
	plansPage    = "/members/plans"
)

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Invalid credentials.")
		http.Redirect(w, r, loginPage, http.StatusSeeOther)
		return
	}

	validPassword, err := app.Models.User.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Invalid credentials.")
		http.Redirect(w, r, loginPage, http.StatusSeeOther)
		return
	}

	if !validPassword {

		app.sendEmail(Message{
			To:      email,
			Subject: "Failed login attempt",
			Data:    "Invalid login attempt!",
		})

		app.Session.Put(r.Context(), notifyError, "Invalid credentials.")
		http.Redirect(w, r, loginPage, http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), userID, user.ID)
	app.Session.Put(r.Context(), userDataKey, user)

	app.Session.Put(r.Context(), notifyFlash, "Successful login!")

	http.Redirect(w, r, homePage, http.StatusSeeOther)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	// clean up session
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	http.Redirect(w, r, loginPage, http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	u := data.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName:  r.Form.Get("last-name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = app.Models.User.Insert(u)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Unable to create user.")
		http.Redirect(w, r, registerPage, http.StatusSeeOther)
		return
	}

	// send an activation email
	url := fmt.Sprintf("%s/activate?email=%s", os.Getenv("HTTP_HOST"), u.Email)
	signedURL := GenerateSignedURL(url)

	app.sendEmail(Message{
		To:       u.Email,
		Subject:  "Activate your account",
		Template: "confirmation-email",
		Data:     template.HTML(signedURL),
	})

	app.Session.Put(r.Context(), notifyFlash, "Confirmation email sent. Check your email.")
	http.Redirect(w, r, loginPage, http.StatusSeeOther)
}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// validate url
	url := r.RequestURI
	testURL := fmt.Sprintf("%s%s", os.Getenv("HTTP_HOST"), url)
	okay := VerifyToken(testURL)

	if !okay {
		app.Session.Put(r.Context(), notifyError, "Invalid token.")
		http.Redirect(w, r, homePage, http.StatusSeeOther)
		return
	}

	u, err := app.Models.User.GetByEmail(r.URL.Query().Get("email"))
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "No user found.")
		http.Redirect(w, r, homePage, http.StatusSeeOther)
		return
	}

	u.Active = 1
	err = app.Models.User.Update(*u)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Unable to update user.")
		http.Redirect(w, r, homePage, http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), notifyFlash, "Account activated. You can now log in.")
	http.Redirect(w, r, loginPage, http.StatusSeeOther)
}

func (app *Config) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	planID, err := strconv.Atoi(id)
	if err != nil {
		app.ErrorLog.Println("Error getting plan id:", err)
	}

	plan, err := app.Models.Plan.GetOne(planID)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Unable to find plan.")
		http.Redirect(w, r, plansPage, http.StatusSeeOther)
		return
	}

	user, ok := app.Session.Get(r.Context(), userDataKey).(data.User)
	if !ok {
		app.Session.Put(r.Context(), notifyError, "Log in first!")
		http.Redirect(w, r, loginPage, http.StatusSeeOther)
		return
	}

	// email invoice
	app.sendEmail(Message{
		To:       user.Email,
		Subject:  "Your invoice",
		Data:     plan.AmountForDisplay(),
		Template: "invoice",
	})

	// generate and email manual
	app.Wait.Add(1)
	go func() {
		defer app.Wait.Done()

		pdf := app.generateManual(user, plan)
		err := pdf.OutputFileAndClose(fmt.Sprintf("%s/%d_manual.pdf", tmpPath, user.ID))
		if err != nil {
			app.ErrorChan <- err
			return
		}

		app.sendEmail(Message{
			To:      user.Email,
			Subject: "Your manual",
			Data:    "Your user manual is attached",
			Attachments: []*mail.File{
				{
					FilePath: fmt.Sprintf("%s/%d_manual.pdf", tmpPath, user.ID),
					Name:     "Manual.pdf",
				},
			},
		})

	}()

	err = app.Models.Plan.SubscribeUserToPlan(user, *plan)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Error subscribing to plan!")
		http.Redirect(w, r, plansPage, http.StatusSeeOther)
		return
	}

	u, err := app.Models.User.GetOne(user.ID)
	if err != nil {
		app.Session.Put(r.Context(), notifyError, "Error getting user from database!")
		http.Redirect(w, r, plansPage, http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), userDataKey, u)

	app.Session.Put(r.Context(), notifyFlash, "Subscribed!")
	http.Redirect(w, r, plansPage, http.StatusSeeOther)
}

func (app *Config) generateManual(u data.User, plan *data.Plan) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)

	importer := gofpdi.NewImporter()

	t := importer.ImportPage(pdf, fmt.Sprintf("%s/manual.pdf", pathToManual), 1, "/MediaBox")
	pdf.AddPage()

	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	pdf.SetX(75)
	pdf.SetY(150)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s %s", u.FirstName, u.LastName), "", "C", false)
	pdf.Ln(5)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s User Guide", plan.PlanName), "", "C", false)

	return pdf
}

func (app *Config) ChooseSubscription(w http.ResponseWriter, r *http.Request) {
	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	dataMap := make(map[string]any)
	dataMap["plans"] = plans

	app.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: dataMap,
	})
}
