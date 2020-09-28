package main
// export REPO_DIRECTORY_PATH="/home/shahid/XMLParser/"
// export REPO_NAME="XMLParser"
// export REPO_USER="ShahidYousuf"
// export INSTALL_DIRECTIVES_PATH="/home/shahid/XMLParser/config.json"
// export GITHUB_ACCESS_TOKEN="54b88be28372ccee6f9656dd3f33fb5096db61f7"

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)
type User struct {
	email string
	first_name string
	last_name string
	is_active bool
	login_count int
	location Location
}
type Configuration struct {
	AppConfig AppConfig `json:"app_config"`
	AppCommands []AppCommand `json:"app_commands"`
}
type AppConfig struct {
	RunInstallation    bool   `json:"run_installation"`
	ContinueOnFail     bool   `json:"continue_on_fail"`
	NotifyEmails       []string `json:"notify_emails"`
}

type AppCommand struct {
	Name string `json:"name"`
	RunDir string `json:"run_dir"`
	Command []string `json:"command"`
	runStatus bool
}

func ReadAppConfig() (Configuration, bool) {
	configFile, isPresent := os.LookupEnv("INSTALL_DIRECTIVES_PATH")
	if !isPresent {
		return Configuration{}, false
	}else {
		fileContent, fileReadErr := ioutil.ReadFile(configFile)
		if fileReadErr != nil {
			fmt.Println("Configuration file is missing. Please contact the admin.")
		}
		var configuration Configuration
		// read app_config config variables from json file
		err := json.Unmarshal(fileContent, &configuration)
		if err != nil {
			fmt.Println("Error reading config data.")
		}
		return configuration, true
	}
}

func FetchGitHubCommit(user string, repo string, accessToken string) (string, string, bool) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	//repos, _, _ := client.Repositories.List(context.Background(), "ShahidYousuf", nil)
	commitListOpts := github.CommitsListOptions{}
	sinceTime := time.Now().Add(-10*time.Minute)
	commitListOpts.Since = sinceTime
	commitListOpts.Until = time.Now()
	commits, _, _ := client.Repositories.ListCommits(ctx, user, repo, &commitListOpts)
	if len(commits) > 0 {
		return commits[0].GetCommit().GetMessage(), commits[0].GetSHA(), true
	}else{
		commits, _, _ = client.Repositories.ListCommits(ctx, user, repo, nil)
		if len(commits) > 0 {
			return commits[0].GetCommit().GetMessage(), commits[0].GetSHA(), true
		}else {
			return "", "", false
		}
	}
}

func RunAppCommands(commands []AppCommand, continueOnFail bool) ([]AppCommand, error) {
	runList := []AppCommand{}
	for _, appcmd := range commands {
		cmd := exec.Command(appcmd.Command[0], appcmd.Command[1:]...)
		cmd.Dir = appcmd.RunDir
		fmt.Println("Run: ", appcmd.Name)
		cmdError := cmd.Run()
		if cmdError != nil {
			fmt.Println("Error ", cmdError.Error())
			appcmd.runStatus = false
			runList = append(runList, appcmd)
			if continueOnFail == true {
				continue
			}else {
				return runList, cmdError
			}
		}else {
			appcmd.runStatus = true
			runList = append(runList, appcmd)
		}
	}
	return runList, nil
}

func getLocalCommit(repoPath string) string {
	chdirErr := os.Chdir(repoPath)
	if chdirErr != nil {
		fmt.Println("Cannot go to repo folder. Something went wrong!")
		os.Exit(1)
	}
	rep, _ := git.PlainOpen(repoPath)
	logOptions := git.LogOptions{Order: git.LogOrder(git.LogOrderCommitterTime)}
	citer, _ := rep.Log(&logOptions)
	commitSHAList := []string{}
	for {
		commit, err := citer.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		commitSHAList = append(commitSHAList, commit.Hash.String())
		if len(commitSHAList) > 3 {
			break
		}
	}
	citer.Close()
	if len(commitSHAList) > 0 {
		return commitSHAList[0]
	}else {
		return ""
	}
}

func sendMail(to []string, subject string, body string) error {
	host := "smtp.gmail.com"
	port := 587
	from := "themshd79@gmail.com"
	passwd := "102@MSpass"
	mesg := gomail.NewMessage()
	mesg.SetHeader("From", from)
	toHeaders := make(map[string][]string)
	toHeaders["To"] = to
	mesg.SetHeaders(toHeaders)
	mesg.SetHeader("Subject", subject)
	mesg.SetBody("text/html", body)
	dialer := gomail.NewDialer(host, port, from, passwd)
	if err := dialer.DialAndSend(mesg); err != nil {
		return err
	}
	return nil
}

