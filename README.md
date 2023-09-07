# bazel-azure

Bazel Azure credential helper that uses the Microsoft Authentication Library (MSAL) to authenticate Bazel users against a custom Azure application via OAuth2. The Azure application could be guarding remote execution requests against a remote execution backend.

```
$ ./bazel-azure --help
Usage of bazel-azure:
  -app string
        Azure application ID
  -authority string
        Azure authority endpoint
  -cachefile string
        Credentials cache file
  -config string
        Config file (default "/home/tweidner/bazel-azure/bazel-azure.conf")
  -login
        Run Azure device code login flow
  -nocredentialsmsg string
        User prompt when no credentials are found (default "No credentials found. Run `/home/tweidner/bazel-azure/bazel-azure --login`")
  -scope string
        Azure OAuth scopes delimited by commas
  -version
        Print version
```


Invoking the `credentialhelper` without any parameters prints the cached credentials to `stdout`.
The output is correctly formatted for `bazel` to consume via the `--credential_helper` flag.

```
$ ./bazel-azure
{
  "headers": {
    "Authorization": [
      "Bearer <OAUTH2TOKEN>"
    ]
  }
}
```

In order to obtain credentials for the configured Azure application run `bazel-azure` with the `--login` flag.

```
$ ./bazel-azure --login
To sign in, use a web browser to open the page https://microsoft.com/devicelogin and enter the code <AUTHCODE> to authenticate.
```

To use the `bazel-azure` credential helper with Bazel add the following parts

```
# .bazelrc
build --credential_helper=%workspace%/credentialhelper
```

```
# bazel-azure.conf
app <azure-app-id>
authority <microsoft-login-url>/<tenant-id>
scope api://<azure-app-id>/<custom-scope>
cachefile ~/.bazel-azure-credentials.json
nocredentialsmsg Please run `./bazel-azure --login`
```

```
cd $YOUR_BAZEL_ROOT

wget https://github.com/timaa2k/bazel-azure/releases/download/v0.1.0/bazel-azure

bazel build //...
```
