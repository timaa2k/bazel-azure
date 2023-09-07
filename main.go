package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/peterbourgon/ff/v3"
	"github.com/timaa2k/bazel-azure/cache"
)

const Version string = "0.1.0"

const (
	ExitSuccess             int = 0
	ExitFailure                 = 1
	ExitMisusedShellBuiltin     = 2
	ExitCredentialsNotFound     = 3
)

type Config struct {
	Login            bool
	AppID            string
	Authority        string
	Scope            string
	Cachefile        string
	NoCredentialsMsg string
}

type CredentialHelper struct {
	config        *Config
	cacheAccessor *cache.TokenCache
}

func New(c *Config) *CredentialHelper {
	return &CredentialHelper{
		config:        c,
		cacheAccessor: &cache.TokenCache{File: c.Cachefile},
	}
}

func (h *CredentialHelper) Login() (int, error) {
	app, err := public.New(h.config.AppID, public.WithCache(h.cacheAccessor), public.WithAuthority(h.config.Authority), public.WithInstanceDiscovery(false))
	if err != nil {
		return ExitFailure, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	dc, err := app.AcquireTokenByDeviceCode(ctx, strings.Split(h.config.Scope, ","))
	if err != nil {
		return ExitFailure, err
	}

	fmt.Fprintf(os.Stdout, "%s\n", dc.Result.Message)

	_, err = dc.AuthenticationResult(ctx)
	if err != nil {
		return ExitFailure, fmt.Errorf("got error while waiting for user to input the device code: %s", err)
	}
	return ExitSuccess, nil
}

func (h *CredentialHelper) PrintCredentials() (int, error) {
	app, err := public.New(h.config.AppID, public.WithCache(h.cacheAccessor), public.WithAuthority(h.config.Authority), public.WithInstanceDiscovery(false))
	if err != nil {
		return ExitFailure, err
	}

	accounts, err := app.Accounts(context.Background())
	if err != nil {
		return ExitCredentialsNotFound, fmt.Errorf("unable to read the cache")
	}

	for _, account := range accounts {
		ar, err := app.AcquireTokenSilent(context.TODO(), strings.Split(h.config.Scope, ","), public.WithSilentAccount(account))
		if err != nil {
			continue
		}

		type Headers struct {
			Authorization []string `json:"Authorization"`
		}

		type Credentials struct {
			Headers *Headers `json:"headers"`
		}

		credentials := &Credentials{
			Headers: &Headers{
				Authorization: []string{"Bearer " + ar.AccessToken},
			},
		}

		credentialsJson, err := json.MarshalIndent(credentials, "", "  ")
		if err != nil {
			return ExitFailure, err
		}

		fmt.Fprintf(os.Stdout, "%s\n", credentialsJson)
		return ExitSuccess, nil
	}

	return ExitCredentialsNotFound, fmt.Errorf(h.config.NoCredentialsMsg)
}

func (h *CredentialHelper) Run() (int, error) {
	if h.config.Login {
		return h.Login()
	} else {
		return h.PrintCredentials()
	}
}

func main() {
	fs := flag.NewFlagSet("bazel-azure", flag.ContinueOnError)

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitFailure)
	}
	exePath := filepath.Dir(exe)

	var (
		_                = fs.String("config", exePath+"/bazel-azure.conf", "Config file")
		version          = fs.Bool("version", false, "Print version")
		login            = fs.Bool("login", false, "Run Azure device code login flow")
		appID            = fs.String("app", "", "Azure application ID")
		authority        = fs.String("authority", "", "Azure authority endpoint")
		scope            = fs.String("scope", "", "Azure OAuth scopes delimited by commas")
		cachefile        = fs.String("cachefile", "", "Credentials cache file")
		nocredentialsmsg = fs.String("nocredentialsmsg", fmt.Sprintf("No credentials found. Run `%s --login`", exe), "User prompt when no credentials are found")
	)

	if err := ff.Parse(
		fs,
		os.Args[1:],
		ff.WithEnvVarPrefix("BAZEL_AZURE"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithAllowMissingConfigFile(true),
	); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitFailure)
	}

	if *version {
		fmt.Fprintf(os.Stdout, "v"+Version+"\n")
		os.Exit(ExitSuccess)
	}

	cachepath, err := resolveTilde(*cachefile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitFailure)
	}

	c := &Config{
		Login:            *login,
		AppID:            *appID,
		Authority:        *authority,
		Scope:            *scope,
		Cachefile:        cachepath,
		NoCredentialsMsg: *nocredentialsmsg,
	}

	h := New(c)

	if exitCode, err := h.Run(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(exitCode)
	}
}

func resolveTilde(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := usr.HomeDir
	if path == "~" {
		return dir, nil
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, path[2:]), nil
	}
	return path, nil
}
