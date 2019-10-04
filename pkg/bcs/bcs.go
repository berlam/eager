package bcs

import (
	"eager/pkg"
	"fmt"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

const (
	bcsLogin             = "/bcs/login/*/display?compose_login_form_only=false&login_form_name=pageform"
	bcsLogout            = "/bcs/login/*/display?logout=true"
	bcsShowEffort        = "/bcs/mybcs/effortlist/display?effortlist,setting=%s&effortlist,Selections,effortDate,day=1&effortlist,Selections,effortDate,month=%s&effortlist,Selections,effortDate,year=%s&effortlist,Selections,effortDate,mode=M"
	bcsGetEffort         = "/bcs/mybcs/effortlist/display/Buchungen.csv?download=component&downloadcontent=formatted&object=effortlist"
	bcsShowProjectEffort = "/bcs/projectdetail/efforts/display?oid=%s&efforts,Choices,effortlist,setting=%s&efforts,Choices,effortlist,Selections,peroid,day=1&efforts,Choices,effortlist,Selections,peroid,month=%s&efforts,Choices,effortlist,Selections,peroid,year=%s"
	bcsGetProjectEffort  = "/bcs/projectdetail/efforts/display/Buchungen.csv?download=component&downloadcontent=formatted&object=efforts%2CChoices%2Ceffortlist"
)

func GetTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, report string) pkg.Timesheet {
	client.Jar, _ = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	err := login(client, server, userinfo)
	if err != nil {
		log.Println("Login did not succeed.", err)
		return pkg.Timesheet{}
	}
	defer func() {
		err = logout(client, server)
		if err != nil {
			log.Println("Logout did not succeed.", err)
		}
	}()

	err = showEffortList(client, server, url.QueryEscape(report), month, year)
	if err != nil {
		log.Println("Could not show effort list.", err)
		return pkg.Timesheet{}
	}

	data, err := retrieveEffortList(client, server)
	if err != nil {
		log.Println("Could not retrieve effort list.", err)
		return pkg.Timesheet{}
	}

	timesheet := pkg.Timesheet{}
	spec := pkg.NewCsvSpecification().Header(true).Project(true).Task(true).Description(true).Date(true).Duration(true)
	timesheet, err = timesheet.ReadCsv(data, &spec)
	if err != nil {
		log.Println("Could not read effort list.", err)
	}

	return timesheet
}

func GetBulkTimesheet(client *http.Client, server *url.URL, userinfo *url.Userinfo, year int, month time.Month, project pkg.Project, report string) pkg.Timesheet {
	client.Jar, _ = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	err := login(client, server, userinfo)
	if err != nil {
		log.Println("Login did not succeed.", err)
		return pkg.Timesheet{}
	}
	defer func() {
		err = logout(client, server)
		if err != nil {
			log.Println("Logout did not succeed.", err)
		}
	}()

	err = showProjectEffortList(client, server, url.QueryEscape(report), month, year, project)
	if err != nil {
		log.Println("Could not show effort list.", err)
		return pkg.Timesheet{}
	}

	data, err := retrieveProjectEffortList(client, server)
	if err != nil {
		log.Println("Could not retrieve effort list.", err)
		return pkg.Timesheet{}
	}

	timesheet := pkg.Timesheet{}
	spec := pkg.NewCsvSpecification().Header(true).User(true).Project(true).Task(true).Description(true).Date(true).Duration(true).Skip()
	timesheet, err = timesheet.ReadCsv(data, &spec)
	if err != nil {
		log.Println("Could not read effort list.", err)
	}

	return timesheet
}

func login(client *http.Client, server *url.URL, auth *url.Userinfo) error {
	password, _ := auth.Password()
	loginUrl, _ := server.Parse(bcsLogin)
	resp, err := client.PostForm(loginUrl.String(), url.Values{
		"user":               {auth.Username()},
		"pwd":                {password},
		"isPassword":         {"pwd"},
		"!set_login_cookies": {"true"},
		"login":              {"Timesheet+In"},
	})
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func logout(client *http.Client, server *url.URL) error {
	logoutUrl, _ := server.Parse(bcsLogout)
	resp, err := client.Get(logoutUrl.String())
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func showEffortList(client *http.Client, server *url.URL, report string, month time.Month, year int) error {
	showEffortUrl, _ := server.Parse(fmt.Sprintf(bcsShowEffort, report, strconv.Itoa(int(month)), strconv.Itoa(year)))
	resp, err := client.Get(showEffortUrl.String())
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func showProjectEffortList(client *http.Client, server *url.URL, report string, month time.Month, year int, project pkg.Project) error {
	showEffortUrl, _ := server.Parse(fmt.Sprintf(bcsShowProjectEffort, url.QueryEscape(string(project)), report, strconv.Itoa(int(month)), strconv.Itoa(year)))
	resp, err := client.Get(showEffortUrl.String())
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func retrieveEffortList(client *http.Client, server *url.URL) ([]byte, error) {
	retrieveEffortUrl, _ := server.Parse(bcsGetEffort)
	resp, err := client.Get(retrieveEffortUrl.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	reader, _ := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	return data, nil
}

func retrieveProjectEffortList(client *http.Client, server *url.URL) ([]byte, error) {
	retrieveEffortUrl, _ := server.Parse(bcsGetProjectEffort)
	resp, err := client.Get(retrieveEffortUrl.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	reader, _ := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	return data, nil
}