func RunInstallation(config Configuration) {
	if config.AppConfig.RunInstallation == true {
		runList, runerr := RunAppCommands(config.AppCommands, config.AppConfig.ContinueOnFail)
		body := "<b style='color:green;'>Congratulations!</b><p>The automatic website deployment was successful.</p>"
		if runerr != nil {
			body = "<b style='color:red;'>We have a problem!</b><p>The automatic website deployment was not successful.</p>"
		}
		fmt.Println("Notifying admins via email...")
		runString := "The following commands were run:\n\n"
		for i, ac := range runList {
			if ac.runStatus == true {
				runString += "<p style='color:green;'><b>" + strconv.Itoa(i+1) + " " + ac.Name + ":</b></p>"
				runString += "<p style='color:blue; margin-left:5px;'>" + strings.Join(ac.Command, " ") + "</p>"
			} else {
				runString += "<p style='color:red;'><b>" + strconv.Itoa(i+1) + " " + ac.Name + " (failed!):</b></p>"
				runString += "<p style='color:red; margin-left:5px;'>" + strings.Join(ac.Command, " ") + "</p>"
			}
		}
		if config.AppConfig.ContinueOnFail == true {
			runString += "<p>Your app_config <span style='color:green;'>continue_on_fail</span> directive was enabled for this installation.</p>"
		}else {
			runString += "<p>Your app_config <span style='color:green;'>continue_on_fail</span> directive was disabled for this installation.</p>"
		}
		subject := "Automatic Deployment Status"
		body += runString
		to := config.AppConfig.NotifyEmails
		mailerr := sendMail(to, subject, body)
		if mailerr != nil {
			fmt.Printf("Error Notifying users of status: %s", mailerr)
		}
	}else {
		fmt.Println("Nothing was run because the installation directive is disabled")
	}

}


func (u User) String() string  {
	return fmt.Sprintf("User:<%v %v>", u.first_name, u.last_name )
}

func (u *User) GetName() string {
	// value receiver
	 name := ""
	 if u.first_name != "" {
	 	name += u.first_name
	 }
	 if u.last_name != "" {
	 	name += " " + u.last_name
	 }
	 return name
}

func (u *User) SetEmail(email string) {
	// use pointer receivers when changing struct values like here
	if email != "" {
		u.email = email
	}
}
func (u *User) SetLocation(latitude, longitue float32) {
	location := Location{
		latitude:  latitude,
		longitude: longitue,
	}
	u.location = location
}
type Location struct {
	latitude float32
	longitude float32
}

func (l Location) String() string {
	return fmt.Sprintf("Location:<%f %f>", l.latitude, l.longitude)
}

func main()  {
	repoPath, repoEnvExists := os.LookupEnv("REPO_DIRECTORY_PATH")
	repoName, repoNameEnvExists := os.LookupEnv("REPO_NAME")
	repoUser, repoUserEnvExists := os.LookupEnv("REPO_USER")
	githubToken, ghTokenEnvExists := os.LookupEnv("GITHUB_ACCESS_TOKEN")
	if !repoEnvExists {
		fmt.Println("Please add REPO_DIRECTORY_PATH Environment varibale")
		os.Exit(1)
	}
	if !repoNameEnvExists {
		fmt.Println("Please add REPO_NAME Environment variable")
		os.Exit(1)
	}
	if !repoUserEnvExists {
		fmt.Println("Please add REPO_USER Environment variable")
		os.Exit(1)
	}
	if !ghTokenEnvExists {
		fmt.Println("Please add GITHUB_ACCESS_TOKEN Environment variable")
		os.Exit(1)
	}
	chdirErr := os.Chdir(repoPath)
	if chdirErr != nil {
		fmt.Println("Cannot go to repo folder. Something went wrong!")
		os.Exit(1)
	}
	for {
		_, remoteSHA, status := FetchGitHubCommit(repoUser, repoName, githubToken)
		fmt.Println("Remote Commit SHA ", remoteSHA)
		if status {
			localSHA := getLocalCommit(repoPath)
			fmt.Println("Local Commit SHA ", localSHA)
			// go to repo directory and check commit there
			if localSHA == remoteSHA {
				// we don't need to pull changes and read config.json file for checking installation
				fmt.Println("Hashes are same. Nothing needs to be done")
			}else {
					// let's pull changes from github. Make sure you are inside the repo
					fmt.Println("Hashes are different. Let's pull changes")
					r, _ := git.PlainOpen(repoPath)
					w, _ := r.Worktree()
					err := w.Pull(&git.PullOptions{RemoteName: "origin"})
					if err != nil {
						fmt.Println("Error Pulling changes ", err.Error())
					}
					config, configPresent := ReadAppConfig()
					if !configPresent {
						fmt.Println("Is INSTALL_DIRECTIVES_PATH set to your json configuration file like some/path/config.json?")
						fmt.Println("You can set it like:")
						fmt.Println("export INSTALL_DIRECTIVES_PATH='/path/to/your/repo/config.json'")
						os.Exit(1)
					}
					RunInstallation(config)
			}
		} else {
			fmt.Println("Cannot fetch commit for this repo")
		}
		time.Sleep(time.Duration(time.Minute))
	}
}
